package request

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode"

	"github.com/aws/smithy-go"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stock-ahora/api-stock/internal/config"
	"github.com/stock-ahora/api-stock/internal/dto"
	"github.com/stock-ahora/api-stock/internal/models"
	"github.com/stock-ahora/api-stock/internal/service/bedrock"
	"github.com/stock-ahora/api-stock/internal/service/eventservice"
	"github.com/stock-ahora/api-stock/internal/service/s3"
	"github.com/stock-ahora/api-stock/internal/service/textract"
	"gorm.io/gorm"
)

type RequestService interface {
	List(ctx context.Context, clientAccountId uuid.UUID, page, size int) (dto.Page[dto.RequestListDto], error)
	Create(*dto.CreateRequestDto, context.Context) (models.Request, error)
	Get(ctx context.Context, uuid uuid.UUID) (dto.RequestDto, error)
	//todo: agregar metodo para confirmar la request y agregar en el open api el metodo igual
	//todo: agregar metodo para modificar la request
	Process(ctx context.Context, requestId uuid.UUID, clientAccountId uuid.UUID, typeIngress int) error
	ProcessCtx(ctx context.Context, requestId uuid.UUID, clientAccountId uuid.UUID, typeIngress int) interface{}
}

type requestService struct {
	db       *gorm.DB
	s3Svc    *s3.S3Svc
	eventSvc *eventservice.MQPublisher
	textract *textract.TextractService
}

func (r requestService) ProcessCtx(ctx context.Context, requestId uuid.UUID, clientAccountId uuid.UUID, typeIngress int) interface{} {

	err := r.Process(ctx, requestId, clientAccountId, typeIngress)
	if err != nil {
		return nil
	}
	return nil
}

func NewRequestService(db *gorm.DB, s3Svc *s3.S3Svc, eventSvc *eventservice.MQPublisher, textract *textract.TextractService) RequestService {
	return &requestService{db: db, s3Svc: s3Svc, eventSvc: eventSvc, textract: textract}
}

func (r requestService) List(ctx context.Context, clientAccountId uuid.UUID, page, size int) (dto.Page[dto.RequestListDto], error) {

	db := config.GetDB()

	offset := (page - 1) * size

	var total int64
	if err := db.Model(&models.Request{}).
		Where("client_account_id = ?", clientAccountId).
		Count(&total).Error; err != nil {
		return dto.Page[dto.RequestListDto]{}, err
	}

	// DATA
	var requests []models.Request
	if err := db.
		Where("client_account_id = ?", clientAccountId).
		Order("create_at DESC").
		Limit(size).
		Offset(offset).
		Find(&requests).Error; err != nil {
		return dto.Page[dto.RequestListDto]{}, err
	}

	items := make([]dto.RequestListDto, 0, len(requests))
	for _, req := range requests {
		items = append(items, dto.RequestListDto{
			ID:              req.ID,
			RequestType:     dto.GetTypeMovementString(req.MovementTypeId),
			Status:          req.Status,
			CreatedAt:       req.CreatedAt,
			UpdatedAt:       req.UpdatedAt,
			ClientAccountId: req.ClientAccountID,
		})
	}

	return dto.Page[dto.RequestListDto]{
		Data:       items,
		Total:      total,
		Page:       page,
		Size:       size,
		TotalPages: int((total + int64(size) - 1) / int64(size)),
	}, nil
}

