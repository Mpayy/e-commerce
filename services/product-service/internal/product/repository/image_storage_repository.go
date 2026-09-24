package repository

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/google/uuid"
)

//go:generate mockery

//mockery:generate: true
//mockery:filename: ../mocks/mock_image_storage_repository.go
type ImageStorage interface {
	Save(ctx context.Context, fileHeader *multipart.FileHeader) (path string, err error)
	Delete(ctx context.Context, path string) error
}

type LocalImageStorage struct {
	basePath        string
	publicURLPrefix string
}

func NewImageStorage(cfg *config.Config, log *logger.Logger) ImageStorage {
	basePath := cfg.BasePath
	if err := os.MkdirAll(basePath, 0755); err != nil {
		log.Fatalf("failed to create image storage directory: %v", err)
	}
	return &LocalImageStorage{
		basePath:        basePath,
		publicURLPrefix: cfg.PublicImageURLPrefix,
	}
}

func (s *LocalImageStorage) Save(ctx context.Context, fileHeader *multipart.FileHeader) (string, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	buf := make([]byte, 512)
	n, err := src.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	contentType := http.DetectContentType(buf[:n])

	allowedExtensions := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}

	ext, ok := allowedExtensions[contentType]
	if !ok {
		return "", apperror.ErrInvalidFileType
	}

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	filename := uuid.NewString() + ext
	dst, err := os.Create(filepath.Join(s.basePath, filename))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return path.Join(s.publicURLPrefix, filename), nil
}

func (s *LocalImageStorage) Delete(ctx context.Context, path string) error {
	if path == "" {
		return nil
	}

	fileName := filepath.Base(path)
	fullPath := filepath.Join(s.basePath, fileName)

	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
