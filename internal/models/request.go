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
	MovementTypeId  int           `gorm:"column:movement_type_id;type:int;default:null"`
	Documents       []Documents   `gorm:"foreignKey:RequestID"`
}

type Documents struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	S3Path     string    `gorm:"column:s3_path;type:varchar"`
	RequestID  uuid.UUID `gorm:"column:request_id;type:uuid"`
	TextractId string    `gorm:"column:textract_id;type:varchar;default:null"`
	bedrockId  string    `gorm:"column:bedrock_id;type:varchar;default:null"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time `gorm:"column:update_at;autoUpdateTime"`
}

type RequestPerProduct struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ProductID  uuid.UUID `gorm:"column:product_id"`
	MovementID uuid.UUID `gorm:"column:movement_id"`
	RequestID  uuid.UUID `gorm:"column:request_id"`

	Product  Product  `gorm:"foreignKey:ProductID"`
	Movement Movement `gorm:"foreignKey:MovementID"`
}

type Movement struct {
	ID             uuid.UUID `gorm:"column:id;type:uuid;default:uuid_generate_v4();primaryKey"`
	Count          int       `gorm:"column:count"`
	ProductID      uuid.UUID `gorm:"column:product_id;type:uuid"` // si lo usas directamente
	DateLimit      time.Time `gorm:"column:date_limit"`
	RequestID      uuid.UUID `gorm:"column:request_id;type:uuid"`
	MovementTypeID int       `gorm:"column:movement_type_id;type:uuid"`

	CreatedAt time.Time `gorm:"column:create_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`

	// Relación N-N via tabla pivot
	Products []Product `gorm:"many2many:request_per_product;joinForeignKey:MovementID;joinReferences:ProductID"`
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

func (RequestPerProduct) TableName() string {
	return "request_per_product"
}

func (Movement) TableName() string { return "movement" }
