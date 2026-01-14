package report

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/your-org/kost/internal/security"
)

// Writer handles writing report outputs to files
type Writer struct {
	baseDir string
}

// NewWriter creates a new report writer
func NewWriter(baseDir string) (*Writer, error) {
	// Validate output path for security
	if err := security.ValidateOutputPath(baseDir); err != nil {
		return nil, fmt.Errorf("invalid output directory: %w", err)
	}

	return &Writer{
		baseDir: baseDir,
	}, nil
}

// WriteMarkdown writes a markdown report to a file
func (w *Writer) WriteMarkdown(filename string, content string) error {
	path := filepath.Join(w.baseDir, filename)

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(w.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", w.baseDir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write markdown file %s: %w", path, err)
	}

	return nil
}

// WriteJSON writes a JSON report to a file
func (w *Writer) WriteJSON(filename string, content string) error {
	path := filepath.Join(w.baseDir, filename)

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(w.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", w.baseDir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write JSON file %s: %w", path, err)
	}

	return nil
}

// WritePatch writes a YAML patch to a file
func (w *Writer) WritePatch(subdir string, filename string, content string) error {
	dir := filepath.Join(w.baseDir, subdir)

	// Create patch directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create patch directory %s: %w", dir, err)
	}

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write patch file %s: %w", path, err)
	}

	return nil
}

// GetFilePath returns the full path for a given filename
func (w *Writer) GetFilePath(filename string) string {
	return filepath.Join(w.baseDir, filename)
}

// GetPatchFilePath returns the full path for a patch file
func (w *Writer) GetPatchFilePath(subdir string, filename string) string {
	return filepath.Join(w.baseDir, subdir, filename)
}
