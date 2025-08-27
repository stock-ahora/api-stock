package repository

import "github.com/stock-ahora/api-stock/internal/domain"

type memoryStockRepo struct {
    data []domain.Stock
}

func NewMemoryStockRepo() memoryStockRepo {
    return &memoryStockRepo{data: make([]domain.Stock, 0)}
}

func (mmemoryStockRepo) FindAll() ([]domain.Stock, error) {
    return m.data, nil
}

func (m *memoryStockRepo) Save(stock domain.Stock) (domain.Stock, error) {
    m.data = append(m.data, stock)
    return stock, nil
}
