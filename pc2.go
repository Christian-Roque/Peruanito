package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"proyecto-nlp/utils"
)

type runResult struct {
	Prueba             string
	Modo               string
	TipoData           string
	Workers            int
	Run                int
	TotalSeconds       float64
	TotalMS            float64
	ThroughputRegSeg   float64
	Speedup            float64
	Eficiencia         float64
	TotalFilas         int
	FilasValidas       int
	FilasInvalidas     int
	ProcesadosUnicos   int
	DuplicadosGlobales int
	VaciosPostPipeline int
	ResultadoCorrecto  bool
}

type summaryResult struct {
	Prueba              string
	Modo                string
	TipoData            string
	Workers             int
	Runs                int
	MediaRecortadaSeg   float64
	PromedioSeg         float64
	MedianaSeg          float64
	MinSeg              float64
	MaxSeg              float64
	StdSeg              float64
	ThroughputPromedio  float64
	Speedup             float64
	Eficiencia          float64
	ResultadosCorrectos int
}

func obtenerIndiceSumilla(headers []string) int {
	for i, h := range headers {
		h = strings.ToLower(strings.TrimSpace(h))
		if strings.Contains(h, "sumilla") {
			return i
		}
	}
	return 5
}

func parseWorkers(value string) ([]int, error) {
	parts := strings.Split(value, ",")
	workers := make([]int, 0, len(parts))
	seen := make(map[int]bool)

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		w, err := strconv.Atoi(p)
		if err != nil || w < 1 {
			return nil, fmt.Errorf("worker inválido: %q", p)
		}
		if !seen[w] {
			workers = append(workers, w)
			seen[w] = true
		}
	}

	if len(workers) == 0 {
		return nil, errors.New("debes indicar al menos un worker")
	}
	sort.Ints(workers)
	return workers, nil
}

func mediaRecortada(valores []float64, trimPct float64) float64 {
	if len(valores) == 0 {
		return 0
	}
	ordenados := append([]float64(nil), valores...)
	sort.Float64s(ordenados)

	trim := int(float64(len(ordenados)) * trimPct)
	if trim*2 >= len(ordenados) {
		trim = 0
	}
	ordenados = ordenados[trim : len(ordenados)-trim]

	sum := 0.0
	for _, v := range ordenados {
		sum += v
	}
	return sum / float64(len(ordenados))
}

func promedio(valores []float64) float64 {
	if len(valores) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range valores {
		sum += v
	}
	return sum / float64(len(valores))
}

func mediana(valores []float64) float64 {
	if len(valores) == 0 {
		return 0
	}
	ordenados := append([]float64(nil), valores...)
	sort.Float64s(ordenados)
	mid := len(ordenados) / 2
	if len(ordenados)%2 == 0 {
		return (ordenados[mid-1] + ordenados[mid]) / 2
	}
	return ordenados[mid]
}

func std(valores []float64) float64 {
	if len(valores) <= 1 {
		return 0
	}
	avg := promedio(valores)
	var sum float64
	for _, v := range valores {
		diff := v - avg
		sum += diff * diff
	}
	return sqrt(sum / float64(len(valores)-1))
}

// sqrt usa Newton para evitar dependencias adicionales en el benchmark.
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 20; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

func statsIguales(a, b utils.ProcessingStats) bool {
	return a.TotalFilas == b.TotalFilas &&
		a.FilasValidas == b.FilasValidas &&
		a.FilasInvalidas == b.FilasInvalidas &&
		a.ProcesadosUnicos == b.ProcesadosUnicos &&
		a.DuplicadosGlobales == b.DuplicadosGlobales &&
		a.VaciosPostPipeline == b.VaciosPostPipeline
}

func limitarRecords(records [][]string, maxRecords int) [][]string {
	if maxRecords <= 0 || len(records) <= maxRecords+1 {
		return records
	}
	limit := maxRecords + 1 // incluye cabecera
	return records[:limit]
}

