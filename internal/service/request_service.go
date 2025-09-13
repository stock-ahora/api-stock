package service

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/stock-ahora/api-stock/internal/models"
	"gorm.io/gorm"
)

type RequestService interface {
	List() ([]models.Request, error)
	Create(w http.ResponseWriter, r *http.Request)
	get(uuid uuid.UUID) (models.Request, error)
	//todo: agregar metodo para confirmar la request y agregar en el open api el metodo igual
	//todo: agregar metodo para modificar la request
}

type requestService struct {
	db    *gorm.DB
	s3Svc *S3Svc
}

func NewRequestService(db *gorm.DB, s3Svc *S3Svc) RequestService {
	return &requestService{db: db, s3Svc: s3Svc}
}

// implementación de metodos

func (rs requestService) List() ([]models.Request, error) {
	//TODO implement me
	panic("implement me")
}

func (rs requestService) Create(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (rs requestService) get(uuid uuid.UUID) (models.Request, error) {
	//TODO implement me
	panic("implement me")
}
