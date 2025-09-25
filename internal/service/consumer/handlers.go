package consumer

import (
	"log"

	"github.com/stock-ahora/api-stock/internal/service/eventservice"
	"github.com/stock-ahora/api-stock/internal/service/request"
)

func handleRequestProcess(e eventservice.RequestProcessEvent, requestService request.RequestService) {
	log.Printf("📑 Procesando RequestProcessEvent: %+v", e)

	if err := requestService.Process(e.RequestID, e.ClientAccountId); err != nil {
		log.Printf("❌ error procesando request: %v", err)
	}
}
