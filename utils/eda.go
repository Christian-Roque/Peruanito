package utils

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

func AnalizarDatasetGlobal(archivos []string) {
	totalFilas := 0
	totalColumnas := 0
	totalNulos := 0
	ejemplos := []string{}

	for _, path := range archivos {

		file, err := os.Open(path)
		if err != nil {
			fmt.Println("Error abriendo:", path)
			continue
		}

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		file.Close()

		if err != nil || len(records) == 0 {
			fmt.Println("Error leyendo:", path)
			continue
		}

		headers := records[0]
		totalColumnas += len(headers)

		// 🔹 Buscar columna SUMILLA dinámicamente
		indiceSumilla := -1
		for i, h := range headers {
			h = strings.ToLower(strings.TrimSpace(h))
			if strings.Contains(h, "sumilla") {
				indiceSumilla = i
				break
			}
		}

		// fallback
		if indiceSumilla == -1 {
			indiceSumilla = 5
		}

		for i, row := range records {
			if i == 0 {
				continue
			}

			totalFilas++

			if len(row) <= indiceSumilla || row[indiceSumilla] == "" {
				totalNulos++
			} else if len(ejemplos) < 3 {
				ejemplos = append(ejemplos, row[indiceSumilla])
			}
		}
	}

	fmt.Println("===== EDA GLOBAL =====")
	fmt.Println("Total archivos:", len(archivos))
	fmt.Println("Total registros:", totalFilas)
	fmt.Println("Promedio columnas:", totalColumnas/len(archivos))
	fmt.Println("Sumillas vacías:", totalNulos)

	fmt.Println("\nEjemplos SUMILLA:")
	for _, e := range ejemplos {
		fmt.Println("-", e)
	}

	fmt.Println("======================")
}
