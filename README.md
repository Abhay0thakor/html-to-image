# HTML-to-Image Converter (Go)

A high-performance, production-ready tool to convert HTML content into images or PDF. Designed for bulk processing and precision capture of web components.

## Featured Showcase: Toolsura

The tool is optimized for capturing complex web tools like those found on [Toolsura.com](https://www.toolsura.com/).

### Examples
Results generated with this tool:
- **[Full Homepage](docs/examples/toolsura/homepage_full.png)** (using `--auto-scroll`)
- **[PDF Export](docs/examples/toolsura/homepage.pdf)** (using `--format pdf`)
- **[Header Only](docs/examples/toolsura/toolsura_header.png)** (using `-s "header"`)

---

## Detailed Features

### 1. Wait Strategies & Reliability
To ensure pages with heavy JavaScript (like Toolsura) load completely, we use:
- **`document.readyState` verification**: Ensures the DOM is fully constructed.
- **`--wait-until networkIdle`**: Provides an extra grace period for AJAX and image lazy-loading.
- **Increased Timeouts**: Handles slow-loading assets up to 90 seconds per page.

### 2. Infinite Scroll Handling
The `--auto-scroll` flag uses a custom JavaScript driver to scroll through the page at intervals, triggering "lazy-load" events for images and dynamic content.

### 3. Smart Component Capture
Instead of simple cropping, the `-s` flag uses Chrome's native element screenshotting. This ensures that the selected component (e.g., `#tool-workspace`) is captured with its full internal scroll height and styling.

### 4. Custom Injection
Modify the page before capture:
- `--inject-css`: Hide ads, banners, or navigation.
- `--inject-js`: Trigger specific UI states (e.g., open a modal).

---

## Technical Architecture

### Browser Context Pooling
Unlike standard tools that launch a new browser for every image, this tool maintains a single **Browser Context Pool**. 
- One long-lived Chromium process.
- Multiple concurrent "Tabs" (Contexts).
- Significant reduction in CPU/RAM overhead during bulk tasks.

### Concurrency Model
Built on Go's `errgroup` and `channels`, it supports high-concurrency workers (`-w`) to process hundreds of URLs in parallel without crashing the system.

---

## Installation & Deployment

### Local Build
```bash
go build -o html-to-image ./cmd/html-to-image/main.go
```

### Docker (Recommended)
Our Docker image includes a lightweight Alpine-based Chromium environment.
```bash
docker build -t html-to-image .
```

## Complete Flag Reference

| Flag | Shorthand | Default | Description |
| :--- | :--- | :--- | :--- |
| `--url` | `-u` | | Direct URL(s) to convert |
| `--file` | `-f` | | Local HTML file path(s) |
| `--raw` | `-r` | | Raw HTML string |
| `--selector`| `-s` | | CSS selector for specific component |
| `--format` | | `png` | `png`, `jpeg`, or `pdf` |
| `--quality` | | `100` | JPEG quality (0-100) |
| `--width` | | `1280` | Viewport width |
| `--height` | | `720` | Viewport height |
| `--scale` | | `1.0` | High-DPI scale factor |
| `--output` | `-o` | `output.zip` | Output filename |
| `--naming` | | `url` | `url`, `title`, or `custom` |
| `--workers` | `-w` | `5` | Concurrent browser tabs |
| `--config` | `-c` | | Path to YAML config file |
| `--auto-scroll`| | `false` | Enable infinite scroll driver |
| `--wait-until` | | `none` | Use `networkIdle` for heavy sites |
| `--inject-css` | | | Custom CSS injection |
| `--inject-js` | | | Custom JS injection |

## License
MIT

