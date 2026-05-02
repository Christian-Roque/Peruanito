package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"

	"proyecto-nlp/utils"
)

// 🔹 Obtener CSV
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

// 🔹 Procesamiento concurrente
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

	// 🔹 Detectar SUMILLA dinámicamente
	headers := records[0]
	indiceSumilla := -1

	for i, h := range headers {
		h = strings.ToLower(strings.TrimSpace(h))
		if strings.Contains(h, "sumilla") {
			indiceSumilla = i
			break
		}
	}

	if indiceSumilla == -1 {
		indiceSumilla = 5
	}

	fmt.Println("===== PROCESAMIENTO CONCURRENTE =====")

	// 🔹 Canal de trabajos
	jobs := make(chan string, 100)

	// 🔹 Control de concurrencia
	var wg sync.WaitGroup
	var mutex sync.Mutex

	// 🔹 Para evitar duplicados
	vistosTexto := make(map[string]bool)

	// 🔹 Contadores
	totalProcesados := 0
	mostrados := 0

	// 🔹 Número de workers (ajustable)
	numWorkers := 4

	// 🔹 Workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for sumilla := range jobs {

				// 🔹 1. Limpieza
				limpio := utils.LimpiarTexto(sumilla)

				// 🔹 Evitar duplicados (protegido con mutex)
				mutex.Lock()
				if vistosTexto[limpio] {
					mutex.Unlock()
					continue
				}
				vistosTexto[limpio] = true
				mutex.Unlock()

				// 🔹 2. Tokenización
				tokens := utils.Tokenizar(limpio)

				// 🔹 3. Stopwords
				sinStopwords := utils.RemoverStopwords(tokens)

				// 🔹 4. Eliminar duplicados de tokens
				final := utils.RemoverDuplicadosTokens(sinStopwords)

				// 🔹 Actualizar contadores (protegido)
				mutex.Lock()
				totalProcesados++

				// 🔹 Mostrar ejemplos (solo 3)
				if mostrados < 3 {
					fmt.Println("Original:")
					fmt.Println("\"" + sumilla + "\"")
					fmt.Println()

					fmt.Println("→", final)
					fmt.Println("Cantidad tokens:", len(final))
					fmt.Println("------")

					mostrados++
				}
				mutex.Unlock()
			}
		}()
	}

	// 🔹 Enviar trabajos
	for i, row := range records {
		if i == 0 {
			continue
		}

		if len(row) <= indiceSumilla {
			continue
		}

		sumilla := row[indiceSumilla]
		if sumilla != "" {
			jobs <- sumilla
		}
	}

	// 🔹 Cerrar canal y esperar
	close(jobs)
	wg.Wait()

	fmt.Println("Total procesados:", totalProcesados)
	fmt.Println("===============================")
}

func main() {

	archivos := obtenerArchivosCSV()

	fmt.Println("================================")

	// 🔹 EDA GLOBAL
	utils.AnalizarDatasetGlobal(archivos)

	fmt.Println("================================")

	// 🔹 Procesamiento concurrente por archivo
	for _, archivo := range archivos {
		fmt.Println("Procesando:", archivo)
		procesarArchivo(archivo)
	}
}
