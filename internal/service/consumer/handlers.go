package consumer

import (
	"log"

	"github.com/stock-ahora/api-stock/internal/service/eventservice"
)

func handleMovement(e eventservice.MovementEvent) {
	log.Printf("➡️ Procesando MovementEvent: %+v", e)
	// aquí tu lógica de negocio
}

func handleRequestProcess(e eventservice.RequestProcessEvent) {
	log.Printf("📑 Procesando RequestProcessEvent: %+v", e)
	// aquí tu lógica de negocio
}
