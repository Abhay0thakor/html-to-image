package utils

import (
	"archive/zip"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Abhay0thakor/html-to-image/pkg/models"
)

func CreateZip(outputPath string, results []models.ConversionResult) error {
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create zip file: %w", err)
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	for _, result := range results {
		if result.Error != nil {
			continue
		}

		f, err := archive.Create(result.Name)
		if err != nil {
			return fmt.Errorf("create file in zip: %w", err)
		}

		_, err = f.Write(result.ImageData)
		if err != nil {
			return fmt.Errorf("write image to zip: %w", err)
		}
	}

	return nil
}

type Namer struct {
	mu        sync.Mutex
	usedNames map[string]int
}

func NewNamer() *Namer {
	return &Namer{
		usedNames: make(map[string]int),
	}
}

func (n *Namer) GetUniqueName(name string, extension string) string {
	n.mu.Lock()
	defer n.mu.Unlock()

	name = sanitizeFilename(name)
	if extension != "" && !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	fullName := name + extension
	count, exists := n.usedNames[fullName]
	if !exists {
		n.usedNames[fullName] = 1
		return fullName
	}

	n.usedNames[fullName] = count + 1
	newName := fmt.Sprintf("%s (%d)%s", name, count, extension)
	return newName
}

func sanitizeFilename(name string) string {
	// Remove invalid characters for filenames
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "_")
	}
	// Limit length
	if len(name) > 200 {
		name = name[:200]
	}
	return name
}

func URLToFilename(url string) string {
	name := strings.TrimPrefix(url, "http://")
	name = strings.TrimPrefix(name, "https://")
	name = strings.ReplaceAll(name, "/", "_")
	return name
}
