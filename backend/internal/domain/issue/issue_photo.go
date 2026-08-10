package issue

import (
	"context"
	"io"
	"time"
)

// PhotoType defines whether a photo is the initial finding or a follow-up proof.
type PhotoType string

const (
	PhotoTypeInitial  PhotoType = "Initial"
	PhotoTypeFollowUp PhotoType = "FollowUp"
	PhotoTypeWOWR     PhotoType = "WOWR"
)

// IssuePhoto represents the Issue_Photo table.
// One table handles both initial finding photos and follow-up evidence photos,
// differentiated by PhotoType field.
type IssuePhoto struct {
	IssuePhotoID     string     `gorm:"column:IssuePhotoID;primaryKey" json:"issue_photo_id"`
	IssueID          string     `gorm:"column:IssueID;not null" json:"issue_id"`
	RefPhotoID       *string    `gorm:"column:RefPhotoID;size:20" json:"ref_photo_id,omitempty"`
	PICUserID        string     `gorm:"column:PICUserID;not null" json:"pic_user_id"`
	PhotoType        PhotoType  `gorm:"column:PhotoType;size:20;not null" json:"photo_type"`
	ImageUrl         string     `gorm:"column:ImageUrl;size:255" json:"image_url"`
	FileName         string     `gorm:"column:FileName;size:255" json:"file_name"`
	Keterangan       string     `gorm:"column:Keterangan;size:255" json:"keterangan,omitempty"`
	HEIID            *string    `gorm:"column:HEIID;size:50" json:"hei_id,omitempty"`
	HEICategory      string     `gorm:"column:HEICategory;size:50" json:"hei_category,omitempty"`
	HabitID          *string    `gorm:"column:HabitID;size:50" json:"habit_id,omitempty"`
	EquipmentID      *string    `gorm:"column:EquipmentID;size:50" json:"equipment_id,omitempty"`
	InfrastructureID *string    `gorm:"column:InfrastructureID;size:50" json:"infrastructure_id,omitempty"`
	FollowUpDate     *time.Time `gorm:"column:FollowUpDate" json:"follow_up_date,omitempty"`
	JumlahFollowUp   *int       `gorm:"column:JumlahFollowUp" json:"jumlah_follow_up,omitempty"`
	PhotoCreatedAt   time.Time  `gorm:"column:PhotoCreatedAt;autoCreateTime" json:"created_at"`
	PhotoUpdatedAt   time.Time  `gorm:"column:PhotoUpdatedAt;autoUpdateTime" json:"updated_at"`

	UploaderName       string `gorm:"column:UploaderName;->" json:"uploader_name,omitempty"`
	HEIName            string `gorm:"column:HEIName;->" json:"hei_name,omitempty"`
	HabitName          string `gorm:"column:HabitName;->" json:"habit_name,omitempty"`
	EquipmentName      string `gorm:"column:EquipmentName;->" json:"equipment_name,omitempty"`
	InfrastructureName string `gorm:"column:InfrastructureName;->" json:"infrastructure_name,omitempty"`
}

func (IssuePhoto) TableName() string { return "Issue_Photo" }

// ─── DTOs ──────────────────────────────────────────────────────────────────

type UploadPhotoRequest struct {
	IssueID          string     `json:"issue_id" validate:"required"`
	RefPhotoID       *string    `json:"ref_photo_id,omitempty"`
	PICUserID        string     `json:"pic_user_id" validate:"required"`
	PhotoType        PhotoType  `json:"photo_type" validate:"required,oneof=Initial FollowUp WOWR"`
	Keterangan       string     `json:"keterangan"`
	HEIID            *string    `json:"hei_id,omitempty"`
	HEICategory      *string    `json:"hei_category,omitempty"`
	HabitID          *string    `json:"habit_id,omitempty"`
	EquipmentID      *string    `json:"equipment_id,omitempty"`
	InfrastructureID *string    `json:"infrastructure_id,omitempty"`
	FollowUpDate     *time.Time `json:"follow_up_date"`
	JumlahFollowUp   *int       `json:"jumlah_follow_up"`
}

type UpdatePhotoHEIRequest struct {
	HEIID            *string `json:"hei_id,omitempty"`
	HEICategory      *string `json:"hei_category,omitempty"`
	HabitID          *string `json:"habit_id,omitempty"`
	EquipmentID      *string `json:"equipment_id,omitempty"`
	InfrastructureID *string `json:"infrastructure_id,omitempty"`
}

// ─── Repository Interface ──────────────────────────────────────────────────

type IssuePhotoRepository interface {
	FindByIssueID(issueID string) ([]IssuePhoto, error)
	FindByID(id string) (*IssuePhoto, error)
	Create(p *IssuePhoto) error
	Update(p *IssuePhoto) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type IssuePhotoUseCase interface {
	GetByIssueID(issueID string) ([]IssuePhoto, error)
	Upload(ctx context.Context, req *UploadPhotoRequest, fileReader io.Reader, fileSize int64, originalFileName, contentType string) (*IssuePhoto, error)
	Update(ctx context.Context, photoID string, keterangan string) (*IssuePhoto, error)
	UpdateHEI(ctx context.Context, photoID string, req *UpdatePhotoHEIRequest) (*IssuePhoto, error)
	Delete(ctx context.Context, id string) error
}
