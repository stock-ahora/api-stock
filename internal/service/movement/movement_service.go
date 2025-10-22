package movement

import (
	"github.com/google/uuid"
	"github.com/stock-ahora/api-stock/internal/dto"
	"github.com/stock-ahora/api-stock/internal/models"
	"gorm.io/gorm"
)

type MovementService interface {
	List(productId uuid.UUID, page, size int) (dto.Page[dto.Movements], error)
}

type movementService struct {
	db *gorm.DB
}

func NewMovementService(db *gorm.DB) MovementService {
	return &movementService{db: db}
}

func (m movementService) List(productId uuid.UUID, page, size int) (dto.Page[dto.Movements], error) {
	offset := (page - 1) * size

	var total int64
	if err := m.db.Model(&models.Movement{}).
		Where("product_id = ?", productId).
		Count(&total).Error; err != nil {
		return dto.Page[dto.Movements]{}, err
	}

	var movements []models.Movement
	if err := m.db.
		Where("product_id = ?", productId).
		Order("create_at DESC").
		Limit(size).
		Offset(offset).
		Find(&movements).Error; err != nil {
		return dto.Page[dto.Movements]{}, err
	}

	items := make([]dto.Movements, 0, len(movements))
	for _, req := range movements {
		items = append(items, dto.Movements{
			Id:        req.ID,
			Count:     req.Count,
			CreatedAt: req.CreatedAt,
			UpdatedAt: req.UpdatedAt,
		})
	}

	return dto.Page[dto.Movements]{
		Data:       items,
		Total:      total,
		Page:       page,
		Size:       size,
		TotalPages: int((total + int64(size) - 1) / int64(size)),
	}, nil
}
