package models

import (
	"time"

	"github.com/google/uuid"
)

type DimFecha struct {
	FechaKey  int       `gorm:"primaryKey;column:fecha_key"`
	Fecha     time.Time `gorm:"column:fecha"`
	Dia       int       `gorm:"column:dia"`
	Mes       int       `gorm:"column:mes"`
	Anio      int       `gorm:"column:anio"`
	Trimestre int       `gorm:"column:trimestre"`
	// ...
}

func (DimFecha) TableName() string {
	return "dim_fecha"
}

// Dimensión de productos
type DimProducto struct {
	ID     int    `gorm:"primaryKey;column:id"`
	Nombre string `gorm:"column:nombre"`
	// ...
}

func (DimProducto) TableName() string {
	return "dim_producto"
}

// Dimensión de clientes
type DimCliente struct {
	ID     int    `gorm:"primaryKey;column:id"`
	Nombre string `gorm:"column:nombre"`
}

func (DimCliente) TableName() string {
	return "dim_cliente"
}

// Dimensión de tipos de movimiento
type DimTipoMovimiento struct {
	ID        int    `gorm:"primaryKey;column:id"`
	Nombre    string `gorm:"column:nombre"`
	Direccion int    `gorm:"column:direccion"` // +1 para ingreso, -1 para egreso
}

func (DimTipoMovimiento) TableName() string {
	return "dim_tipo_movimiento"
}

// Tabla de hechos de movimientos
type FactProductMovement struct {
	ID               int       `gorm:"primaryKey;column:id"`
	MovimientoUUID   uuid.UUID `gorm:"column:movimiento_uuid"`
	FechaKey         int       `gorm:"column:fecha_key"`
	ClienteID        int       `gorm:"column:cliente_id"`
	ProductoID       int       `gorm:"column:producto_id"`
	SKUId            *int      `gorm:"column:sku_id"`
	TipoMovimientoID int       `gorm:"column:tipo_movimiento_id"`
	Cantidad         int       `gorm:"column:cantidad"`
	Signo            int       `gorm:"column:signo"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (FactProductMovement) TableName() string {
	return "fact_product_movement"
}
