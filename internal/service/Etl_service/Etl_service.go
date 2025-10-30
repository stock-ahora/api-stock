package Etl_service

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/stock-ahora/api-stock/internal/service/eventservice"
	"gorm.io/gorm"
)

type EtlService struct {
	Db *gorm.DB
}

func (e EtlService) StartETLConsumer(evt eventservice.ProductEvent) {

	// --- TRANSFORMACIÓN + CARGA ---
	var clienteID int
	e.Db.Raw("SELECT id FROM dim_cliente WHERE cliente_uuid = ?", evt.ClienteID).Scan(&clienteID)

	productoID := e.getProductId(evt)

	fechaKey := e.getFechaKey(evt)

	solicitudID := e.getSolicitudId(evt, clienteID)

	// Insertar fila en la tabla de hechos
	e.Db.Exec(`
      INSERT INTO fact_product_movement (producto_id, cliente_id, cantidad, signo, tipo_movimiento_id, fecha_key, solicitudID, created_at)
      VALUES (?, ?, ?, ?, (SELECT id FROM dim_tipo_movimiento WHERE nombre = ? LIMIT 1), ?, ?, NOW())
    `, productoID, clienteID, evt.Cantidad, evt.Signo, evt.TipoMovimiento, fechaKey, solicitudID)

}

func (e EtlService) getSolicitudId(evt eventservice.ProductEvent, clienteID int) int {
	var solicitudID int
	if evt.SolicitudId != "" {
		e.Db.Raw("SELECT id FROM dim_solicitud WHERE solicitud_uuid = ?", evt.SolicitudId).Scan(&solicitudID)
		if solicitudID == 0 {
			// Si no existe, creamos la nueva solicitud con los datos disponibles
			clienteUUID, _ := uuid.Parse(evt.ClienteID)

			err := e.Db.Exec(`
				INSERT INTO dim_solicitud (solicitud_uuid, cliente_uuid, status, creado_en)
				VALUES (?, ?, ?, NOW())
			`, evt.SolicitudId, clienteUUID, evt.StatusSolicitud).Error
			if err != nil {
				log.Println(err)
			}
			e.Db.Raw("SELECT id FROM dim_solicitud WHERE solicitud_uuid = ?", evt.SolicitudId).Scan(&solicitudID)
		}
	}
	return solicitudID
}

func (e EtlService) getProductId(evt eventservice.ProductEvent) int {
	var productoID int
	e.Db.Raw("SELECT id FROM dim_producto WHERE producto_uuid = ?", evt.ProductoID).Scan(&productoID)
	if productoID == 0 {
		// Si el producto no existe en la dimensión, lo insertamos
		e.Db.Exec(`
		INSERT INTO dim_producto (producto_uuid, nombre, creado_en, stock, status, cliente_uuid)
		VALUES (?, ?, ?, ?, ?, ?)
	  `, evt.ProductoID, evt.NombreProducto, time.Now(), evt.Cantidad*evt.Signo, "activo", evt.ClienteID)
		// Obtenemos el ID del producto recién insertado
		e.Db.Raw("SELECT id FROM dim_producto WHERE producto_uuid = ?", evt.ProductoID).Scan(&productoID)

	} else {
		// Si el producto ya existe, actualizamos su stock
		e.Db.Exec(`
		UPDATE dim_producto
		SET stock = stock + ?
		WHERE producto_uuid = ?
	  `, evt.Cantidad*evt.Signo, evt.ProductoID)

	}
	return productoID
}

func (e EtlService) getFechaKey(evt eventservice.ProductEvent) int {
	fechaMovimiento := evt.Fecha
	// Extraer partes
	dia := fechaMovimiento.Day()
	mes := int(fechaMovimiento.Month())
	anio := fechaMovimiento.Year()
	trimestre := (mes-1)/3 + 1
	diaSemana := int(fechaMovimiento.Weekday())
	nombreMes := fechaMovimiento.Month().String()
	nombreDia := fechaMovimiento.Weekday().String()
	fechaKey := anio*10000 + mes*100 + dia // Ej: 20251030

	// 3️⃣ Buscar si la fecha ya está en dim_fecha
	var existe bool
	e.Db.Raw("SELECT EXISTS(SELECT 1 FROM dim_fecha WHERE fecha_key = ?)", fechaKey).Scan(&existe)
	if !existe {
		// Insertar si no existe
		err := e.Db.Exec(`
			INSERT INTO dim_fecha (fecha_key, fecha, dia, mes, anio, trimestre, dia_semana, nombre_dia, nombre_mes)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			fechaKey, fechaMovimiento, dia, mes, anio, trimestre, diaSemana, nombreDia, nombreMes).Error
		if err != nil {
			log.Println(err)
		}
	}
	return fechaKey
}