func leerCSV(path string, maxRecords int) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errors.New("CSV vacío")
	}
	return limitarRecords(records, maxRecords), nil
}

func asegurarDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

func writeRawCSV(path string, rows []runResult) error {
	if err := asegurarDir(path); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	header := []string{
		"prueba", "modo", "tipo_data", "workers", "run", "total_seconds", "total_ms",
		"throughput_reg_seg", "speedup", "eficiencia", "total_filas", "filas_validas",
		"filas_invalidas", "procesados_unicos", "duplicados_globales", "vacios_post_pipeline",
		"resultado_correcto",
	}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, r := range rows {
		row := []string{
			r.Prueba,
			r.Modo,
			r.TipoData,
			strconv.Itoa(r.Workers),
			strconv.Itoa(r.Run),
			fmt.Sprintf("%.9f", r.TotalSeconds),
			fmt.Sprintf("%.3f", r.TotalMS),
			fmt.Sprintf("%.3f", r.ThroughputRegSeg),
			fmt.Sprintf("%.6f", r.Speedup),
			fmt.Sprintf("%.6f", r.Eficiencia),
			strconv.Itoa(r.TotalFilas),
			strconv.Itoa(r.FilasValidas),
			strconv.Itoa(r.FilasInvalidas),
			strconv.Itoa(r.ProcesadosUnicos),
			strconv.Itoa(r.DuplicadosGlobales),
			strconv.Itoa(r.VaciosPostPipeline),
			strconv.FormatBool(r.ResultadoCorrecto),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func writeSummaryCSV(path string, rows []summaryResult) error {
	if err := asegurarDir(path); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	header := []string{
		"prueba", "modo", "tipo_data", "workers", "runs", "media_recortada_seg",
		"promedio_seg", "mediana_seg", "min_seg", "max_seg", "std_seg",
		"throughput_promedio", "speedup", "eficiencia", "resultados_correctos",
	}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, r := range rows {
		row := []string{
			r.Prueba,
			r.Modo,
			r.TipoData,
			strconv.Itoa(r.Workers),
			strconv.Itoa(r.Runs),
			fmt.Sprintf("%.9f", r.MediaRecortadaSeg),
			fmt.Sprintf("%.9f", r.PromedioSeg),
			fmt.Sprintf("%.9f", r.MedianaSeg),
			fmt.Sprintf("%.9f", r.MinSeg),
			fmt.Sprintf("%.9f", r.MaxSeg),
			fmt.Sprintf("%.9f", r.StdSeg),
			fmt.Sprintf("%.3f", r.ThroughputPromedio),
			fmt.Sprintf("%.6f", r.Speedup),
			fmt.Sprintf("%.6f", r.Eficiencia),
			strconv.Itoa(r.ResultadosCorrectos),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func resumir(rows []runResult, baselineTrimmed float64, trimPct float64) []summaryResult {
	groups := make(map[string][]runResult)
	keys := make([]string, 0)
	for _, r := range rows {
		key := fmt.Sprintf("%s|%s|%s|%d", r.Prueba, r.Modo, r.TipoData, r.Workers)
		if _, ok := groups[key]; !ok {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], r)
	}
	sort.Slice(keys, func(i, j int) bool {
		ai := groups[keys[i]][0]
		aj := groups[keys[j]][0]
		if ai.Modo != aj.Modo {
			return ai.Modo < aj.Modo
		}
		return ai.Workers < aj.Workers
	})

	out := make([]summaryResult, 0, len(keys))
	for _, key := range keys {
		items := groups[key]
		vals := make([]float64, 0, len(items))
		throughput := make([]float64, 0, len(items))
		correctos := 0
		for _, r := range items {
			vals = append(vals, r.TotalSeconds)
			throughput = append(throughput, r.ThroughputRegSeg)
			if r.ResultadoCorrecto {
				correctos++
			}
		}
		sort.Float64s(vals)
		mediaTrim := mediaRecortada(vals, trimPct)
		speedup := 1.0
		eficiencia := 1.0
		if items[0].Modo != "secuencial" && mediaTrim > 0 {
			speedup = baselineTrimmed / mediaTrim
			eficiencia = speedup / float64(items[0].Workers)
		}
		out = append(out, summaryResult{
			Prueba:              items[0].Prueba,
			Modo:                items[0].Modo,
			TipoData:            items[0].TipoData,
			Workers:             items[0].Workers,
			Runs:                len(items),
			MediaRecortadaSeg:   mediaTrim,
			PromedioSeg:         promedio(vals),
			MedianaSeg:          mediana(vals),
			MinSeg:              vals[0],
			MaxSeg:              vals[len(vals)-1],
			StdSeg:              std(vals),
			ThroughputPromedio:  promedio(throughput),
			Speedup:             speedup,
			Eficiencia:          eficiencia,
			ResultadosCorrectos: correctos,
		})
	}
	return out
}

func main() {
	inputPath := flag.String("input", "salida/dataset_para_go.csv", "Ruta del dataset consolidado para benchmark")
	runs := flag.Int("runs", 100, "Número de ejecuciones por configuración")
	workersArg := flag.String("workers", "1,2,4,8,16,32,64,100", "Lista de workers separada por comas")
	outRaw := flag.String("out", "resultados/benchmark_pc2_raw.csv", "CSV de resultados por ejecución")
	outSummary := flag.String("summary", "resultados/benchmark_pc2_resumen.csv", "CSV resumen por configuración")
	tipoData := flag.String("tipo-data", "mixto", "Etiqueta del tipo de data: robotizado, no_robotizado o mixto")
	maxRecords := flag.Int("max-records", 0, "Límite opcional de registros para pruebas rápidas; 0 usa todo el dataset")
	trimPct := flag.Float64("trim-pct", 0.05, "Porcentaje de recorte por cola para media recortada. Ej.: 0.05 recorta 5% inferior y superior")
	flag.Parse()

	if *runs < 1 {
		fmt.Fprintln(os.Stderr, "ERROR: runs debe ser mayor o igual a 1")
		os.Exit(1)
	}
	workersList, err := parseWorkers(*workersArg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	fmt.Println("===== BENCHMARK PC2: NLP LEGAL EL PERUANO =====")
	fmt.Println("Entrada:", *inputPath)
	fmt.Println("Runs por configuración:", *runs)
	fmt.Println("Workers:", workersList)
	fmt.Println("Tipo de data:", *tipoData)
	fmt.Println("CPU lógica detectada:", runtime.NumCPU())

	records, err := leerCSV(*inputPath, *maxRecords)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR leyendo dataset:", err)
		fmt.Fprintln(os.Stderr, "Sugerencia: primero genera el dataset con: go run generar_dataset_sintetico_elperuano_enriquecido.go --input-dir ./datasets --output-dir ./salida --target-total 1000000")
		os.Exit(1)
	}
	indice := obtenerIndiceSumilla(records[0])
	fmt.Println("Registros cargados, incluyendo cabecera:", len(records))
	fmt.Println("Índice SUMILLA:", indice)

	allRows := make([]runResult, 0, (*runs)*(len(workersList)+1))
	var baselineTimes []float64
	var baselineStats utils.ProcessingStats

	fmt.Println("\n===== BASELINE SECUENCIAL =====")
	for i := 1; i <= *runs; i++ {
		start := time.Now()
		stats := utils.ProcesarSecuencialStats(records, indice)
		totalSeconds := time.Since(start).Seconds()
		if i == 1 {
			baselineStats = stats
		}
		correcto := statsIguales(baselineStats, stats)
		throughput := 0.0
		if totalSeconds > 0 {
			throughput = float64(stats.FilasValidas) / totalSeconds
		}

		baselineTimes = append(baselineTimes, totalSeconds)
		allRows = append(allRows, runResult{
			Prueba:             "pc2_benchmark_nlp_legal",
			Modo:               "secuencial",
			TipoData:           *tipoData,
			Workers:            1,
			Run:                i,
			TotalSeconds:       totalSeconds,
			TotalMS:            totalSeconds * 1000,
			ThroughputRegSeg:   throughput,
			Speedup:            1,
			Eficiencia:         1,
			TotalFilas:         stats.TotalFilas,
			FilasValidas:       stats.FilasValidas,
			FilasInvalidas:     stats.FilasInvalidas,
			ProcesadosUnicos:   stats.ProcesadosUnicos,
			DuplicadosGlobales: stats.DuplicadosGlobales,
			VaciosPostPipeline: stats.VaciosPostPipeline,
			ResultadoCorrecto:  correcto,
		})
		fmt.Printf("Secuencial run %03d/%03d -> %.6fs | únicos=%d | correcto=%v\n", i, *runs, totalSeconds, stats.ProcesadosUnicos, correcto)
	}
	baselineTrimmed := mediaRecortada(baselineTimes, *trimPct)
	fmt.Printf("Media recortada secuencial: %.6fs\n", baselineTrimmed)

	fmt.Println("\n===== WORKER POOL CONCURRENTE =====")
	for _, workers := range workersList {
		fmt.Printf("\nWorkers: %d\n", workers)
		for i := 1; i <= *runs; i++ {
			start := time.Now()
			stats := utils.ProcesarConcurrenteStats(records, indice, workers)
			totalSeconds := time.Since(start).Seconds()
			correcto := statsIguales(baselineStats, stats)
			throughput := 0.0
			if totalSeconds > 0 {
				throughput = float64(stats.FilasValidas) / totalSeconds
			}
			speedup := 0.0
			eficiencia := 0.0
			if totalSeconds > 0 {
				speedup = baselineTrimmed / totalSeconds
				eficiencia = speedup / float64(workers)
			}

			allRows = append(allRows, runResult{
				Prueba:             "pc2_benchmark_nlp_legal",
				Modo:               "concurrente_worker_pool_mutex",
				TipoData:           *tipoData,
				Workers:            workers,
				Run:                i,
				TotalSeconds:       totalSeconds,
				TotalMS:            totalSeconds * 1000,
				ThroughputRegSeg:   throughput,
				Speedup:            speedup,
				Eficiencia:         eficiencia,
				TotalFilas:         stats.TotalFilas,
				FilasValidas:       stats.FilasValidas,
				FilasInvalidas:     stats.FilasInvalidas,
				ProcesadosUnicos:   stats.ProcesadosUnicos,
				DuplicadosGlobales: stats.DuplicadosGlobales,
				VaciosPostPipeline: stats.VaciosPostPipeline,
				ResultadoCorrecto:  correcto,
			})
			fmt.Printf("Concurrente workers=%d run %03d/%03d -> %.6fs | speedup=%.4f | correcto=%v\n", workers, i, *runs, totalSeconds, speedup, correcto)
		}
	}

	summary := resumir(allRows, baselineTrimmed, *trimPct)
	if err := writeRawCSV(*outRaw, allRows); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR escribiendo CSV raw:", err)
		os.Exit(1)
	}
	if err := writeSummaryCSV(*outSummary, summary); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR escribiendo CSV resumen:", err)
		os.Exit(1)
	}

	fmt.Println("\n===== ARCHIVOS GENERADOS =====")
	fmt.Println("Raw:", *outRaw)
	fmt.Println("Resumen:", *outSummary)
	fmt.Println("\nUsa Python para graficar:")
	fmt.Printf("python scripts/graficos_benchmark.py --raw %s --summary %s --out-dir graficos\n", *outRaw, *outSummary)
}
