package uploadusecase

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/monitoring-system/backend/internal/domain/upload"
)

// DOCXProcessor handles DOCX file processing (extract, compress images, repack)
type DOCXProcessor struct{}

// NewDOCXProcessor creates a new DOCX processor instance
func NewDOCXProcessor() *DOCXProcessor {
	return &DOCXProcessor{}
}

// DOCXContent holds the extracted DOCX content
type DOCXContent struct {
	Files    map[string][]byte // filename -> content
	MediaDir string
}

// ProcessDOCX processes a DOCX file: extracts, compresses images, and repacks
func (p *DOCXProcessor) ProcessDOCX(reader io.Reader, filename string) ([]byte, error) {
	// 1. Extract the DOCX
	content, err := p.ExtractDOCX(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to extract DOCX: %w", err)
	}

	// 2. Compress images in the media directory
	if err := p.CompressImages(content); err != nil {
		return nil, fmt.Errorf("failed to compress images: %w", err)
	}

	// 3. Repack into DOCX
	result, err := p.RepackDOCX(content, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to repack DOCX: %w", err)
	}

	return result, nil
}

// ExtractDOCX extracts the contents of a DOCX file
func (p *DOCXProcessor) ExtractDOCX(reader io.Reader) (*DOCXContent, error) {
	content := &DOCXContent{
		Files:    make(map[string][]byte),
		MediaDir: "word/media/",
	}

	// Read all content into memory for processing
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return nil, fmt.Errorf("failed to read DOCX: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return nil, fmt.Errorf("failed to read ZIP: %w", err)
	}

	// Check for required structure
	hasDocumentXML := false
	for _, f := range zipReader.File {
		if f.Name == "word/document.xml" {
			hasDocumentXML = true
			break
		}
	}

	if !hasDocumentXML {
		return nil, upload.ErrCorruptedDOCX
	}

	// Extract all files
	for _, f := range zipReader.File {
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file in ZIP: %w", err)
		}

		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read file from ZIP: %w", err)
		}

		content.Files[f.Name] = data
	}

	return content, nil
}

// CompressImages compresses all images in the DOCX media directory
func (p *DOCXProcessor) CompressImages(content *DOCXContent) error {
	var errors []string

	for filename, data := range content.Files {
		// Check if this is an image in the media directory
		if !strings.HasPrefix(filename, content.MediaDir) {
			continue
		}

		// Determine image type by extension
		ext := strings.ToLower(filepath.Ext(filename))
		if ext == ".png" {
			compressed, err := p.compressPNG(data)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", filename, err))
				continue
			}
			content.Files[filename] = compressed
		} else if ext == ".jpg" || ext == ".jpeg" {
			compressed, err := p.compressJPEG(data)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", filename, err))
				continue
			}
			content.Files[filename] = compressed
		} else if ext == ".gif" {
			compressed, err := p.compressGIF(data)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", filename, err))
				continue
			}
			content.Files[filename] = compressed
		}
	}

	if len(errors) > 0 {
		// Log errors but don't fail - we can still proceed with uncompressed images
		fmt.Printf("Warning: some images failed to compress: %v\n", errors)
	}

	return nil
}

// compressPNG compresses a PNG image by recompressing with lower settings
func (p *DOCXProcessor) compressPNG(data []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("PNG decode error: %w", err)
	}

	var buf bytes.Buffer
	enc := &png.Encoder{
		CompressionLevel: png.BestSpeed, // Use best speed for smaller size, faster processing
	}
	if err := enc.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("PNG encode error: %w", err)
	}

	// Only use compressed version if it's actually smaller
	if buf.Len() < len(data) {
		return buf.Bytes(), nil
	}
	return data, nil
}

// compressJPEG compresses a JPEG image with quality settings and resize if needed
func (p *DOCXProcessor) compressJPEG(data []byte) ([]byte, error) {
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("JPEG decode error: %w", err)
	}

	// Resize if image is larger than 1920px wide
	bounds := img.Bounds()
	maxWidth := 1920
	if bounds.Dx() > maxWidth {
		img = p.resizeImage(img, maxWidth)
	}

	var buf bytes.Buffer
	opts := &jpeg.Options{
		Quality: 80, // 80% quality - good balance between size and quality
	}
	if err := jpeg.Encode(&buf, img, opts); err != nil {
		return nil, fmt.Errorf("JPEG encode error: %w", err)
	}

	// Only use compressed version if it's actually smaller
	if buf.Len() < len(data) {
		return buf.Bytes(), nil
	}
	return data, nil
}

// compressGIF compresses a GIF image
func (p *DOCXProcessor) compressGIF(data []byte) ([]byte, error) {
	img, err := gif.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("GIF decode error: %w", err)
	}

	var buf bytes.Buffer
	opts := &gif.Options{
		NumColors: 256, // Standard GIF colors
	}
	if err := gif.Encode(&buf, img, opts); err != nil {
		return nil, fmt.Errorf("GIF encode error: %w", err)
	}

	// Only use compressed version if it's actually smaller
	if buf.Len() < len(data) {
		return buf.Bytes(), nil
	}
	return data, nil
}

// resizeImage resizes an image to the specified max width while maintaining aspect ratio
func (p *DOCXProcessor) resizeImage(img image.Image, maxWidth int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width <= maxWidth {
		return img
	}

	// Calculate new height maintaining aspect ratio
	newHeight := (height * maxWidth) / width
	newWidth := maxWidth

	// Create new image with new dimensions
	newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Simple nearest-neighbor resize
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := (x * width) / newWidth
			srcY := (y * height) / newHeight
			newImg.Set(x, y, img.At(srcX, srcY))
		}
	}

	return newImg
}

// RepackDOCX repacks the content back into a DOCX file
func (p *DOCXProcessor) RepackDOCX(content *DOCXContent, filename string) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Get sorted file names for consistent ordering
	names := make([]string, 0, len(content.Files))
	for name := range content.Files {
		names = append(names, name)
	}

	// Write files to ZIP
	for _, name := range names {
		w, err := zipWriter.Create(name)
		if err != nil {
			return nil, fmt.Errorf("failed to create ZIP entry for %s: %w", name, err)
		}
		if _, err := w.Write(content.Files[name]); err != nil {
			return nil, fmt.Errorf("failed to write %s to ZIP: %w", name, err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close ZIP: %w", err)
	}

	return buf.Bytes(), nil
}

// CreateTempDir creates a temporary directory for processing
func (p *DOCXProcessor) CreateTempDir() (string, error) {
	return os.MkdirTemp("", "docx_processing_*")
}

// CleanupTempDir removes a temporary directory
func (p *DOCXProcessor) CleanupTempDir(dir string) error {
	return os.RemoveAll(dir)
}
