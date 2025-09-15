package models

import (
	"time"

	"github.com/google/uuid"
)

type Request struct {
	ID              uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ClientAccountID uuid.UUID     `gorm:"column:client_account_id;type:uuid;not null"`
	Status          RequestStatus `gorm:"type:varchar(50)"`
	CreatedAt       time.Time     `gorm:"column:create_at;autoCreateTime"`
	UpdatedAt       time.Time     `gorm:"column:updated_at;autoUpdateTime"`
}

type Documents struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	S3Path    string    `gorm:"column:s3_path;type:varchar"`
	RequestID uuid.UUID `gorm:"column:request_id;type:uuid"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:update_at;autoUpdateTime"`
}

type RequestStatus string

const (
	RequestCreated         RequestStatus = "created"
	RequestStatusPending   RequestStatus = "pending"
	RequestStatusApproved  RequestStatus = "approved"
	RequestStatusRejected  RequestStatus = "rejected"
	RequestStatusCancelled RequestStatus = "cancelled"
)

func (s RequestStatus) IsValid() bool {
	switch s {
	case RequestCreated, RequestStatusPending, RequestStatusApproved, RequestStatusRejected, RequestStatusCancelled:
		return true
	}
	return false
}

func (Request) TableName() string {
	return "request"
}

func (Documents) TableName() string {
	return "documents"
}
