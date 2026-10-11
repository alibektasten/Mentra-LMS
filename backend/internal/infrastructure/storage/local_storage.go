package storage

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const MaxUploadBytes int64 = 10 << 20 // 10 MiB

var ErrInvalidUpload = errors.New("invalid upload")

// allowedExtensions deliberately excludes formats that a browser can execute
// directly (HTML, SVG, JavaScript). Uploaded files are served from the app's
// public /uploads path, so accepting those formats would create an XSS risk.
var allowedExtensions = map[string]bool{
	".pdf": true, ".png": true, ".jpg": true, ".jpeg": true, ".webp": true,
	".txt": true, ".csv": true, ".docx": true, ".xlsx": true, ".pptx": true,
}

type LocalStorage struct {
	BaseDir string
}

func NewLocalStorage(baseDir string) *LocalStorage {
	if baseDir == "" {
		baseDir = "./uploads"
	}
	return &LocalStorage{BaseDir: baseDir}
}

func (s *LocalStorage) Save(file multipart.File, header *multipart.FileHeader, orgID string) (string, error) {
	if header == nil || header.Filename == "" {
		return "", fmt.Errorf("%w: file is required", ErrInvalidUpload)
	}
	if header.Size <= 0 {
		return "", fmt.Errorf("%w: file is empty", ErrInvalidUpload)
	}
	if header.Size > MaxUploadBytes {
		return "", fmt.Errorf("%w: file exceeds the 10 MB limit", ErrInvalidUpload)
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExtensions[ext] {
		return "", fmt.Errorf("%w: unsupported file type", ErrInvalidUpload)
	}

	// Do not trust the multipart Content-Type supplied by a client. Sniff the
	// actual content and restore the stream before persisting it.
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("%w: could not inspect file", ErrInvalidUpload)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("%w: could not read file", ErrInvalidUpload)
	}
	detectedType := http.DetectContentType(buf[:n])
	if !isAllowedContentType(ext, detectedType) {
		return "", fmt.Errorf("%w: file content does not match an allowed type", ErrInvalidUpload)
	}

	orgDir := filepath.Join(s.BaseDir, orgID)
	if err := os.MkdirAll(orgDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	storedName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().UnixMilli(), ext)
	storedPath := filepath.Join(orgDir, storedName)

	dst, err := os.Create(storedPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Return relative path for serving
	return filepath.Join("uploads", orgID, storedName), nil
}

func isAllowedContentType(ext, contentType string) bool {
	switch ext {
	case ".pdf":
		return contentType == "application/pdf"
	case ".png":
		return contentType == "image/png"
	case ".jpg", ".jpeg":
		return contentType == "image/jpeg"
	case ".webp":
		return contentType == "image/webp"
	case ".txt", ".csv":
		return strings.HasPrefix(contentType, "text/") || contentType == "application/octet-stream"
	case ".docx", ".xlsx", ".pptx":
		// Office Open XML documents are ZIP containers. Their precise browser
		// upload MIME varies, but executable formats are never in this allowlist.
		return contentType == "application/zip" || contentType == "application/octet-stream"
	default:
		return false
	}
}
