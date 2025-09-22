package textract

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	"github.com/aws/aws-sdk-go-v2/service/textract/types"
)

type Service interface {
	AnalyzeInvoice(ctx context.Context, bucket, key string) (*textract.AnalyzeDocumentOutput, error)
}

type TextractService struct {
	client *textract.Client
}

func NewTextractService(region string) *TextractService {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	client := textract.NewFromConfig(cfg)
	return &TextractService{client: client}
}

func (s *TextractService) AnalyzeInvoice(ctx context.Context, bucket, key string) (*[]TablaProducto, *string, error) {
	input := &textract.StartDocumentAnalysisInput{
		DocumentLocation: &types.DocumentLocation{
			S3Object: &types.S3Object{
				Bucket: aws.String(bucket),
				Name:   aws.String(key),
			},
		},
		FeatureTypes: []types.FeatureType{
			types.FeatureTypeTables,
			types.FeatureTypeForms,
		},
	}

	resp, err := s.client.StartDocumentAnalysis(context.TODO(), input)
	if err != nil {
		log.Fatal(err)
	}
	jobId := resp.JobId

	fmt.Println("Job ID:", *resp.JobId)

	var allBlocks []types.Block
	for {
		result, err := s.client.GetDocumentAnalysis(ctx, &textract.GetDocumentAnalysisInput{
			JobId: aws.String(*jobId),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("error obteniendo resultado: %w", err)
		}

		fmt.Println("Estado actual:", result.JobStatus)

		if result.JobStatus == types.JobStatusSucceeded {
			// Recolectar todos los bloques, incluyendo paginación
			allBlocks = append(allBlocks, result.Blocks...)

			// Obtener bloques adicionales si hay paginación
			for result.NextToken != nil {
				nextResult, err := s.client.GetDocumentAnalysis(ctx, &textract.GetDocumentAnalysisInput{
					JobId:     aws.String(*jobId),
					NextToken: result.NextToken,
				})
				if err != nil {
					return nil, nil, fmt.Errorf("error en paginación: %w", err)
				}
				allBlocks = append(allBlocks, nextResult.Blocks...)
				result.NextToken = nextResult.NextToken
			}

			// Extraer datos de tablas
			tablas := extraerTablas(allBlocks)

			// Imprimir información para debug
			for i, tabla := range tablas {
				fmt.Printf("Tabla %d encontrada con %d filas\n", i+1, len(tabla.Filas))
			}

			// Devolver análisis con los bloques para procesamiento posterior
			return &tablas, resp.JobId, nil
		}

		if result.JobStatus == types.JobStatusFailed {
			return nil, nil, fmt.Errorf("el análisis falló")
		}

		// Esperar antes de volver a consultar
		time.Sleep(2 * time.Second)
	}

}

type TablaProducto struct {
	Filas []FilaProducto
}

type FilaProducto struct {
	Celdas []string
}

// extraerTablas encuentra todas las tablas en los bloques y extrae su contenido
func extraerTablas(blocks []types.Block) []TablaProducto {
	var tablas []TablaProducto

	// Mapeo de bloques por ID para referencia rápida
	blockMap := make(map[string]types.Block)
	for _, block := range blocks {
		if block.Id != nil {
			blockMap[*block.Id] = block
		}
	}

	// Buscar bloques de tipo TABLE
	for _, block := range blocks {
		if block.BlockType == types.BlockTypeTable {
			tabla := TablaProducto{}

			// Mapeo de celdas por coordenadas (fila, columna)
			cellsByPosition := make(map[int]map[int]string)

			// Para cada relación del bloque tabla, obtener las celdas
			if len(block.Relationships) > 0 {
				for _, cellID := range block.Relationships[0].Ids {
					cellBlock := blockMap[cellID]

					// Asegurarse de que sea una celda y tenga índices válidos
					if cellBlock.BlockType == types.BlockTypeCell &&
						cellBlock.RowIndex != nil &&
						cellBlock.ColumnIndex != nil {

						rowIdx := int(*cellBlock.RowIndex)
						colIdx := int(*cellBlock.ColumnIndex)

						// Inicializar el mapa de columnas si no existe
						if _, exists := cellsByPosition[int(rowIdx)]; !exists {
							cellsByPosition[rowIdx] = make(map[int]string)
						}

						// Extraer el texto de la celda
						cellText := ""
						if len(cellBlock.Relationships) > 0 {
							for _, wordID := range cellBlock.Relationships[0].Ids {
								wordBlock := blockMap[wordID]
								if wordBlock.Text != nil {
									if cellText != "" {
										cellText += " "
									}
									cellText += *wordBlock.Text
								}
							}
						}

						cellsByPosition[rowIdx][colIdx] = cellText
					}
				}
			}

			// Construir filas ordenadas
			maxRows := 0
			for rowIdx := range cellsByPosition {
				if rowIdx > maxRows {
					maxRows = rowIdx
				}
			}

			// Crear filas en orden
			for rowIdx := 1; rowIdx <= maxRows; rowIdx++ {
				if rowMap, exists := cellsByPosition[rowIdx]; exists {
					// Determinar número máximo de columnas
					maxCols := 0
					for colIdx := range rowMap {
						if colIdx > maxCols {
							maxCols = colIdx
						}
					}

					// Crear fila con celdas ordenadas
					fila := FilaProducto{
						Celdas: make([]string, maxCols),
					}

					for colIdx := 1; colIdx <= maxCols; colIdx++ {
						if text, exists := rowMap[colIdx]; exists {
							fila.Celdas[colIdx-1] = text
						}
					}

					tabla.Filas = append(tabla.Filas, fila)
				}
			}

			tablas = append(tablas, tabla)
		}
	}

	return tablas
}

func (t TablaProducto) String() string {
	var result string
	for i, fila := range t.Filas {
		if i > 0 {
			result += "\n"
		}
		for j, celda := range fila.Celdas {
			if j > 0 {
				result += "\t"
			}
			result += celda
		}
	}
	return result
}

func TablasToString(tablas []TablaProducto) string {
	var result string
	for i, tabla := range tablas {
		if i > 0 {
			result += "\n\n--- Tabla " + fmt.Sprintf("%d", i+1) + " ---\n\n"
		}
		result += tabla.String()
	}
	return result
}
