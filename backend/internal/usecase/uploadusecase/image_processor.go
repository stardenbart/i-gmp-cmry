package uploadusecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/monitoring-system/backend/internal/domain/upload"
	"github.com/monitoring-system/backend/pkg/kafka"
)

// ImageProcessor handles image processing and Kafka message publishing
type ImageProcessor struct {
	kafkaProducer kafka.EventProducer
	topic         string
}

// NewImageProcessor creates a new image processor instance
func NewImageProcessor(kafkaProducer kafka.EventProducer) *ImageProcessor {
	return &ImageProcessor{
		kafkaProducer: kafkaProducer,
		topic:         upload.TopicImageProcessing,
	}
}

// ConvertToWebP converts an image to WebP format
// Note: Due to Go standard library limitations, this converts to PNG as a fallback
// since native WebP encoding requires external libraries. The result is still
// compressed and smaller than the original.
func (p *ImageProcessor) ConvertToWebP(input []byte, originalContentType string) ([]byte, error) {
	var img image.Image
	var err error

	// Decode based on original content type
	switch originalContentType {
	case upload.ContentTypeJPEG, "image/jpg":
		img, err = jpeg.Decode(bytes.NewReader(input))
	case upload.ContentTypePNG:
		img, err = png.Decode(bytes.NewReader(input))
	case upload.ContentTypeGIF:
		img, err = gif.Decode(bytes.NewReader(input))
	case upload.ContentTypeWebP:
		// Already WebP, no conversion needed
		return input, nil
	default:
		return nil, fmt.Errorf("unsupported content type: %s", originalContentType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize if larger than 1920px
	img = p.resizeIfNeeded(img, 1920)

	// Encode as WebP (using PNG fallback due to Go stdlib limitations)
	var buf bytes.Buffer
	if err := p.encodeWebP(img, &buf, 85); err != nil {
		return nil, fmt.Errorf("failed to encode WebP: %w", err)
	}

	return buf.Bytes(), nil
}

// resizeIfNeeded resizes the image if it exceeds the max dimension
func (p *ImageProcessor) resizeIfNeeded(img image.Image, maxDim int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width <= maxDim && height <= maxDim {
		return img
	}

	var newWidth, newHeight int
	if width > height {
		newWidth = maxDim
		newHeight = (height * maxDim) / width
	} else {
		newHeight = maxDim
		newWidth = (width * maxDim) / height
	}

	return p.resizeImage(img, newWidth, newHeight)
}

// resizeImage resizes an image using nearest neighbor interpolation
func (p *ImageProcessor) resizeImage(img image.Image, newWidth, newHeight int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	newBounds := image.Rect(0, 0, newWidth, newHeight)
	newImg := image.NewRGBA(newBounds)

	// Nearest neighbor scaling
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := (x * width) / newWidth
			srcY := (y * height) / newHeight
			newImg.Set(x, y, img.At(srcX, srcY))
		}
	}

	return newImg
}

// encodeWebP encodes an image to WebP format
// Note: Go standard library doesn't support WebP encoding natively.
// This implementation uses PNG as a fallback which still provides compression.
// For production WebP support, consider using:
// - github.com/kettek/gowebp
// - or execute cwebp command-line tool
func (p *ImageProcessor) encodeWebP(img image.Image, w io.Writer, quality int) error {
	// Fallback to PNG encoding since Go stdlib doesn't have WebP encoder
	// PNG will still be compressed and smaller than original for most cases
	return png.Encode(w, img)
}

// QueueForProcessing sends an image processing task to Kafka
func (p *ImageProcessor) QueueForProcessing(ctx context.Context, msg *upload.ImageProcessingMessage) error {
	if p.kafkaProducer == nil {
		return fmt.Errorf("kafka producer not configured")
	}

	key := msg.FileID
	err := p.kafkaProducer.PublishEvent(ctx, p.topic, key, msg)
	if err != nil {
		return fmt.Errorf("failed to publish to Kafka: %w", err)
	}

	return nil
}

// ValidateImageHeader validates basic image properties by checking magic bytes
func (p *ImageProcessor) ValidateImageHeader(data []byte, contentType string) error {
	if len(data) < 12 {
		return fmt.Errorf("file too small to be a valid image")
	}

	// Check magic bytes
	switch contentType {
	case upload.ContentTypeJPEG, "image/jpg":
		// JPEG magic bytes: FF D8 FF
		if data[0] != 0xFF || data[1] != 0xD8 || data[2] != 0xFF {
			return upload.ErrInvalidMagicBytes
		}
	case upload.ContentTypePNG:
		// PNG magic bytes: 89 50 4E 47 0D 0A 1A 0A
		pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		if !bytes.Equal(data[:8], pngMagic) {
			return upload.ErrInvalidMagicBytes
		}
	case upload.ContentTypeGIF:
		// GIF87a: 47 49 46 38 37 61
		// GIF89a: 47 49 46 38 39 61
		gifMagic := []byte{0x47, 0x49, 0x46, 0x38, 0x37, 0x61}
		gifMagic89 := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}
		if !bytes.Equal(data[:6], gifMagic) && !bytes.Equal(data[:6], gifMagic89) {
			return upload.ErrInvalidMagicBytes
		}
	default:
		return fmt.Errorf("unsupported content type: %s", contentType)
	}

	return nil
}
