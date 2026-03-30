package utils

import (
	"testing"
)

func TestNamer(t *testing.T) {
	n := NewNamer()
	
	name1 := n.GetUniqueName("test", "png")
	if name1 != "test.png" {
		t.Errorf("expected test.png, got %s", name1)
	}
	
	name2 := n.GetUniqueName("test", "png")
	if name2 != "test (1).png" {
		t.Errorf("expected test (1).png, got %s", name2)
	}
	
	name3 := n.GetUniqueName("test", ".png") // with dot
	if name3 != "test (2).png" {
		t.Errorf("expected test (2).png, got %s", name3)
	}
}

func TestURLToFilename(t *testing.T) {
	url := "https://example.com/path/to/page"
	expected := "example.com_path_to_page"
	got := URLToFilename(url)
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}
