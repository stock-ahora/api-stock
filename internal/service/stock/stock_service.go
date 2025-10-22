package stock

import (
	"github.com/google/uuid"
	"github.com/stock-ahora/api-stock/internal/dto"
	"github.com/stock-ahora/api-stock/internal/models"
	"gorm.io/gorm"
)

type StockService interface {
	List(clientAccountId uuid.UUID, page, size int) (dto.Page[dto.ProductDto], error)
	Get(productId uuid.UUID) (dto.ProductDto, error)
}

type stockService struct {
	db *gorm.DB
}

func NewStockService(db *gorm.DB) StockService {
	return &stockService{db: db}
}

func (s stockService) List(clientAccountId uuid.UUID, page, size int) (dto.Page[dto.ProductDto], error) {
	offset := (page - 1) * size
	var total int64
	if err := s.db.Model(&models.Product{}).
		Where("client_account_id = ?", clientAccountId).
		Count(&total).Error; err != nil {
		return dto.Page[dto.ProductDto]{}, err
	}

	var products []models.Product
	if err := s.db.
		Where("client_account_id = ?", clientAccountId).
		Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&products).Error; err != nil {
		return dto.Page[dto.ProductDto]{}, err
	}

	items := make([]dto.ProductDto, 0, len(products))
	for _, req := range products {
		items = append(items, dto.ProductDto{
			ID:          req.ID,
			Referencial: req.ReferencialID,
			Status:      req.Status,
			Name:        req.Name,
			Description: req.Description,
			Stock:       req.Stock,
			CreatedAt:   req.CreatedAt,
			UpdatedAt:   req.UpdatedAt,
		})
	}

	return dto.Page[dto.ProductDto]{
		Data:       items,
		Total:      total,
		Page:       page,
		Size:       size,
		TotalPages: int((total + int64(size) - 1) / int64(size)),
	}, nil

}

func (s stockService) Get(productId uuid.UUID) (dto.ProductDto, error) {
	var product models.Product

	err := s.db.
		Where("id = ?", productId).
		Find(&product).Error
	if err != nil {
		return dto.ProductDto{}, err
	}

	productDto := dto.ProductDto{
		ID:          product.ID,
		Referencial: product.ReferencialID,
		Name:        product.Name,
		Description: product.Description,
		Stock:       product.Stock,
		Status:      product.Status,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}

	return productDto, nil
}
