package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"proyecto-nlp/utils"
)

// 🔹 Obtener CSV
func obtenerArchivosCSV() []string {
	files, _ := os.ReadDir("datasets")

	var archivos []string

	for _, f := range files {
		if strings.HasSuffix(strings.ToLower(f.Name()), ".csv") {
			archivos = append(archivos, "datasets/"+f.Name())
		}
	}

	return archivos
}

// 🔹 Leer todos los CSV y unirlos
func cargarDataset(archivos []string) [][]string {

	var todos [][]string

	for _, path := range archivos {

		file, err := os.Open(path)
		if err != nil {
			continue
		}

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		file.Close()

		if err != nil || len(records) == 0 {
			continue
		}

		todos = append(todos, records...)
	}

	return todos
}

// 🔹 Detectar columna SUMILLA
func obtenerIndiceSumilla(headers []string) int {

	for i, h := range headers {
		h = strings.ToLower(strings.TrimSpace(h))
		if strings.Contains(h, "sumilla") {
			return i
		}
	}

	return 5
}

// 🔹 Media recortada
func mediaRecortada(valores []float64) float64 {

	sort.Float64s(valores)

	if len(valores) > 2 {
		valores = valores[1 : len(valores)-1]
	}

	sum := 0.0
	for _, v := range valores {
		sum += v
	}

	return sum / float64(len(valores))
}

func main() {

	fmt.Println("===== BENCHMARK NLP =====")

	archivos := obtenerArchivosCSV()
	records := cargarDataset(archivos)

	if len(records) == 0 {
		fmt.Println("No hay datos")
		return
	}

	indice := obtenerIndiceSumilla(records[0])

	fmt.Println("Total registros:", len(records))

	runs := 5
	workers := 4

	var tiemposSec []float64
	var tiemposConc []float64

	for i := 0; i < runs; i++ {

		fmt.Println("Ejecución:", i+1)

		// 🔹 Secuencial
		start := time.Now()
		utils.ProcesarSecuencial(records, indice)
		tSec := time.Since(start).Seconds()

		// 🔹 Concurrente
		start = time.Now()
		utils.ProcesarConcurrente(records, indice, workers)
		tConc := time.Since(start).Seconds()

		fmt.Println("Secuencial:", tSec)
		fmt.Println("Concurrente:", tConc)
		fmt.Println("-----------")

		tiemposSec = append(tiemposSec, tSec)
		tiemposConc = append(tiemposConc, tConc)
	}

	avgSec := mediaRecortada(tiemposSec)
	avgConc := mediaRecortada(tiemposConc)

	speedup := avgSec / avgConc

	fmt.Println("===== RESULTADOS =====")
	fmt.Println("Tiempo promedio secuencial:", avgSec)
	fmt.Println("Tiempo promedio concurrente:", avgConc)
	fmt.Println("Speedup:", speedup)
}
