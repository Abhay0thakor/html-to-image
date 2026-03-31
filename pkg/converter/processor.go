package converter

import (
	"context"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/Abhay0thakor/html-to-image/pkg/models"
	"github.com/Abhay0thakor/html-to-image/pkg/utils"
)

type Processor struct {
	converter *Converter
	namer     *utils.Namer
}

func NewProcessor(c *Converter) *Processor {
	return &Processor{
		converter: c,
		namer:     utils.NewNamer(),
	}
}

func (p *Processor) ProcessBulk(ctx context.Context, configs []models.ConversionConfig, workers int) ([]models.ConversionResult, error) {
	results := make([]models.ConversionResult, len(configs))
	bar := progressbar.Default(int64(len(configs)))

	var wg sync.WaitGroup
	jobs := make(chan int, len(configs))

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				i := i // Capture i
				start := time.Now()
				cfg := configs[i] // Get config for this job

				// Get name first if based on title
				if cfg.NamingType == "title" {
					title, err := p.converter.GetPageTitle(ctx, cfg.Input, cfg.InputType)
					if err == nil && title != "" {
						cfg.OutputName = title
					}
				}

				name := p.namer.GetUniqueName(cfg.OutputName, cfg.OutputFormat)
				cfg.OutputName = name

				data, err := p.converter.Convert(ctx, cfg)
				results[i] = models.ConversionResult{
					Name:      name,
					ImageData: data,
					Error:     err,
					Duration:  time.Since(start),
				}
				bar.Add(1)
			}
		}()
	}

	for i := range configs {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	return results, nil
}
