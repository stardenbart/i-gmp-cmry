package uploadusecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/monitoring-system/backend/internal/domain/upload"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/uploadrepo"
	"github.com/monitoring-system/backend/pkg/storage"
)

// UploadUseCase handles file upload business logic
type UploadUseCase struct {
	minioStorage *storage.MinioStorage
	docxProc     *DOCXProcessor
	imageProc    *ImageProcessor
	uploadRepo   *uploadrepo.UploadRepository
}

// NewUploadUseCase creates a new upload use case instance
func NewUploadUseCase(
	minioStorage *storage.MinioStorage,
	docxProc *DOCXProcessor,
	imageProc *ImageProcessor,
	uploadRepo *uploadrepo.UploadRepository,
) *UploadUseCase {
	return &UploadUseCase{
		minioStorage: minioStorage,
		docxProc:     docxProc,
		imageProc:    imageProc,
		uploadRepo:   uploadRepo,
	}
}

// UploadFile handles the complete upload flow for both DOCX and image files
func (uc *UploadUseCase) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, req *upload.UploadRequest) (*upload.UploadResponse, error) {
	// 1. Validate and normalize file type
	ft := strings.ToLower(strings.TrimSpace(req.FileType))
	if ft == "initial" || ft == "followup" || ft == "wowr" || ft == "photo" || ft == "img" || ft == "image" {
		req.FileType = upload.FileTypeImage
	} else if ft == "docx" || ft == "doc" {
		req.FileType = upload.FileTypeDOCX
	}

	if req.FileType != upload.FileTypeDOCX && req.FileType != upload.FileTypeImage {
		return nil, fmt.Errorf("invalid file type: %s", req.FileType)
	}

	// 2. Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 3. Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 4. Validate based on file type
	if req.FileType == upload.FileTypeDOCX {
		if err := uc.validateDOCX(fileHeader, content); err != nil {
			return nil, err
		}
	} else {
		if err := uc.validateImage(fileHeader, content); err != nil {
			return nil, err
		}
	}

	// 5. Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	storedFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	objectName := fmt.Sprintf("uploads/%s/%s", req.InspectionID, storedFilename)

	// 6. Process based on file type
	var processedContent []byte
	var contentType string
	var status string

	if req.FileType == upload.FileTypeDOCX {
		// DOCX: Process synchronously (extract, compress images, repack)
		contentType = upload.ContentTypeDOCX
		processedContent, err = uc.docxProc.ProcessDOCX(bytes.NewReader(content), storedFilename)
		if err != nil {
			return nil, fmt.Errorf("failed to process DOCX: %w", err)
		}
		status = upload.StatusCompleted
	} else {
		// Image: Upload original first, then queue for async processing
		contentType = fileHeader.Header.Get("Content-Type")
		if contentType == "" {
			contentType = uc.detectContentType(content)
		}
		processedContent = content
		status = upload.StatusProcessing
	}

	// 7. Upload to MinIO
	fileURL, err := uc.minioStorage.UploadStream(
		ctx,
		objectName,
		bytes.NewReader(processedContent),
		int64(len(processedContent)),
		contentType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	// 8. Create upload record
	uploadRecord := &upload.Upload{
		ID:               uuid.New().String(),
		InspectionID:     req.InspectionID,
		OriginalFilename: fileHeader.Filename,
		StoredFilename:   storedFilename,
		FilePath:         objectName,
		FileSize:         int64(len(processedContent)),
		ContentType:      contentType,
		FileType:         req.FileType,
		Status:           status,
	}

	if uc.uploadRepo != nil {
		_ = uc.uploadRepo.Create(uploadRecord)
	}

	// 9. For images, queue for async processing
	if req.FileType == upload.FileTypeImage {
		processingMsg := &upload.ImageProcessingMessage{
			FileID:     uploadRecord.ID,
			SourceURL:  fileURL,
			ObjectName: objectName,
		}
		if err := uc.imageProc.QueueForProcessing(ctx, processingMsg); err != nil {
			// Log error but don't fail - the original is already uploaded
			fmt.Printf("Warning: failed to queue image for processing: %v\n", err)
		}
	}

	// 10. Build response
	response := &upload.UploadResponse{
		ID:          uploadRecord.ID,
		FileName:    fileHeader.Filename,
		FileURL:     fileURL,
		FileSize:    int64(len(processedContent)),
		ContentType: contentType,
		Status:      status,
	}

	return response, nil
}

// GetByID retrieves an upload by ID
func (uc *UploadUseCase) GetByID(id string) (*upload.Upload, error) {
	return uc.uploadRepo.GetByID(id)
}

// GetByInspectionID retrieves all uploads for an inspection
func (uc *UploadUseCase) GetByInspectionID(inspectionID string) ([]upload.Upload, error) {
	return uc.uploadRepo.GetByInspectionID(inspectionID)
}

// validateDOCX validates a DOCX file
func (uc *UploadUseCase) validateDOCX(header *multipart.FileHeader, content []byte) error {
	// Check extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !uc.isAllowedExtension(ext, upload.AllowedDOCXExtensions) {
		return upload.ErrInvalidExtension
	}

	// Check file size
	if header.Size > upload.MaxDOCXSize {
		return upload.ErrFileTooLarge
	}

	// Check magic bytes (DOCX is a ZIP file, starts with PK)
	if len(content) < 4 || content[0] != 0x50 || content[1] != 0x4B {
		return upload.ErrInvalidMagicBytes
	}

	// Check MIME type
	contentType := header.Header.Get("Content-Type")
	if contentType != "" && contentType != upload.ContentTypeDOCX {
		return upload.ErrInvalidMIMEType
	}

	return nil
}

