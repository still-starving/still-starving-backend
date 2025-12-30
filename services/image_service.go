package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/yourusername/food-sharing-backend/config"
)

type ImageService struct {
	minioClient *minio.Client
	cfg         *config.MinIOConfig
	maxSize     int64
}

func NewImageService(minioClient *minio.Client, cfg *config.MinIOConfig, maxSize int64) *ImageService {
	return &ImageService{
		minioClient: minioClient,
		cfg:         cfg,
		maxSize:     maxSize,
	}
}

func (s *ImageService) UploadImage(file *multipart.FileHeader) (string, error) {
	// Validate file size
	if file.Size > s.maxSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", s.maxSize)
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if !s.isValidImageType(contentType) {
		return "", fmt.Errorf("invalid file type: %s. Allowed types: image/jpeg, image/png, image/webp", contentType)
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Upload to MinIO
	ctx := context.Background()
	_, err = s.minioClient.PutObject(
		ctx,
		s.cfg.Bucket,
		filename,
		src,
		file.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	// Generate public URL
	imageURL := fmt.Sprintf("%s/%s/%s", s.cfg.PublicURL, s.cfg.Bucket, filename)
	return imageURL, nil
}

func (s *ImageService) isValidImageType(contentType string) bool {
	validTypes := []string{"image/jpeg", "image/png", "image/webp", "image/jpg"}
	for _, validType := range validTypes {
		if strings.EqualFold(contentType, validType) {
			return true
		}
	}
	return false
}

func (s *ImageService) DeleteImage(imageURL string) error {
	// Extract filename from URL
	parts := strings.Split(imageURL, "/")
	if len(parts) == 0 {
		return fmt.Errorf("invalid image URL")
	}
	filename := parts[len(parts)-1]

	// Delete from MinIO
	ctx := context.Background()
	err := s.minioClient.RemoveObject(ctx, s.cfg.Bucket, filename, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete from MinIO: %w", err)
	}

	return nil
}
