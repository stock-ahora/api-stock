package models

import (
	"time"

	"github.com/google/uuid"
)

type Sku struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	NameSku   string    `gorm:"type:varchar(255);not null" json:"name_sku"`
	Status    bool      `gorm:"not null" json:"status"`
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Product   Product   `gorm:"references:ID" json:"product,omitempty"`
}

func (Sku) TableName() string { return "sku" }
