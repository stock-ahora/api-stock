package dto

import (
	"mime/multipart"

	"github.com/google/uuid"
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

type TypeStatus int

const (
	TypeStatusIn TypeStatus = iota
	TypeStatusOn
)

func (t TypeStatus) String() string {
	return [...]string{"in", "on"}[t]
}

func ParseTypeStatus(typeStr string) TypeStatus {
	switch typeStr {
	case "in":
		return TypeStatusIn
	case "on":
		return TypeStatusOn
	default:
		return TypeStatusIn
	}
}
