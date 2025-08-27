package repository

import "github.com/stock-ahora/api-stock/internal/domain"

type MemoryStockRepo struct {
	data []domain.Stock
}

func NewMemoryStockRepo() *MemoryStockRepo {
	return &MemoryStockRepo{data: make([]domain.Stock, 0)}
}

func (m *MemoryStockRepo) FindAll() ([]domain.Stock, error) {
	return m.data, nil
}

func (m *MemoryStockRepo) Save(stock domain.Stock) (domain.Stock, error) {
	m.data = append(m.data, stock)
	return stock, nil
}
