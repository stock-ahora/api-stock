package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/stock-ahora/api-stock/internal/dto"
	"github.com/stock-ahora/api-stock/internal/models"
	"gorm.io/gorm"
)

type RequestService interface {
	List() ([]models.Request, error)
	Create(*dto.CreateRequestDto) (models.Request, error)
	Get(uuid uuid.UUID) (models.Request, error)
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

func (r requestService) List() ([]models.Request, error) {
	//TODO implement me
	panic("implement me")
}

func (r requestService) Create(requestDto *dto.CreateRequestDto) (models.Request, error) {
	//TODO implement me

	key, err := r.s3Svc.doHandleUpload(requestDto, "requests/")

	if err != nil {
		return models.Request{}, err
	}

	uuidClient, _ := uuid.Parse("8d1b88f0-e5c7-4670-8bbb-3045f9ab3995")

	request := models.Request{
		ID:     uuid.New(),
		Status: models.RequestCreated,
		//ClientAccountID: requestDto.ClientAccountId,
		ClientAccountID: uuidClient,
		CreatedAt:       time.Now(),
	}

	document := models.Documents{
		ID:        uuid.New(),
		S3Path:    key,
		RequestID: request.ID,
		CreatedAt: time.Now(),
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Debug().Create(&request).Error; err != nil {
			return err
		}
		if err := tx.Debug().Create(&document).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return models.Request{}, err
	}

	return request, nil
}

func (r requestService) Get(uuid uuid.UUID) (models.Request, error) {
	//TODO implement me
	panic("implement me")
}