// validateImage validates an image file
func (uc *UploadUseCase) validateImage(header *multipart.FileHeader, content []byte) error {
	// Check extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !uc.isAllowedExtension(ext, upload.AllowedImageExtensions) {
		return upload.ErrInvalidExtension
	}

	// Check file size
	if header.Size > upload.MaxImageSize {
		return upload.ErrFileTooLarge
	}

	// Validate magic bytes
	if err := uc.imageProc.ValidateImageHeader(content, uc.detectContentType(content)); err != nil {
		return err
	}

	// Check MIME type
	contentType := header.Header.Get("Content-Type")
	if contentType != "" {
		validType := false
		for _, allowed := range []string{upload.ContentTypeJPEG, upload.ContentTypePNG, upload.ContentTypeGIF, upload.ContentTypeWebP} {
			if contentType == allowed {
				validType = true
				break
			}
		}
		if !validType {
			return upload.ErrInvalidMIMEType
		}
	}

	return nil
}

// isAllowedExtension checks if an extension is in the allowed list
func (uc *UploadUseCase) isAllowedExtension(ext string, allowed []string) bool {
	for _, a := range allowed {
		if ext == a {
			return true
		}
	}
	return false
}

// detectContentType detects content type from file content
func (uc *UploadUseCase) detectContentType(content []byte) string {
	if len(content) < 4 {
		return "application/octet-stream"
	}

	// WebP (starts with RIFF and has WEBP at byte 8)
	if len(content) >= 12 && string(content[0:4]) == "RIFF" && string(content[8:12]) == "WEBP" {
		return upload.ContentTypeWebP
	}
	// JPEG
	if content[0] == 0xFF && content[1] == 0xD8 && content[2] == 0xFF {
		return upload.ContentTypeJPEG
	}
	// PNG
	if len(content) >= 8 && content[0] == 0x89 && content[1] == 0x50 && content[2] == 0x4E && content[3] == 0x47 {
		return upload.ContentTypePNG
	}
	// GIF
	if len(content) >= 6 && content[0] == 0x47 && content[1] == 0x49 && content[2] == 0x46 {
		return upload.ContentTypeGIF
	}

	return "application/octet-stream"
}