func (r requestService) Create(requestDto *dto.CreateRequestDto, ctx context.Context) (models.Request, error) {

	db := config.GetDB()

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

	err = db.Transaction(func(tx *gorm.DB) error {
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

	err = r.eventSvc.PublishRequest(event, ctx)
	if err != nil {
		return request, err
	}

	return request, nil
}

func (r requestService) Get(ctx context.Context, requestId uuid.UUID) (dto.RequestDto, error) {
	db := config.GetDB()

	var request models.Request
	Request := db.Preload("Documents").First(&request, "id = ?", requestId)
	if Request.Error != nil {
		return dto.RequestDto{}, Request.Error
	}

	var rpp []models.RequestPerProduct

	err := db.
		Preload("Product").
		Preload("Movement").
		Where("request_id = ?", requestId).
		Find(&rpp).Error
	if err != nil {
		return dto.RequestDto{}, err
	}

	// mapear al DTO
	movements := make([]dto.Movements, 0, len(rpp))
	for _, x := range rpp {
		movements = append(movements, dto.Movements{
			Id:        x.Movement.ID,
			Nombre:    x.Product.Name,
			Count:     x.Movement.Count,
			CreatedAt: x.Movement.CreatedAt,
			UpdatedAt: x.Movement.UpdatedAt,
		})
	}

	requestDto := dto.RequestDto{
		ID:              requestId,
		RequestType:     dto.GetTypeMovementString(request.MovementTypeId),
		Status:          request.Status,
		CreatedAt:       request.CreatedAt,
		UpdatedAt:       request.UpdatedAt,
		ClientAccountId: request.ClientAccountID,
		Movements:       movements,
	}

	return requestDto, nil

}

func createdRequestProcess(requestId uuid.UUID, clientAccountId uuid.UUID, movementType int) eventservice.RequestProcessEvent {
	return eventservice.RequestProcessEvent{
		RequestID:       requestId,
		ClientAccountId: clientAccountId,
		TypeIngress:     movementType,
	}
}

func (r requestService) Process(ctx context.Context, requestId uuid.UUID, clientAccountId uuid.UUID, typeIngress int) error {

	db := config.GetDB()

	log.Printf("Procesando solicitud con ID: %s", requestId)

	var request models.Request
	result := db.Preload("Documents").First(&request, "id = ?", requestId)
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

	bedrockService := bedrock.NewBedrockService(ctx, bedrock.NOVA_PRO_AWS, "us-east-1")

	inputModel := textract.TablasToString(*resultTextract)

	log.Printf("Input para Bedrock: %s", inputModel)

	resultBedrock, err := bedrockService.FormatProduct(ctx, inputModel, bedrock.ProductoPrompt)

	if err != nil {
		log.Printf("Error al procesar con Bedrock: %v", err)
	}
	log.Printf("Resultado de Bedrock para la solicitud %s: %+v", requestId, resultBedrock)

	request.Status = models.RequestStatusPending

	r.updateProduct(ctx, *resultBedrock, db, typeIngress, clientAccountId, requestId)

	db.Save(&request)

	return nil
}

func (r requestService) updateProduct(ctx context.Context, productsFind []bedrock.ProductResponse, db *gorm.DB, typeIngress int, clientAccountId uuid.UUID, requestId uuid.UUID) {

	listMovement := make([]eventservice.ProductPerMovement, 0, len(productsFind))

	for _, product := range productsFind {

		var requestSku models.Sku
		var productUpdate models.Product

		var existSku = false

		existSku = findSku(product, db, &requestSku, existSku, ctx)

		if existSku {
			countUpdate := product.Count * typeIngress

			_ = db.First(&productUpdate, "id = ?", &requestSku.ProductID)

			productUpdate.Stock = productUpdate.Stock + countUpdate

			db.Save(&productUpdate) //
		} else {

			productUpdate.ID = uuid.New()
			productUpdate.Name = product.Name
			productUpdate.Stock = product.Count * typeIngress
			productUpdate.CreatedAt = time.Now()
			productUpdate.Status = "active"
			productUpdate.ClientAccount = clientAccountId

			db.Create(&productUpdate)

			db.Save(&productUpdate) //
			requestSku.ID = uuid.New()
			requestSku.NameSku = product.SKUs[0]
			requestSku.Status = true
			requestSku.ProductID = productUpdate.ID
			requestSku.CreatedAt = time.Now()

			db.Create(&requestSku)

			db.Save(&requestSku) //
		}

		movement := createMovement(productUpdate, product.Count, typeIngress)
		listMovement = append(listMovement, movement)
		r.publicProductEtl(productUpdate, requestSku, clientAccountId, movement, typeIngress, requestId)
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

func (r requestService) publicProductEtl(product models.Product, sku models.Sku, id uuid.UUID, movement eventservice.ProductPerMovement, typeIngress int, requestId uuid.UUID) {

	err := r.eventSvc.PublishProductEtl(eventservice.ProductEvent{
		ProductoID:      product.ID.String(),
		NombreProducto:  product.Name,
		ClienteID:       id.String(),
		Cantidad:        movement.Count,
		Signo:           typeIngress,
		Fecha:           movement.CreatedAt,
		SolicitudId:     requestId.String(),
		StatusSolicitud: "pending",
		TipoMovimiento:  fmt.Sprintf("%d", movement.MovementTypeId),
	})
	if err != nil {
		return
	}
}

func notificationMovement() {
	//TODO implement me
}

func findSku(product bedrock.ProductResponse, db *gorm.DB, requestSku *models.Sku, existSku bool, ctx context.Context) bool {
	for _, sku := range product.SKUs {

		skuNormalized := normalizeSKU(sku)

		resultSku := db.Where("name_sku ILIKE ?", "%"+skuNormalized+"%").Find(&requestSku)
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

func normalizeSKU(s string) string {
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	// Elimina caracteres no alfanuméricos
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		}
		return -1
	}, s)
	return s
}
