package consumer

import (
	"log"

	"github.com/stock-ahora/api-stock/internal/service/eventservice"
	"github.com/stock-ahora/api-stock/internal/service/request"
)

func handleMovement(e eventservice.MovementEvent) {
	log.Printf("➡️ Procesando MovementEvent: %+v", e)

	// aquí tu lógica de negocio
}

func handleRequestProcess(e eventservice.RequestProcessEvent, requestService request.RequestService) {
	log.Printf("📑 Procesando RequestProcessEvent: %+v", e)

	err := requestService.Process(e.RequestID, e.ClientAccountId)
	if err != nil {
		return
	}
	// aquí tu lógica de negocio
}
