package models

import "time"

type InputType int

const (
	URL InputType = iota
	File
	RawHTML
)

type ConversionConfig struct {
	Input        string
	InputType    InputType
	Selector     string // CSS selector for specific component
	OutputFormat string // "png" or "jpeg"
	Quality      int    // For JPEG
	Width        int
	Height       int
	Scale        float64
	OutputName   string
	NamingType   string // "url", "title", "custom"
	WaitUntil    string // "none", "networkIdle"
	CustomCSS    string
	CustomJS     string
	Headers      map[string]interface{}
	Cookies      []map[string]interface{}
	AutoScroll   bool
	WaitTime     int // Extra wait time in seconds
}

type ConversionResult struct {
	Name      string
	ImageData []byte
	Error     error
	Duration  time.Duration
}
