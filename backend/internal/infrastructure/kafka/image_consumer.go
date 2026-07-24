package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/monitoring-system/backend/internal/domain/upload"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/uploadrepo"
	"github.com/monitoring-system/backend/internal/usecase/uploadusecase"
	pkgkafka "github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/storage"
	"github.com/segmentio/kafka-go"
)

// ImageProcessingConsumer handles async image processing via Kafka
type ImageProcessingConsumer struct {
	consumer     *pkgkafka.EventConsumer
	minioStorage *storage.MinioStorage
	imageProc    *uploadusecase.ImageProcessor
	uploadRepo   *uploadrepo.UploadRepository
	log          *logger.Logger
}

// NewImageProcessingConsumer creates a new image processing consumer
func NewImageProcessingConsumer(
	brokers []string,
	groupID string,
	topic string,
	minioStorage *storage.MinioStorage,
	imageProc *uploadusecase.ImageProcessor,
	uploadRepo *uploadrepo.UploadRepository,
	log *logger.Logger,
) *ImageProcessingConsumer {
	return &ImageProcessingConsumer{
		consumer:     pkgkafka.NewConsumer(brokers, groupID, topic, log),
		minioStorage: minioStorage,
		imageProc:    imageProc,
		uploadRepo:   uploadRepo,
		log:          log,
	}
}

// Start begins consuming messages from Kafka
func (c *ImageProcessingConsumer) Start(ctx context.Context) {
	c.log.Info("Starting image processing consumer")
	c.consumer.Start(ctx, c.handleMessage)
}

// handleMessage processes a single image processing message
func (c *ImageProcessingConsumer) handleMessage(ctx context.Context, msg kafka.Message) error {
	var task upload.ImageProcessingMessage
	if err := json.Unmarshal(msg.Value, &task); err != nil {
		c.log.Error("Failed to unmarshal image processing message", logger.Error(err))
		return err
	}

	c.log.Info("Processing image",
		logger.String("file_id", task.FileID),
		logger.String("object_name", task.ObjectName),
	)

	// 1. Download original image from source URL
	originalData, err := c.downloadFromURL(task.SourceURL)
	if err != nil {
		c.log.Error("Failed to download image from URL", logger.Error(err))
		c.updateStatus(task.FileID, upload.StatusFailed, nil)
		return err
	}

	// 2. Validate image header
	contentType := c.detectContentType(originalData)
	if err := c.imageProc.ValidateImageHeader(originalData, contentType); err != nil {
		c.log.Error("Failed to validate image", logger.Error(err))
		c.updateStatus(task.FileID, upload.StatusFailed, nil)
		return err
	}

	// 3. Convert to WebP (or compressed format)
	processedData, err := c.imageProc.ConvertToWebP(originalData, contentType)
	if err != nil {
		c.log.Error("Failed to convert image", logger.Error(err))
		c.updateStatus(task.FileID, upload.StatusFailed, nil)
		return err
	}

	// 4. Upload processed image to MinIO
	processedObjectName := c.getWebPObjectName(task.ObjectName)
	processedURL, err := c.minioStorage.UploadStream(
		ctx,
		processedObjectName,
		bytes.NewReader(processedData),
		int64(len(processedData)),
		"image/png", // Using PNG as WebP fallback
	)
	if err != nil {
		c.log.Error("Failed to upload processed image", logger.Error(err))
		c.updateStatus(task.FileID, upload.StatusFailed, nil)
		return err
	}

	// 5. Update database status
	if err := c.updateStatus(task.FileID, upload.StatusCompleted, &processedURL); err != nil {
		c.log.Error("Failed to update upload status", logger.Error(err))
		// Don't return error - the processing succeeded, just the status update failed
	}

	c.log.Info("Image processing completed",
		logger.String("file_id", task.FileID),
		logger.String("processed_url", processedURL),
	)

	return nil
}

// downloadFromURL downloads file content from a URL
func (c *ImageProcessingConsumer) downloadFromURL(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// detectContentType detects content type from magic bytes
func (c *ImageProcessingConsumer) detectContentType(data []byte) string {
	if len(data) < 4 {
		return "application/octet-stream"
	}

	// JPEG
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return upload.ContentTypeJPEG
	}
	// PNG
	if len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return upload.ContentTypePNG
	}
	// GIF
	if len(data) >= 6 && data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return upload.ContentTypeGIF
	}

	return "application/octet-stream"
}

// getWebPObjectName converts object name to WebP version
// e.g., "uploads/inspection123/photo.jpg" -> "uploads/inspection123/photo.webp"
func (c *ImageProcessingConsumer) getWebPObjectName(originalName string) string {
	idx := strings.LastIndex(originalName, ".")
	if idx == -1 {
		return originalName + ".webp"
	}
	return originalName[:idx] + ".webp"
}

// updateStatus updates the upload status in database
func (c *ImageProcessingConsumer) updateStatus(fileID string, status string, processedURL *string) error {
	if c.uploadRepo == nil {
		return nil
	}
	return c.uploadRepo.UpdateStatus(fileID, status, processedURL)
}
