package report

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWriter(t *testing.T) {
	tests := []struct {
		name        string
		baseDir     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid directory",
			baseDir: "./test-output",
			wantErr: false,
		},
		{
			name:        "path traversal",
			baseDir:     "../../etc/passwd",
			wantErr:     true,
			errContains: "path traversal detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := NewWriter(tt.baseDir)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, writer)
				assert.Equal(t, tt.baseDir, writer.baseDir)
			}
		})
	}
}

func TestWriteMarkdown(t *testing.T) {
	tempDir := t.TempDir()
	writer, err := NewWriter(tempDir)
	require.NoError(t, err)

	content := "# Test Markdown\n\nThis is a test."
	err = writer.WriteMarkdown("test.md", content)
	require.NoError(t, err)

	// Verify file was created
	filePath := filepath.Join(tempDir, "test.md")
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestWriteJSON(t *testing.T) {
	tempDir := t.TempDir()
	writer, err := NewWriter(tempDir)
	require.NoError(t, err)

	content := `{"key": "value"}`
	err = writer.WriteJSON("test.json", content)
	require.NoError(t, err)

	// Verify file was created
	filePath := filepath.Join(tempDir, "test.json")
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestWritePatch(t *testing.T) {
	tempDir := t.TempDir()
	writer, err := NewWriter(tempDir)
	require.NoError(t, err)

	content := "apiVersion: apps/v1\nkind: Deployment\n"
	err = writer.WritePatch("patches/prod", "test.yaml", content)
	require.NoError(t, err)

	// Verify file was created in subdirectory
	filePath := filepath.Join(tempDir, "patches/prod", "test.yaml")
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestGetFilePath(t *testing.T) {
	writer, err := NewWriter("./out")
	require.NoError(t, err)

	path := writer.GetFilePath("report.md")
	assert.Equal(t, filepath.Join("./out", "report.md"), path)
}

func TestGetPatchFilePath(t *testing.T) {
	writer, err := NewWriter("./out")
	require.NoError(t, err)

	path := writer.GetPatchFilePath("patches/prod", "deployment.yaml")
	assert.Equal(t, filepath.Join("./out", "patches/prod", "deployment.yaml"), path)
}
