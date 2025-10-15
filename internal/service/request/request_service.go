package request

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/smithy-go"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stock-ahora/api-stock/internal/dto"
	"github.com/stock-ahora/api-stock/internal/models"
	"github.com/stock-ahora/api-stock/internal/service/bedrock"
	"github.com/stock-ahora/api-stock/internal/service/eventservice"
	"github.com/stock-ahora/api-stock/internal/service/s3"
	"github.com/stock-ahora/api-stock/internal/service/textract"
	"gorm.io/gorm"
)

type RequestService interface {
	List() ([]models.Request, error)
	Create(*dto.CreateRequestDto) (models.Request, error)
	Get(uuid uuid.UUID) (models.Request, error)
	//todo: agregar metodo para confirmar la request y agregar en el open api el metodo igual
	//todo: agregar metodo para modificar la request
	Process(requestId uuid.UUID, clientAccountId uuid.UUID, typeIngress int) error
}

type requestService struct {
	db       *gorm.DB
	s3Svc    *s3.S3Svc
	eventSvc *eventservice.MQPublisher
	textract *textract.TextractService
}

func NewRequestService(db *gorm.DB, s3Svc *s3.S3Svc, eventSvc *eventservice.MQPublisher, textract *textract.TextractService) RequestService {
	return &requestService{db: db, s3Svc: s3Svc, eventSvc: eventSvc, textract: textract}
}

// implementación de metodos

func (r requestService) List() ([]models.Request, error) {
	//TODO implement me
	panic("implement me")
}

func (r requestService) Create(requestDto *dto.CreateRequestDto) (models.Request, error) {

	key, err := r.s3Svc.DoHandleUpload(requestDto, "requests/")

	if err != nil {
		return models.Request{}, err
	}

	requestUuid := uuid.New()

	request := models.Request{
		ID:              requestUuid,
		Status:          models.RequestCreated,
		ClientAccountID: requestDto.ClientAccountId,
		MovementTypeId:  requestDto.GetTypeStatus(),
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

	var event = createdRequestProcess(requestUuid, requestDto.ClientAccountId, requestDto.GetMovementToUpOrLessStock())

	err = r.eventSvc.PublishRequest(event)
	if err != nil {
		return request, err
	}

	return request, nil
}

func (r requestService) Get(uuid uuid.UUID) (models.Request, error) {
	//TODO implement me
	panic("implement me")
}

func createdRequestProcess(requestId uuid.UUID, clientAccountId uuid.UUID, movementType int) eventservice.RequestProcessEvent {
	return eventservice.RequestProcessEvent{
		RequestID:       requestId,
		ClientAccountId: clientAccountId,
		TypeIngress:     movementType,
	}
}

func (r requestService) Process(requestId uuid.UUID, clientAccountId uuid.UUID, typeIngress int) error {
	log.Printf("Procesando solicitud con ID: %s", requestId)

	var request models.Request
	result := r.db.Preload("Documents").First(&request, "id = ?", requestId)
	log.Printf("Resultado de la consulta: %+v", result)
	if result.Error != nil {
		log.Printf("Error al obtener la solicitud: %v", result.Error)
		return result.Error
	}
	log.Printf("Procesando con ID: %s", requestId)
	document := request.Documents[0]

	key := request.Documents[0].S3Path
	bucket := r.s3Svc.GetBucket()

	resultTextract, idAnalizys, err := r.textract.AnalyzeInvoice(context.Background(), bucket, key)
	if err != nil {
		var respErr *smithy.GenericAPIError
		if errors.As(err, &respErr) {
			fmt.Println("Code:", respErr.Code)
			fmt.Println("Message:", respErr.Message)
		} else {
			fmt.Printf("unexpected error: %v\n", err)
		}
	}
	document.TextractId = *idAnalizys

	log.Printf("Resultado de Textract para la solicitud %s:", requestId)

	ctx := context.Background()

	bedrockService := bedrock.NewBedrockService(ctx, bedrock.NOVA_PRO_AWS, "us-east-1")

	inputModel := textract.TablasToString(*resultTextract)

	log.Printf("Input para Bedrock: %s", inputModel)

	resultBedrock, err := bedrockService.FormatProduct(ctx, inputModel)

	if err != nil {
		log.Printf("Error al procesar con Bedrock: %v", err)
	}
	log.Printf("Resultado de Bedrock para la solicitud %s: %+v", requestId, resultBedrock)

	r.updateProduct(*resultBedrock, r.db, typeIngress, clientAccountId, requestId)

	return nil
}

func (r requestService) updateProduct(productsFind []bedrock.ProductResponse, db *gorm.DB, typeIngress int, clientAccountId uuid.UUID, requestId uuid.UUID) {

	listMovement := make([]eventservice.ProductPerMovement, 0, len(productsFind))

	for _, product := range productsFind {

		var requestSku models.Sku
		var productUpdate models.Product

		var existSku = false

		existSku = findSku(product, db, &requestSku, existSku)

		if existSku {
			countUpdate := product.Count * typeIngress

			_ = db.First(&productUpdate, "id = ?", &requestSku.ProductID)

			productUpdate.Stock = productUpdate.Stock + countUpdate

			db.Save(&productUpdate)

		} else {

			productUpdate.ID = uuid.New()
			productUpdate.Name = product.Name
			productUpdate.Stock = product.Count * typeIngress
			productUpdate.CreatedAt = time.Now()
			productUpdate.Status = "active"
			productUpdate.ClientAccount = clientAccountId

			db.Create(&productUpdate)

			requestSku.ID = uuid.New()
			requestSku.NameSku = product.SKUs[0]
			requestSku.Status = true
			requestSku.ProductID = productUpdate.ID
			requestSku.CreatedAt = time.Now()

			db.Create(&requestSku)
		}

		listMovement = append(listMovement, createMovement(productUpdate, product.Count, typeIngress))
	}

	movementsRequest := eventservice.MovementsEvent{
		Id:                 uuid.New(),
		ProductPerMovement: listMovement,
		RequestId:          requestId,
	}

	r.eventMovement(movementsRequest)
	notificationMovement()

}

func createMovement(product models.Product, count int, typeMovement int) eventservice.ProductPerMovement {

	return eventservice.ProductPerMovement{
		Id:             uuid.New().String(),
		ProductID:      product.ID,
		Count:          count,
		MovementId:     uuid.New(),
		MovementTypeId: dto.TypeMovement[typeMovement],
		CreatedAt:      time.Now(),
	}
}

func (r requestService) eventMovement(movements eventservice.MovementsEvent) {

	err := r.eventSvc.PublishMovements(movements)
	if err != nil {
		log.Printf("Error al publicar el evento de movimientos: %v", err)
	}

}

func notificationMovement() {
	//TODO implement me
}

func findSku(product bedrock.ProductResponse, db *gorm.DB, requestSku *models.Sku, existSku bool) bool {
	for _, sku := range product.SKUs {

		resultSku := db.Where("name_sku ILIKE ?", "%"+sku+"%").Find(&requestSku)
		if resultSku.Error != nil {
			log.Printf("Error al procesar con Bedrock: %v", resultSku.Error)
		}
		if resultSku.RowsAffected > 0 {
			existSku = true
			break
		}

	}
	return existSku
}
