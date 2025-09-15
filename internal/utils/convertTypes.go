package utils

import (
	"fmt"
	"strconv"
)

func ConverToint(str string) (int, error) {
	portInt, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("error al convertir puerto a entero: %w", err)
	}
	return portInt, nil

}
