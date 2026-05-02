package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"proyecto-nlp/utils"
)

// 🔹 Obtener todos los CSV de la carpeta
func obtenerArchivosCSV() []string {
	files, err := os.ReadDir("datasets")
	if err != nil {
		panic(err)
	}

	var archivos []string

	for _, file := range files {
		name := strings.ToLower(file.Name())

		if strings.HasSuffix(name, ".csv") {
			archivos = append(archivos, "datasets/"+file.Name())
		}
	}

	return archivos
}

// 🔹 Procesar cada archivo
func procesarArchivo(path string) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error abriendo archivo:", path)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error leyendo CSV:", path)
		return
	}

	if len(records) == 0 {
		fmt.Println("Archivo vacío:", path)
		return
	}

	// 🔹 Detectar columna SUMILLA dinámicamente
	headers := records[0]
	indiceSumilla := -1

	for i, h := range headers {
		h = strings.ToLower(strings.TrimSpace(h))

		if strings.Contains(h, "sumilla") {
			indiceSumilla = i
			break
		}
	}

	// 🔹 fallback si no encuentra SUMILLA
	if indiceSumilla == -1 {
		fmt.Println("No se encontró SUMILLA, usando columna 5 por defecto en:", path)
		indiceSumilla = 5
	}

	fmt.Println("===== PROCESAMIENTO =====")

	// 🔹 Mapa para eliminar duplicados (DESPUÉS de limpiar)
	vistos := make(map[string]bool)

	for i, row := range records {
		if i == 0 {
			continue
		}

		if len(row) <= indiceSumilla {
			continue
		}

		sumilla := row[indiceSumilla]

		if sumilla == "" {
			continue
		}

		// 🔹 Pipeline NLP
		limpio := utils.LimpiarTexto(sumilla)

		// 🔹 eliminar duplicados por texto limpio
		if vistos[limpio] {
			continue
		}
		vistos[limpio] = true

		tokens := utils.Tokenizar(limpio)
		filtrado := utils.RemoverStopwords(tokens)
		unicos := utils.RemoverDuplicados(filtrado)

		// 🔹 Mostrar solo ejemplos
		if i <= 3 {
			fmt.Println("Archivo:", path)

			// (opcional) truncar para que la captura sea limpia
			if len(sumilla) > 120 {
				sumilla = sumilla[:120] + "..."
			}

			fmt.Println("Original:")
			fmt.Println("\"" + sumilla + "\"")
			fmt.Println()

			fmt.Print("→ [")
			for j, token := range unicos {
				if j > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("\"%s\"", token)
			}
			fmt.Println("]")

			fmt.Println("Cantidad de tokens:", len(unicos))
			fmt.Println("------")
		}
	}

	fmt.Println("=========================")
}

func main() {
	archivos := obtenerArchivosCSV()
	fmt.Println("Total archivos encontrados:", len(archivos))
	fmt.Println("======================================")

	for _, archivo := range archivos {

		fmt.Println("Procesando:", archivo)

		// 🔹 1. EDA
		utils.AnalizarDataset(archivo)

		// 🔹 2. NLP
		procesarArchivo(archivo)
	}
}
