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

	fmt.Println("===== BENCHMARK NLP (Workers Scaling) =====")

	// 🔥 Dataset grande
	path := "salida/dataset_para_go.csv"

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error abriendo dataset")
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error leyendo dataset")
		return
	}

	if len(records) == 0 {
		fmt.Println("No hay datos")
		return
	}

	indice := obtenerIndiceSumilla(records[0])

	fmt.Println("Total registros:", len(records))

	runs := 5
	workersList := []int{1, 2, 4, 6, 8, 10, 12}

	// =============================
	// 🔹 BASELINE SECUENCIAL
	// =============================
	fmt.Println("\n===== TIEMPO SECUENCIAL BASE =====")

	var tiemposSec []float64

	for i := 0; i < runs; i++ {
		start := time.Now()
		utils.ProcesarSecuencial(records, indice)
		t := time.Since(start).Seconds()

		fmt.Println("Run", i+1, ":", t)

		tiemposSec = append(tiemposSec, t)
	}

	avgSec := mediaRecortada(tiemposSec)

	fmt.Println("Promedio secuencial (media recortada):", avgSec)

	// =============================
	// 🔹 PRUEBAS CON WORKERS
	// =============================
	fmt.Println("\n===== RESULTADOS POR WORKERS =====")

	fmt.Printf("%-10s %-20s %-20s\n", "Workers", "Tiempo (s)", "Speedup")

	for _, workers := range workersList {

		var tiemposConc []float64

		fmt.Println("\nWorkers:", workers)

		for i := 0; i < runs; i++ {
			start := time.Now()
			utils.ProcesarConcurrente(records, indice, workers)
			t := time.Since(start).Seconds()

			fmt.Println("Run", i+1, ":", t)

			tiemposConc = append(tiemposConc, t)
		}

		avgConc := mediaRecortada(tiemposConc)
		speedup := avgSec / avgConc

		fmt.Printf("\n➡ Workers: %d | Tiempo promedio: %.6f | Speedup: %.4f\n",
			workers, avgConc, speedup)
	}
}
