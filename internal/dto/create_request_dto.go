package dto

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/stock-ahora/api-stock/internal/models"
)

type CreateRequestDto struct {
	File            multipart.File
	RequestType     string
	Type            TypeStatus
	FileName        string
	FileSize        int64
	FileType        string
	ClientAccountId uuid.UUID
}

type RequestListDto struct {
	ID              uuid.UUID            `json:"id"`
	RequestType     string               `json:"request_type"`
	Status          models.RequestStatus `json:"status"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	ClientAccountId uuid.UUID            `json:"client_account_id"`
}

type RequestDto struct {
	ID              uuid.UUID            `json:"id"`
	RequestType     string               `json:"request_type"`
	Status          models.RequestStatus `json:"status"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	ClientAccountId uuid.UUID            `json:"client_account_id"`
	Movements       []Movements          `json:"movements"`
}

type Movements struct {
	Id           uuid.UUID `json:"id"`
	Nombre       string    `json:"nombre"`
	Count        int       `json:"count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	TypeMovement int       `json:"type_movement"`
}

type ProductDto struct {
	ID          uuid.UUID `json:"id"`
	Referencial uuid.UUID `json:"referencial_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Stock       int       `json:"stock"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TypeStatus int

const (
	TypeStatusIn TypeStatus = iota
	TypeStatusOut
)

func (t TypeStatus) String() string {
	return [...]string{"in", "out"}[t]
}

func ParseTypeStatus(typeStr string) TypeStatus {
	switch typeStr {
	case "in":
		return TypeStatusIn
	case "out":
		return TypeStatusOut
	default:
		return TypeStatusIn
	}
}

func GetTypeMovementString(movement int) string {
	switch movement {
	case 1:
		return "in"
	case 2:
		return "out"
	default:
		return "unknown"
	}

}

func (s CreateRequestDto) GetTypeStatus() int {
	return int(s.Type) + 1
}

func (s CreateRequestDto) GetMovementToUpOrLessStock() int {
	switch s.Type {
	case TypeStatusIn:
		return 1
	case TypeStatusOut:
		return -1
	default:
		return 1
	}
}

var TypeMovement = map[int]int{
	-1: 2,
	1:  1,
}

type Page[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	TotalPages int   `json:"total_pages"`
}
