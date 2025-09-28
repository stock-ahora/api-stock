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
	TypeStatusOut
)

func (t TypeStatus) String() string {
	return [...]string{"in", "on"}[t]
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
