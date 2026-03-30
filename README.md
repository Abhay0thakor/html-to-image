# HTML-to-Image Converter (Go)

A powerful, production-ready tool to convert HTML content (URLs, local files, or raw HTML) into images (PNG, JPEG) or PDF using headless Chromium.

## Features

- **Multi-source Input**: Supports direct URLs, local `.html` files, and raw HTML strings.
- **Bulk Processing**: Efficiently handles 100+ URLs with concurrent workers and browser pooling.
- **Selective Component Capture**: Use CSS selectors to capture specific elements (e.g., a div, section, or header).
- **Infinite Scroll Support**: Automatically scrolls to the bottom to trigger lazy-loaded content.
- **Custom Injection**: Inject custom CSS and JavaScript before capturing.
- **Smart Naming**: Naming strategies based on URL, Page Title, or custom templates with duplicate detection.
- **YAML Configuration**: Store complex settings (headers, cookies, viewports) in a configuration file.
- **Library & CLI**: Can be used as a standalone Go package or a CLI tool.
- **Dockerized**: Easy deployment with a bundled headless Chromium environment.

---

## CLI Usage

### Basic Conversion
```bash
./html-to-image -u https://example.com -o example.png
```

### Capture Specific Component
```bash
./html-to-image -u https://example.com -s ".main-content" -o content.png
```

### Bulk Processing from File
```bash
./html-to-image -b urls.txt -o results.zip --workers 10
```

### Using a Config File
```bash
./html-to-image --config config.yaml
```

---

## Configuration (`config.yaml`)

```yaml
urls:
  - "https://google.com"
  - "https://github.com"
format: "jpeg"
quality: 85
width: 1920
height: 1080
auto_scroll: true
wait_until: "networkIdle"
headers:
  User-Agent: "Mozilla/5.0 Custom"
  Authorization: "Bearer token"
cookies:
  - name: "session"
    value: "123"
    domain: "example.com"
custom_css: |
  nav { display: none !important; }
  .ads { display: none; }
```

---

## Library Usage (Go)

```go
import (
    "context"
    "github.com/user/html-to-image/pkg/converter"
    "github.com/user/html-to-image/pkg/models"
)

func main() {
    conv := converter.NewConverter()
    conv.Start(context.Background())
    defer conv.Shutdown()

    cfg := models.ConversionConfig{
        Input: "https://example.com",
        InputType: models.URL,
        OutputFormat: "png",
        Width: 1280,
        Height: 720,
    }

    imgData, err := conv.Convert(context.Background(), cfg)
    // ... handle err and imgData
}
```

---

## Docker Support

Build and run using Docker:

```bash
docker build -t html-to-image .
docker run --rm -v $(pwd):/app/out html-to-image -u https://example.com -o /app/out/result.png
```

---

## Installation

```bash
git clone https://github.com/user/html-to-image
cd html-to-image
go build -o html-to-image ./cmd/html-to-image/main.go
```

## License
MIT
