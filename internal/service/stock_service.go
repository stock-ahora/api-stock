package service

import "github.com/stock-ahora/api-stock/internal/domain"

type StockRepository interface {
    FindAll() ([]domain.Stock, error)
    Save(stock domain.Stock) (domain.Stock, error)
}

type StockService interface {
    List() ([]domain.Stock, error)
    Create(stock domain.Stock) (domain.Stock, error)
}

type stockService struct {
    repo StockRepository
}

func NewStockService(repo StockRepository) StockService {
    return &stockService{repo: repo}
}

func (s stockService) List() ([]domain.Stock, error) {
    return s.repo.FindAll()
}

func (sstockService) Create(stock domain.Stock) (domain.Stock, error) {
    if stock.ID == "" {
        stock.ID = "1" // placeholder; luego usar UUID
    }
    return s.repo.Save(stock)
}
