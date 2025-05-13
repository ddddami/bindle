package uploads

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ddddami/bindle/random"
)

var (
	ErrEmptyFile       = errors.New("uploaded file is empty")
	ErrFileTooLarge    = errors.New("uploaded file is too large")
	ErrInvalidFileType = errors.New("file extension is not allowed")
)

type FileUploadOptions struct {
	DestinationDir string
	MaxSize        int64
	// AllowedExts is a list of allowed file extensions (without the dot). Leave empty to allow all.
	AllowedExts       []string
	AllowedMimeTypes  []string
	FilenamePrefix    string
	RandomizeFilename bool
	// TODO: (custom validator) Validator func(file *multipart.FileHeader) error
}

type SavedFile struct {
	OriginalName string
	SavedName    string
	SavedPath    string
	Size         int64
	MIMEType     string
}

func DefaultOptions() FileUploadOptions {
	return FileUploadOptions{
		DestinationDir:   "uploads",
		MaxSize:          10 * 1024 * 1024, // 10 MB
		AllowedExts:      []string{"jpg", "jpeg", "png", "gif", "pdf"},
		AllowedMimeTypes: []string{"image/jpeg", "image/png", "image/gif", "application/pdf"},
	}
}

func SaveUploadedFile(file *multipart.FileHeader, opts *FileUploadOptions) (*SavedFile, error) {
	if file.Size == 0 {
		return nil, ErrEmptyFile
	}

	if opts == nil {
		defaultOpts := DefaultOptions()
		opts = &defaultOpts
	}

	if file.Size > opts.MaxSize {
		return nil, ErrFileTooLarge
	}

	origFileName := file.Filename
	ext := strings.ToLower(filepath.Ext(origFileName))
	if ext != "" {
		ext = ext[1:]
	}
	if len(opts.AllowedExts) > 0 {
		allowed := false
		for _, allowedExt := range opts.AllowedExts {
			if strings.EqualFold(ext, allowedExt) {
				allowed = true
				break
			}
		}

		if !allowed {
			return nil, ErrInvalidFileType
		}
	}

	if err := os.MkdirAll(opts.DestinationDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	var destFileName string

	baseFileName := strings.TrimSuffix(origFileName, filepath.Ext(origFileName))
	if opts.FilenamePrefix != "" {
		baseFileName = fmt.Sprintf("%s_%s", opts.FilenamePrefix, baseFileName)
	}

	if opts.RandomizeFilename {
		random, err := random.Generate(random.Options{Length: 8})
		if err != nil {
			return nil, fmt.Errorf("failed to generate random string: %w", err)
		}

		destFileName = fmt.Sprintf("%s_%s.%s", baseFileName, random, ext)

	} else {
		destFileName = origFileName
	}

	destFileName = santizeFileName(destFileName)

	destPath := filepath.Join(opts.DestinationDir, destFileName)

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}

	defer src.Close()

	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}

	buffer = buffer[:n]

	mimeType := http.DetectContentType(buffer)

	if _, err = src.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset file reader: %w", err)
	}

	if opts.AllowedMimeTypes != nil {
		allowed := false
		for _, allowedMime := range opts.AllowedMimeTypes {
			if strings.HasPrefix(mimeType, allowedMime) {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("unsupported MIME type: %s", mimeType)
		}
	} else {
		if !strings.HasPrefix(mimeType, "image/") &&
			mimeType != "application/pdf" &&
			!strings.HasPrefix(mimeType, "text/") {
			return nil, fmt.Errorf("unsupported MIME type: %s", mimeType)
		}
	}

	dst, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to created destination file: %w", err)
	}

	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	savedFile := &SavedFile{
		OriginalName: origFileName,
		SavedName:    destFileName,
		SavedPath:    destPath,
		Size:         file.Size,
		MIMEType:     file.Header.Get("Content-Type"),
	}
	return savedFile, nil
}

func santizeFileName(fileName string) string {
	replacer := strings.NewReplacer(
		" ", "_",
		"\\", "",
		"/", "",
		":", "",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "",
	)
	return replacer.Replace(fileName)
}

func SaveSingleFormFile(r *http.Request, fieldName string, opts *FileUploadOptions) (*SavedFile, error) {
	if r.MultipartForm == nil {
		if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
			return nil, fmt.Errorf("failed to parse multipart form: %w", err)
		}
	}

	files := r.MultipartForm.File[fieldName]
	if len(files) == 0 {
		return nil, fmt.Errorf("no file uploaded with field name '%s'", fieldName)
	}

	return SaveUploadedFile(files[0], opts)
}

func SaveMultipleFormFiles(r *http.Request, fieldName string, opts *FileUploadOptions) ([]*SavedFile, error) {
	if r.MultipartForm == nil {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return nil, fmt.Errorf("failed to parse multipart from %w", err)
		}
	}

	files := r.MultipartForm.File[fieldName]
	if len(files) == 0 {
		return nil, fmt.Errorf("no files uploaded with the given field name '%s'", fieldName)
	}

	var savedFiles []*SavedFile
	for _, fileHeader := range files {
		savedFile, err := SaveUploadedFile(fileHeader, opts)
		if err != nil {
			for _, saved := range savedFiles {
				os.Remove(saved.SavedPath)
			}
		}

		savedFiles = append(savedFiles, savedFile)

	}
	return savedFiles, nil
}
