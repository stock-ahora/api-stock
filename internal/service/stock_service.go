package service

import "github.com/stock-ahora/api-stock/internal/domain"

type StockService interface {
	List() ([]domain.Stock, error)
	Create(stock domain.Stock) (domain.Stock, error)
}
