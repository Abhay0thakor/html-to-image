package converter

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/user/html-to-image/pkg/models"
)

type Converter struct {
	allocCtx    context.Context
	allocCancel context.CancelFunc
}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) Start(ctx context.Context) error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.DisableGPU,
	)

	c.allocCtx, c.allocCancel = chromedp.NewExecAllocator(ctx, opts...)

	// Start browser
	tmpCtx, cancel := chromedp.NewContext(c.allocCtx)
	defer cancel()
	return chromedp.Run(tmpCtx)
}

func (c *Converter) Shutdown() {
	if c.allocCancel != nil {
		c.allocCancel()
	}
}

func (c *Converter) Convert(ctx context.Context, cfg models.ConversionConfig) ([]byte, error) {
	if c.allocCtx == nil {
		return nil, fmt.Errorf("converter not started")
	}

	tabCtx, cancel := chromedp.NewContext(c.allocCtx)
	defer cancel()

	tabCtx, cancel = context.WithTimeout(tabCtx, 120*time.Second)
	defer cancel()

	var buf []byte
	var actions []chromedp.Action

	// 0. Set Headers and Cookies
	if len(cfg.Headers) > 0 {
		actions = append(actions, network.SetExtraHTTPHeaders(network.Headers(cfg.Headers)))
	}

	for _, cookie := range cfg.Cookies {
		name, _ := cookie["name"].(string)
		value, _ := cookie["value"].(string)
		domain, _ := cookie["domain"].(string)
		path, _ := cookie["path"].(string)
		if name != "" && value != "" {
			actions = append(actions, network.SetCookie(name, value).WithDomain(domain).WithPath(path))
		}
	}

	// 1. Initial Viewport
	actions = append(actions, emulation.SetDeviceMetricsOverride(int64(cfg.Width), int64(cfg.Height), cfg.Scale, false))

	// 2. Load Content
	switch cfg.InputType {
	case models.URL:
		actions = append(actions, chromedp.Navigate(cfg.Input))
	case models.File:
		actions = append(actions, chromedp.Navigate("file://"+cfg.Input))
	case models.RawHTML:
		actions = append(actions, chromedp.Navigate("data:text/html;charset=utf-8,"+cfg.Input))
	}

	// 3. Wait for load
	actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
		var state string
		for i := 0; i < 40; i++ {
			if err := chromedp.Evaluate(`document.readyState`, &state).Do(ctx); err != nil {
				return err
			}
			if state == "complete" {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		return nil
	}))

	// 3.1 Extra Wait
	waitTime := time.Duration(cfg.WaitTime) * time.Second
	if waitTime == 0 {
		waitTime = 2 * time.Second
	}
	actions = append(actions, chromedp.Sleep(waitTime))

	// 3.5 Inject CSS/JS
	if cfg.CustomCSS != "" {
		actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
			script := fmt.Sprintf(`
				(function() {
					var style = document.createElement('style');
					style.type = 'text/css';
					style.innerHTML = %q;
					document.head.appendChild(style);
				})()
			`, cfg.CustomCSS)
			return chromedp.Evaluate(script, nil).Do(ctx)
		}))
	}

	if cfg.CustomJS != "" {
		actions = append(actions, chromedp.Evaluate(cfg.CustomJS, nil))
	}

	// 3.6 Auto Scroll
	if cfg.AutoScroll {
		actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
			script := `
				(async () => {
					await new Promise((resolve) => {
						var totalHeight = 0;
						var distance = 200;
						var timer = setInterval(() => {
							var scrollHeight = document.body.scrollHeight;
							window.scrollBy(0, distance);
							totalHeight += distance;
							if(totalHeight >= scrollHeight){
								clearInterval(timer);
								resolve();
							}
						}, 150);
					});
				})()
			`
			return chromedp.Evaluate(script, nil).Do(ctx)
		}))
		actions = append(actions, chromedp.Sleep(2*time.Second))
	}

	// 4. Capture
	if cfg.Selector != "" {
		actions = append(actions, chromedp.WaitVisible(cfg.Selector, chromedp.ByQuery))
		
		actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
			// Deep expansion and measurement script
			script := fmt.Sprintf(`
				(function() {
					var el = document.querySelector(%q);
					if (!el) return null;
					
					// 1. Force expand target
					el.style.height = 'auto';
					el.style.minHeight = '0px';
					el.style.maxHeight = 'none';
					el.style.overflow = 'visible';
					el.style.display = 'block';

					// 2. Expand all parents to ensure no clipping
					var parent = el.parentElement;
					while (parent && parent !== document.body) {
						parent.style.overflow = 'visible';
						parent.style.height = 'auto';
						parent.style.maxHeight = 'none';
						parent = parent.parentElement;
					}

					// 3. Measure including scroll dimensions
					var rect = el.getBoundingClientRect();
					var width = Math.ceil(Math.max(rect.width, el.scrollWidth));
					var height = Math.ceil(Math.max(rect.height, el.scrollHeight));
					
					return {
						width: width,
						height: height,
						x: Math.floor(rect.left + window.pageXOffset),
						y: Math.floor(rect.top + window.pageYOffset)
					};
				})()
			`, cfg.Selector)
			
			var bounds struct {
				Width  float64 `json:"width"`
				Height float64 `json:"height"`
				X      float64 `json:"x"`
				Y      float64 `json:"y"`
			}
			
			if err := chromedp.Evaluate(script, &bounds).Do(ctx); err != nil {
				return err
			}

			if bounds.Width <= 0 || bounds.Height <= 0 {
				return fmt.Errorf("invalid element dimensions: %+v", bounds)
			}

			// Ensure viewport is large enough to contain the element at its offset
			vWidth := int64(bounds.X + bounds.Width + 50)
			vHeight := int64(bounds.Y + bounds.Height + 50)
			
			// Update viewport to fit the whole expanded element
			if err := emulation.SetDeviceMetricsOverride(vWidth, vHeight, cfg.Scale, false).Do(ctx); err != nil {
				return err
			}
			
			// Allow layout to reflow after viewport expansion
			time.Sleep(2 * time.Second)

			// Capture with precise clip
			var err error
			buf, err = page.CaptureScreenshot().
				WithClip(&page.Viewport{
					X:      bounds.X,
					Y:      bounds.Y,
					Width:  bounds.Width,
					Height: bounds.Height,
					Scale:  1.0, // Scale is already applied via SetDeviceMetricsOverride
				}).
				WithFormat(func() page.CaptureScreenshotFormat {
					if cfg.OutputFormat == "jpeg" { return page.CaptureScreenshotFormatJpeg }
					return page.CaptureScreenshotFormatPng
				}()).
				WithQuality(int64(cfg.Quality)).
				Do(ctx)
			return err
		}))
	} else {
		// Full page capture logic
		if cfg.OutputFormat == "jpeg" {
			actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				buf, err = page.CaptureScreenshot().WithFormat(page.CaptureScreenshotFormatJpeg).WithQuality(int64(cfg.Quality)).Do(ctx)
				return err
			}))
		} else if cfg.OutputFormat == "pdf" {
			actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				buf, _, err = page.PrintToPDF().Do(ctx)
				return err
			}))
		} else {
			actions = append(actions, chromedp.FullScreenshot(&buf, 100))
		}
	}

	if err := chromedp.Run(tabCtx, actions...); err != nil {
		return nil, fmt.Errorf("chromedp run: %w", err)
	}

	return buf, nil
}

func (c *Converter) GetPageTitle(ctx context.Context, input string, inputType models.InputType) (string, error) {
	if c.allocCtx == nil { return "", fmt.Errorf("converter not started") }
	tabCtx, cancel := chromedp.NewContext(c.allocCtx)
	defer cancel()
	var title string
	var action chromedp.Action
	switch inputType {
	case models.URL: action = chromedp.Navigate(input)
	case models.RawHTML: action = chromedp.Navigate("data:text/html;charset=utf-8,"+input)
	default: return "untitled", nil
	}
	if err := chromedp.Run(tabCtx, action, chromedp.Title(&title)); err != nil { return "", err }
	return title, nil
}
