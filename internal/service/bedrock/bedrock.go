package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type ProductResponse struct {
	Name  string   `json:"name"`
	Count int      `json:"count"`
	SKUs  []string `json:"skus"`
}

type Service struct {
	client *bedrockruntime.Client
	model  string
}

func NewBedrockService(ctx context.Context, model string, region string) *Service {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
		return nil
	}
	client := bedrockruntime.NewFromConfig(cfg)
	return &Service{client: client, model: model}
}

func (b *Service) FormatProduct(ctx context.Context, input string) (*[]ProductResponse, error) {
	prompt := fmt.Sprintf(`Eres un asistente que formatea datos de productos.
Entrada: "%s"
Devuelve un JSON válido con este formato, en caso de que sean mas de unproducto devuelve un array de productos, cada 
producto con su nombre, cantidad y genera SKUs si no tiene ningun codigo o sku en el detalle con los nombres de 10 digitos 
para poder compararlos con una tabla que tengo en mi db, para la generacion del sku solo considera el nombre y no agregues
numeros si no tiene en el nombre en mayusculas, y solo genera max 3 skus si no viene algun codigo en el detalle, ademas considera la respuesta solo el json, no agregues texto adicional, ademas la cantidad
tienen que venir como numero entero, si no hay productos devuelve un array vacio:
{
  "name": "nombre del producto",
  "count": "cantidad de productos",
  "skus": ["sku1", "sku2", "sku3"]
}`, input)

	body := fmt.Sprintf(`{
      "messages": [
        {
          "role": "user",
          "content": [
            { "text": %q }
          ]
        }
      ],
      "inferenceConfig": {
        "maxTokens": 712,
        "temperature": 0
      }
    }`, prompt)

	resp, err := b.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(NOVA_PRO_AWS),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        []byte(body),
	})
	if err != nil {
		log.Println("Error invoking Bedrock model:", err)
		return nil, err
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(resp.Body, &parsed); err != nil {
		return nil, err
	}

	resultBytes, _ := json.Marshal(parsed)
	log.Println("Raw response:", string(resultBytes))

	products, err := parseBedrockResponse(resp.Body)
	if err != nil {
		return nil, err
	}
	return &products, nil
}

type Response struct {
	Output struct {
		Message struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			Role string `json:"role"`
		} `json:"message"`
	} `json:"output"`
	StopReason string                 `json:"stopReason"`
	Usage      map[string]interface{} `json:"usage"`
}

func parseBedrockResponse(respBody []byte) ([]ProductResponse, error) {
	// 1. Parsear la respuesta de Bedrock
	var br Response
	if err := json.Unmarshal(respBody, &br); err != nil {
		return nil, fmt.Errorf("error parsing Bedrock response: %w", err)
	}

	if len(br.Output.Message.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// 2. Extraer el texto
	raw := br.Output.Message.Content[0].Text

	// 3. Limpiar los backticks y el ```json
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	// 4. Unmarshal al modelo
	var products []ProductResponse
	if err := json.Unmarshal([]byte(raw), &products); err != nil {
		return nil, fmt.Errorf("error unmarshalling product json: %w", err)
	}

	return products, nil
}
