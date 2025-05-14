package uploads

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestSaveUploadedFile(t *testing.T) {
	testDir := "test_uploads"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	fileContents := []byte("test file contents")

	testCases := []struct {
		name          string
		options       FileUploadOptions
		expectError   bool
		validateSaved func(*SavedFile) bool
	}{
		{
			name: "basic upload with default options",
			options: FileUploadOptions{
				DestinationDir: testDir,
				MaxSize:        10 * 1024 * 1024,
				AllowedExts:    []string{"jpg", "jpeg", "png", "gif", "txt"},
			},
			expectError: false,
			validateSaved: func(sf *SavedFile) bool {
				return strings.HasSuffix(sf.SavedName, ".txt") &&
					sf.Size == int64(len(fileContents)) &&
					sf.OriginalName == "test.txt"
			},
		},
		{
			name: "fixed filename",
			options: FileUploadOptions{
				DestinationDir:    testDir,
				RandomizeFilename: false,
				MaxSize:           10 * 1024 * 1024,
			},
			validateSaved: func(sf *SavedFile) bool {
				return sf.SavedName == "test.txt" &&
					sf.Size == int64(len(fileContents))
			},
		},
		{
			name: "randomized filename",
			options: FileUploadOptions{
				DestinationDir:    testDir,
				RandomizeFilename: true,
				MaxSize:           10 * 1024 * 1024,
			},
			validateSaved: func(sf *SavedFile) bool {
				return len(sf.SavedName) > len("test.txt") &&
					sf.Size == int64(len(fileContents)) &&
					sf.OriginalName == "test.txt"
			},
		},
		{
			name: "file size limit too small",
			options: FileUploadOptions{
				DestinationDir:    testDir,
				MaxSize:           5,
				RandomizeFilename: true,
			},
			expectError: true,
			validateSaved: func(sf *SavedFile) bool {
				return false
			},
		},

		{
			name: "with extension restriction - allowed",
			options: FileUploadOptions{
				DestinationDir:    testDir,
				AllowedExts:       []string{"txt", "doc"},
				RandomizeFilename: true,
				MaxSize:           10 * 1024 * 1024,
			},
			expectError: false,
			validateSaved: func(sf *SavedFile) bool {
				return strings.HasSuffix(sf.SavedName, ".txt")
			},
		},

		{
			name: "with extension restriction - disallowed",
			options: FileUploadOptions{
				DestinationDir:    testDir,
				AllowedExts:       []string{"jpg", "png"},
				RandomizeFilename: true,
				MaxSize:           10 * 1024 * 1024,
			},
			expectError: true,
			validateSaved: func(sf *SavedFile) bool {
				return false
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fh := createTestFileHeader("test.txt", fileContents)

			savedFile, err := SaveUploadedFile(fh, &tc.options)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error but got: %v", err)
					return
				}

				if !tc.validateSaved(savedFile) {
					t.Errorf("Saved file validation failed")
				}

				if _, err := os.Stat(savedFile.SavedPath); os.IsNotExist(err) {
					t.Errorf("Expected file to exist at %s, but it does not", savedFile.SavedPath)
				}

			}
		})
	}
}

func createTestFileHeader(filename string, content []byte) *multipart.FileHeader {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	part, _ := writer.CreateFormFile("file", filename)
	part.Write(content)
	writer.Close()

	req, _ := http.NewRequest("POST", "/", buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	req.ParseMultipartForm(32 << 20)
	file := req.MultipartForm.File["file"][0]

	return file
}
