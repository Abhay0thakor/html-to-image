package converter

import (
	"context"
	"testing"
	"github.com/Abhay0thakor/html-to-image/pkg/models"
)

func TestConvertRawHTML(t *testing.T) {
	c := NewConverter()
	if err := c.Start(context.Background()); err != nil {
		t.Fatalf("failed to start converter: %v", err)
	}
	defer c.Shutdown()

	cfg := models.ConversionConfig{
		Input:        "<h1>Hello World</h1>",
		InputType:    models.RawHTML,
		OutputFormat: "png",
		Width:        200,
		Height:       100,
		Scale:        1.0,
	}

	buf, err := c.Convert(context.Background(), cfg)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	if len(buf) == 0 {
		t.Fatal("empty image buffer returned")
	}
}
