package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID            uuid.UUID `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ReferencialID uuid.UUID `gorm:"type:uuid;not null" json:"referencial_id"`
	Name          string    `gorm:"type:varchar(255);not null" json:"name"`
	Description   string    `gorm:"type:varchar(255);not null" json:"description"`
	Stock         int64     `gorm:"not null" json:"stock"`
	Status        string    `gorm:"type:varchar(50)" json:"status"`
	ClientAccount uuid.UUID `gorm:"column:client_account_id;type:uuid;not null" json:"client_account_id"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:update_at;autoUpdateTime" json:"updated_at"`
	Sku           []Sku     `gorm:"foreignKey:ProductID" json:"sku,omitempty"`
}

func (Product) TableName() string {
	return "product"
}
