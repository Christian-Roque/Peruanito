package utils

import (
	"encoding/csv"
	"fmt"
	"os"
)

func AnalizarDataset(path string) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error abriendo:", path)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error leyendo:", path)
		return
	}

	totalFilas := len(records) - 1
	totalColumnas := len(records[0])

	nulos := 0

	for i, row := range records {
		if i == 0 {
			continue
		}

		if len(row) < 6 || row[5] == "" {
			nulos++
		}
	}

	fmt.Println("===== EDA =====")
	fmt.Println("Archivo:", path)
	fmt.Println("Filas:", totalFilas)
	fmt.Println("Columnas:", totalColumnas)
	fmt.Println("Sumillas vacías:", nulos)

	// Mostrar ejemplo
	fmt.Println("Ejemplo SUMILLA:")
	for i := 1; i < len(records) && i <= 3; i++ {
		fmt.Println("-", records[i][5])
	}

	fmt.Println("================")
}
