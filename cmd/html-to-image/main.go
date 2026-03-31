package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/html-to-image/pkg/converter"
	"github.com/user/html-to-image/pkg/models"
	"github.com/user/html-to-image/pkg/utils"
	"gopkg.in/yaml.v3"
)

var (
	urls         []string
	filePaths    []string
	rawHTML      string
	selector     string
	format       string
	quality      int
	width        int
	height       int
	scale        float64
	output       string
	naming       string
	workers      int
	bulkFilePath string
	waitUntil    string
	customCSS    string
	customJS     string
	configPath   string
	autoScroll   bool
	waitTime     int
	headers      map[string]interface{}
	cookies      []map[string]interface{}
)

type Config struct {
	URLs         []string                 `yaml:"urls"`
	FilePaths    []string                 `yaml:"file_paths"`
	Selector     string                   `yaml:"selector"`
	Format       string                   `yaml:"format"`
	Quality      int                      `yaml:"quality"`
	Width        int                      `yaml:"width"`
	Height       int                      `yaml:"height"`
	Scale        float64                  `yaml:"scale"`
	Output       string                   `yaml:"output"`
	Naming       string                   `yaml:"naming"`
	Workers      int                      `yaml:"workers"`
	WaitUntil    string                   `yaml:"wait_until"`
	CustomCSS    string                   `yaml:"custom_css"`
	CustomJS     string                   `yaml:"custom_js"`
	AutoScroll   bool                     `yaml:"auto_scroll"`
	WaitTime     int                      `yaml:"wait_time"`
	Headers      map[string]interface{}   `yaml:"headers"`
	Cookies      []map[string]interface{} `yaml:"cookies"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "html-to-image",
		Short: "A tool to convert HTML content into images",
		Run:   run,
	}

	rootCmd.Flags().StringSliceVarP(&urls, "url", "u", []string{}, "Direct URLs to convert")
	rootCmd.Flags().StringSliceVarP(&filePaths, "file", "f", []string{}, "Local HTML files to convert")
	rootCmd.Flags().StringVarP(&rawHTML, "raw", "r", "", "Raw HTML code to convert")
	rootCmd.Flags().StringVarP(&selector, "selector", "s", "", "CSS selector to capture specific component")
	rootCmd.Flags().StringVar(&format, "format", "png", "Output format (png, jpeg, pdf)")
	rootCmd.Flags().IntVar(&quality, "quality", 100, "JPEG quality (0-100)")
	rootCmd.Flags().IntVar(&width, "width", 1280, "Viewport width")
	rootCmd.Flags().IntVar(&height, "height", 720, "Viewport height")
	rootCmd.Flags().Float64Var(&scale, "scale", 1.0, "Resolution scale")
	rootCmd.Flags().StringVarP(&output, "output", "o", "output.zip", "Output filename (for multiple files it's a ZIP)")
	rootCmd.Flags().StringVar(&naming, "naming", "url", "Naming strategy (url, title, custom)")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 5, "Number of concurrent workers")
	rootCmd.Flags().StringVarP(&bulkFilePath, "bulk-file", "b", "", "Path to a text file containing URLs (one per line)")
	rootCmd.Flags().StringVar(&waitUntil, "wait-until", "none", "Wait strategy (none, networkIdle)")
	rootCmd.Flags().StringVar(&customCSS, "inject-css", "", "CSS code to inject before capture")
	rootCmd.Flags().StringVar(&customJS, "inject-js", "", "JavaScript code to execute before capture")
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to a YAML configuration file")
	rootCmd.Flags().BoolVar(&autoScroll, "auto-scroll", false, "Scroll to bottom of page before capture")
	rootCmd.Flags().IntVar(&waitTime, "wait-time", 0, "Extra wait time in seconds")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	// Load config if provided
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatalf("failed to read config file: %v", err)
		}
		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			log.Fatalf("failed to parse config file: %v", err)
		}

		// Apply config to variables if flags are NOT changed from defaults
		if !cmd.Flags().Changed("url") && len(cfg.URLs) > 0 {
			urls = cfg.URLs
		}
		if !cmd.Flags().Changed("file") && len(cfg.FilePaths) > 0 {
			filePaths = cfg.FilePaths
		}
		if !cmd.Flags().Changed("selector") && cfg.Selector != "" {
			selector = cfg.Selector
		}
		if !cmd.Flags().Changed("format") && cfg.Format != "" {
			format = cfg.Format
		}
		if !cmd.Flags().Changed("quality") && cfg.Quality != 0 {
			quality = cfg.Quality
		}
		if !cmd.Flags().Changed("width") && cfg.Width != 0 {
			width = cfg.Width
		}
		if !cmd.Flags().Changed("height") && cfg.Height != 0 {
			height = cfg.Height
		}
		if !cmd.Flags().Changed("scale") && cfg.Scale != 0 {
			scale = cfg.Scale
		}
		if !cmd.Flags().Changed("output") && cfg.Output != "" {
			output = cfg.Output
		}
		if !cmd.Flags().Changed("naming") && cfg.Naming != "" {
			naming = cfg.Naming
		}
		if !cmd.Flags().Changed("workers") && cfg.Workers != 0 {
			workers = cfg.Workers
		}
		if !cmd.Flags().Changed("wait-until") && cfg.WaitUntil != "" {
			waitUntil = cfg.WaitUntil
		}
		if !cmd.Flags().Changed("inject-css") && cfg.CustomCSS != "" {
			customCSS = cfg.CustomCSS
		}
		if !cmd.Flags().Changed("inject-js") && cfg.CustomJS != "" {
			customJS = cfg.CustomJS
		}
		if !cmd.Flags().Changed("auto-scroll") && cfg.AutoScroll {
			autoScroll = cfg.AutoScroll
		}
		if !cmd.Flags().Changed("wait-time") && cfg.WaitTime != 0 {
			waitTime = cfg.WaitTime
		}
		headers = cfg.Headers
		cookies = cfg.Cookies
	}

	ctx := context.Background()
	conv := converter.NewConverter()
	if err := conv.Start(ctx); err != nil {
		log.Fatalf("failed to start browser: %v", err)
	}
	defer conv.Shutdown()

	p := converter.NewProcessor(conv)

	var configs []models.ConversionConfig

	// 1. Process URLs
	for _, url := range urls {
		configs = append(configs, createConfig(url, models.URL))
	}

	// 2. Process Files
	for _, path := range filePaths {
		configs = append(configs, createConfig(path, models.File))
	}

	// 3. Process Raw HTML
	if rawHTML != "" {
		configs = append(configs, createConfig(rawHTML, models.RawHTML))
	}

	// 4. Process Bulk URL File
	if bulkFilePath != "" {
		file, err := os.Open(bulkFilePath)
		if err != nil {
			log.Fatalf("failed to open bulk file: %v", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			url := strings.TrimSpace(scanner.Text())
			if url != "" {
				configs = append(configs, createConfig(url, models.URL))
			}
		}
	}

	if len(configs) == 0 {
		fmt.Println("No input provided. Use -u, -f, -r, or -b flags.")
		return
	}

	fmt.Printf("Starting conversion for %d items using %d workers...\n", len(configs), workers)
	results, err := p.ProcessBulk(ctx, configs, workers)
	if err != nil {
		log.Fatalf("bulk processing failed: %v", err)
	}

	// Check if only one result and not ZIP output was requested (simple case)
	if len(results) == 1 && !strings.HasSuffix(output, ".zip") {
		result := results[0]
		if result.Error != nil {
			log.Fatalf("conversion failed: %v", result.Error)
		}
		err = os.WriteFile(output, result.ImageData, 0644)
		if err != nil {
			log.Fatalf("failed to write output: %v", err)
		}
		fmt.Printf("Saved to %s\n", output)
	} else {
		// Default to ZIP for multiple or if ZIP explicitly requested
		if !strings.HasSuffix(output, ".zip") {
			output += ".zip"
		}
		err = utils.CreateZip(output, results)
		if err != nil {
			log.Fatalf("failed to create zip: %v", err)
		}
		fmt.Printf("Bulk results saved to %s\n", output)
	}
}

func createConfig(input string, itype models.InputType) models.ConversionConfig {
	name := "result"
	switch naming {
	case "url":
		if itype == models.URL {
			name = utils.URLToFilename(input)
		}
	case "title":
		name = "loading_title" // Will be updated in processor
	case "custom":
		name = "image" // Fallback or can be enhanced
	}

	return models.ConversionConfig{
		Input:        input,
		InputType:    itype,
		Selector:     selector,
		OutputFormat: format,
		Quality:      quality,
		Width:        width,
		Height:       height,
		Scale:        scale,
		OutputName:   name,
		NamingType:   naming,
		WaitUntil:    waitUntil,
		CustomCSS:    customCSS,
		CustomJS:     customJS,
		Headers:      headers,
		Cookies:      cookies,
		AutoScroll:   autoScroll,
		WaitTime:     waitTime,
	}
}
