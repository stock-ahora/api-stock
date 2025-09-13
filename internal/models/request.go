package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Request struct {
	gorm.Model
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ClientAccountID uuid.UUID `gorm:"type:uuid;not null" json:"client_account_id"`
	Status          string    `gorm:"type:varchar(50)" json:"status"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
