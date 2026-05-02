package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var baseCols = []string{
	"FECHA_PUBLICACION",
	"OP",
	"ENTIDAD",
	"DISPOSITIVO",
	"NUMERO",
	"SUMILLA",
	"LINK",
	"FECHA_CORTE",
}

var auditCols = []string{
	"FECHA_PUBLICACION",
	"OP",
	"ENTIDAD",
	"DISPOSITIVO",
	"NUMERO",
	"SUMILLA",
	"LINK",
	"FECHA_CORTE",
	"ARCHIVO_ORIGEN",
	"ORIGEN_REGISTRO",
	"TIPO_SINTETICO",
	"REGLA_GENERACION",
	"MES_DISTRIBUCION",
	"OP_ORIGEN",
}

var headerMap = map[string]string{
	"fecha publicaci¢n": "FECHA_PUBLICACION",
	"fecha publicación": "FECHA_PUBLICACION",
	"fecha publicacion": "FECHA_PUBLICACION",
	"fecha_publicacion": "FECHA_PUBLICACION",
	"fecha corte":       "FECHA_CORTE",
	"fecha_corte":       "FECHA_CORTE",
	"op":                "OP",
	"entidad":           "ENTIDAD",
	"dispositivo":       "DISPOSITIVO",
	"n£mero":            "NUMERO",
	"número":            "NUMERO",
	"numero":            "NUMERO",
	"sumilla":           "SUMILLA",
	"link":              "LINK",
}

var multiSpace = regexp.MustCompile(`\s+`)

var accentReplacer = strings.NewReplacer(
	"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n",
	"Á", "A", "É", "E", "Í", "I", "Ó", "O", "Ú", "U", "Ñ", "N",
)
var corruptReplacer = strings.NewReplacer(
	"ó", "¢", "Ó", "¢", "ú", "£", "Ú", "£", "ñ", "¤", "Ñ", "¤",
)

var abbrevReplacer = strings.NewReplacer(
	"Resolución", "Res.", "RESOLUCIÓN", "RES.",
	"Ministerial", "Min.", "MINISTERIAL", "MIN.",
	"Municipalidad", "Muni.", "MUNICIPALIDAD", "MUNI.",
	"Administración", "Adm.", "ADMINISTRACIÓN", "ADM.",
	"Dirección", "Dir.", "DIRECCIÓN", "DIR.",
)

type Record struct {
	Fields        [8]string
	ArchivoOrigen string
	MonthKey      string
	Fecha         time.Time
	HasDate       bool
}

type ConsolidationResult struct {
	Records          []Record
	CountsByFile     map[string]int
	TotalBeforeDedup int
	Duplicates       int
	MonthCounts      map[string]int
}

type Config struct {
	InputDir                  string
	OutputDir                 string
	TargetTotal               int
	NSyntheticFlag            int
	Seed                      int64
	ChunkSize                 int
	SkipAudit                 bool
	TextVariantRate           float64
	NoiseRate                 float64
	AnomalyRate               float64
	PreserveMonthDistribution bool
	RandomizeDateWithinMonth  bool
	MarkSumilla               bool
}

type GenerationPlan struct {
	Buckets []BucketPlan
	Total   int
}

type BucketPlan struct {
	Key     string
	Records []Record
	Quota   int
}

type GenerationStats struct {
	Total          int
	Real           int
	Synthetic      int
	Remuestreo     int
	VariacionTexto int
	RuidoTexto     int
	Anomalia       int
	FechaGenerada  int
	ByMonth        map[string]int
	ByRule         map[string]int
}

type SyntheticResult struct {
	Fields [8]string
	Tipo   string
	Regla  string
	Month  string
}

func main() {
	cfg := parseFlags()

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		exitWithError(fmt.Errorf("no se pudo crear output-dir: %w", err))
	}

	result, err := consolidate(cfg.InputDir)
	if err != nil {
		exitWithError(err)
	}

	nReal := len(result.Records)
	nSynthetic := 0
	if cfg.NSyntheticFlag >= 0 {
		nSynthetic = cfg.NSyntheticFlag
	} else {
		nSynthetic = cfg.TargetTotal - nReal
		if nSynthetic < 0 {
			nSynthetic = 0
		}
	}

	fmt.Printf("\nRegistros reales consolidados: %s\n", formatInt(nReal))
	fmt.Printf("Registros sintéticos a generar: %s\n", formatInt(nSynthetic))
	fmt.Printf("Total esperado: %s\n", formatInt(nReal+nSynthetic))
	fmt.Printf("Distribución mensual preservada: %v\n", cfg.PreserveMonthDistribution)
	fmt.Printf("Variación texto: %.2f%% | Ruido: %.2f%% | Anomalías: %.2f%%\n\n",
		cfg.TextVariantRate*100, cfg.NoiseRate*100, cfg.AnomalyRate*100)

	stats, err := writeOutputsStreaming(result, cfg, nSynthetic)
	if err != nil {
		exitWithError(err)
	}

	if err := buildReports(result, stats, cfg.OutputDir); err != nil {
		exitWithError(err)
	}

	fmt.Println("\nArchivos generados:")
	fmt.Printf("- %s\n", filepath.Join(cfg.OutputDir, "dataset_para_go.csv"))
	if !cfg.SkipAudit {
		fmt.Printf("- %s\n", filepath.Join(cfg.OutputDir, "dataset_con_auditoria.csv"))
	}
	fmt.Printf("- %s\n", filepath.Join(cfg.OutputDir, "reporte_generacion_sintetica.csv"))
	fmt.Printf("- %s\n", filepath.Join(cfg.OutputDir, "resumen_por_origen.csv"))
	fmt.Printf("- %s\n", filepath.Join(cfg.OutputDir, "resumen_archivos_reales.csv"))
	fmt.Printf("- %s\n", filepath.Join(cfg.OutputDir, "resumen_sintetico_por_tipo.csv"))
	fmt.Printf("- %s\n", filepath.Join(cfg.OutputDir, "resumen_sintetico_por_mes.csv"))
	fmt.Printf("- %s\n", filepath.Join(cfg.OutputDir, "resumen_reglas_generacion.csv"))
	fmt.Printf("\nTotal final esperado: %s registros\n", formatInt(nReal+nSynthetic))
}

func parseFlags() Config {
	inputDir := flag.String("input-dir", "", "Carpeta con archivos CSV de entrada")
	outputDir := flag.String("output-dir", "salida_enriquecida", "Carpeta donde se escribirán los CSV de salida")
	targetTotal := flag.Int("target-total", 1000000, "Total final deseado: reales + sintéticos")
	nSyntheticFlag := flag.Int("n-synthetic", -1, "Cantidad exacta de registros sintéticos. Si se omite, se calcula con target-total")
	seed := flag.Int64("seed", 2026, "Semilla para reproducibilidad")
	chunkSize := flag.Int("chunk-size", 100000, "Cada cuántos registros sintéticos mostrar progreso")
	skipAudit := flag.Bool("skip-audit", false, "Si se activa, no genera dataset_con_auditoria.csv")
	variantRate := flag.Float64("text-variant-rate", 0.35, "Proporción de sintéticos con variación controlada de SUMILLA")
	noiseRate := flag.Float64("noise-rate", 0.03, "Proporción de sintéticos con ruido textual controlado")
	anomalyRate := flag.Float64("anomaly-rate", 0.02, "Proporción de sintéticos con anomalías controladas")
	preserveMonthDistribution := flag.Bool("preserve-month-distribution", true, "Si se activa, genera sintéticos respetando la distribución por año-mes")
	randomizeDateWithinMonth := flag.Bool("randomize-date-within-month", true, "Si se activa, genera fechas aleatorias dentro del mismo mes del registro base")
	markSumilla := flag.Bool("mark-sumilla", false, "Si se activa, añade marca textual en SUMILLA sintética. Por defecto no se recomienda")
	noMarkSumilla := flag.Bool("no-mark-sumilla", false, "Compatibilidad: si se activa, fuerza a no marcar SUMILLA")
	flag.Parse()

	if strings.TrimSpace(*inputDir) == "" {
		exitWithError(errors.New("debes indicar --input-dir ./datasets"))
	}
	if *targetTotal < 0 {
		exitWithError(errors.New("--target-total no puede ser negativo"))
	}
	if *nSyntheticFlag < -1 {
		exitWithError(errors.New("--n-synthetic no puede ser menor que -1"))
	}
	if *chunkSize <= 0 {
		exitWithError(errors.New("--chunk-size debe ser mayor que 0"))
	}
	if !validRate(*variantRate) || !validRate(*noiseRate) || !validRate(*anomalyRate) {
		exitWithError(errors.New("las tasas deben estar entre 0 y 1"))
	}
	if *variantRate+*noiseRate+*anomalyRate > 1.0 {
		exitWithError(errors.New("text-variant-rate + noise-rate + anomaly-rate no puede superar 1"))
	}

	finalMarkSumilla := *markSumilla
	if *noMarkSumilla {
		finalMarkSumilla = false
	}

	return Config{
		InputDir:                  *inputDir,
		OutputDir:                 *outputDir,
		TargetTotal:               *targetTotal,
		NSyntheticFlag:            *nSyntheticFlag,
		Seed:                      *seed,
		ChunkSize:                 *chunkSize,
		SkipAudit:                 *skipAudit,
		TextVariantRate:           *variantRate,
		NoiseRate:                 *noiseRate,
		AnomalyRate:               *anomalyRate,
		PreserveMonthDistribution: *preserveMonthDistribution,
		RandomizeDateWithinMonth:  *randomizeDateWithinMonth,
		MarkSumilla:               finalMarkSumilla,
	}
}

func validRate(v float64) bool {
	return v >= 0 && v <= 1
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	os.Exit(1)
}

func consolidate(inputDir string) (ConsolidationResult, error) {
	files, err := listCSVFiles(inputDir)
	if err != nil {
		return ConsolidationResult{}, err
	}
	if len(files) == 0 {
		return ConsolidationResult{}, fmt.Errorf("no se encontraron CSV en %s", inputDir)
	}

	seen := make(map[string]struct{})
	records := make([]Record, 0, 300000)
	countsByFile := make(map[string]int)
	monthCounts := make(map[string]int)
	totalBeforeDedup := 0
	duplicates := 0

	for _, path := range files {
		recs, err := readOneCSV(path)
		if err != nil {
			return ConsolidationResult{}, err
		}
		totalBeforeDedup += len(recs)
		keptThisFile := 0

		for _, rec := range recs {
			key := recordKey(rec.Fields)
			if _, ok := seen[key]; ok {
				duplicates++
				continue
			}
			seen[key] = struct{}{}
			records = append(records, rec)
			keptThisFile++
			monthCounts[rec.MonthKey]++
		}

		countsByFile[filepath.Base(path)] = keptThisFile
		fmt.Printf("Leído: %-30s -> %s registros; conservados únicos: %s\n",
			filepath.Base(path), formatInt(len(recs)), formatInt(keptThisFile))
	}

	if duplicates > 0 {
		fmt.Printf("Duplicados exactos eliminados: %s\n", formatInt(duplicates))
	}

	return ConsolidationResult{
		Records:          records,
		CountsByFile:     countsByFile,
		TotalBeforeDedup: totalBeforeDedup,
		Duplicates:       duplicates,
		MonthCounts:      monthCounts,
	}, nil
}

func listCSVFiles(inputDir string) ([]string, error) {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer la carpeta %s: %w", inputDir, err)
	}

	files := make([]string, 0)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".csv" {
			files = append(files, filepath.Join(inputDir, name))
		}
	}
	sort.Strings(files)
	return files, nil
}

func readOneCSV(path string) ([]Record, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer %s: %w", filepath.Base(path), err)
	}
	text := decodeToUTF8(raw)
	text = strings.TrimPrefix(text, "\ufeff")

	delimiter := detectDelimiter(text)
	reader := csv.NewReader(strings.NewReader(text))
	reader.Comma = delimiter
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.ReuseRecord = false

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer encabezado de %s: %w", filepath.Base(path), err)
	}

	baseIndex, cleanedHeaders, err := buildBaseIndex(header)
	if err != nil {
		return nil, fmt.Errorf("%s: %w. Columnas detectadas: %v", filepath.Base(path), err, cleanedHeaders)
	}

	out := make([]Record, 0)
	lineNumber := 1
	for {
		row, err := reader.Read()
		lineNumber++
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Advertencia: línea omitida en %s, línea %d: %v\n", filepath.Base(path), lineNumber, err)
			continue
		}

		var fields [8]string
		for i, colName := range baseCols {
			idx := baseIndex[colName]
			if idx >= 0 && idx < len(row) {
				fields[i] = strings.TrimSpace(row[idx])
			} else {
				fields[i] = ""
			}
		}

		if allEmpty(fields) {
			continue
		}

		fecha, hasDate := parseDate(fields[0])
		monthKey := "SIN_FECHA"
		if hasDate {
			monthKey = fecha.Format("2006-01")
		}

		out = append(out, Record{
			Fields:        fields,
			ArchivoOrigen: filepath.Base(path),
			MonthKey:      monthKey,
			Fecha:         fecha,
			HasDate:       hasDate,
		})
	}

	return out, nil
}

func decodeToUTF8(raw []byte) string {
	if utf8.Valid(raw) {
		return string(raw)
	}

	var b strings.Builder
	b.Grow(len(raw))
	for _, c := range raw {
		if c < 0x80 {
			b.WriteByte(c)
			continue
		}

		r, ok := cp1252Rune(c)
		if ok {
			b.WriteRune(r)
		} else {
			b.WriteRune(rune(c))
		}
	}
	return b.String()
}

func cp1252Rune(c byte) (rune, bool) {
	mapping := map[byte]rune{
		0x80: '€', 0x82: '‚', 0x83: 'ƒ', 0x84: '„', 0x85: '…', 0x86: '†', 0x87: '‡',
		0x88: 'ˆ', 0x89: '‰', 0x8A: 'Š', 0x8B: '‹', 0x8C: 'Œ', 0x8E: 'Ž',
		0x91: '‘', 0x92: '’', 0x93: '“', 0x94: '”', 0x95: '•', 0x96: '–', 0x97: '—',
		0x98: '˜', 0x99: '™', 0x9A: 'š', 0x9B: '›', 0x9C: 'œ', 0x9E: 'ž', 0x9F: 'Ÿ',
	}
	r, ok := mapping[c]
	return r, ok
}

func detectDelimiter(text string) rune {
	firstLine := text
	if idx := strings.IndexAny(text, "\r\n"); idx >= 0 {
		firstLine = text[:idx]
	}

	candidates := []rune{',', ';', '\t'}
	best := ','
	bestCount := -1
	for _, cand := range candidates {
		count := strings.Count(firstLine, string(cand))
		if count > bestCount {
			best = cand
			bestCount = count
		}
	}
	return best
}

func buildBaseIndex(header []string) (map[string]int, []string, error) {
	baseIndex := make(map[string]int)
	cleaned := make([]string, 0, len(header))

	for idx, h := range header {
		clean := cleanHeader(h)
		if clean == "" || strings.HasPrefix(strings.ToLower(clean), "unnamed") {
			continue
		}
		cleaned = append(cleaned, clean)
		if _, exists := baseIndex[clean]; !exists {
			baseIndex[clean] = idx
		}
	}

	missing := make([]string, 0)
	for _, col := range baseCols {
		if _, ok := baseIndex[col]; !ok {
			missing = append(missing, col)
		}
	}
	if len(missing) > 0 {
		return baseIndex, cleaned, fmt.Errorf("faltan columnas esperadas: %v", missing)
	}
	return baseIndex, cleaned, nil
}

func cleanHeader(col string) string {
	raw := strings.TrimSpace(strings.TrimPrefix(col, "\ufeff"))
	raw = strings.ReplaceAll(raw, "\u00a0", " ")
	key := strings.ToLower(raw)
	key = multiSpace.ReplaceAllString(key, " ")
	key = strings.TrimSpace(key)

	if mapped, ok := headerMap[key]; ok {
		return mapped
	}

	key2 := strings.ReplaceAll(key, "¢", "o")
	key2 = strings.ReplaceAll(key2, "£", "u")
	key2 = strings.ReplaceAll(key2, "á", "a")
	key2 = strings.ReplaceAll(key2, "é", "e")
	key2 = strings.ReplaceAll(key2, "í", "i")
	key2 = strings.ReplaceAll(key2, "ó", "o")
	key2 = strings.ReplaceAll(key2, "ú", "u")
	key2 = strings.ReplaceAll(key2, "ñ", "n")
	if mapped, ok := headerMap[key2]; ok {
		return mapped
	}

	clean := strings.ToUpper(raw)
	clean = multiSpace.ReplaceAllString(clean, "_")
	clean = strings.ReplaceAll(clean, " ", "_")
	return clean
}

func recordKey(fields [8]string) string {
	return strings.Join(fields[:], "\x1f")
}

func allEmpty(fields [8]string) bool {
	for _, v := range fields {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}

func writeOutputsStreaming(result ConsolidationResult, cfg Config, nSynthetic int) (GenerationStats, error) {
	if len(result.Records) == 0 {
		return GenerationStats{}, errors.New("no hay registros reales para muestrear")
	}

	start := time.Now()
	rng := rand.New(rand.NewSource(cfg.Seed))
	plan := buildGenerationPlan(result.Records, nSynthetic, cfg.PreserveMonthDistribution)

	goPath := filepath.Join(cfg.OutputDir, "dataset_para_go.csv")
	goFile, err := createCSVWithBOM(goPath)
	if err != nil {
		return GenerationStats{}, err
	}
	defer goFile.Close()
	goWriter := csv.NewWriter(goFile)
	defer goWriter.Flush()

	if err := goWriter.Write(baseCols); err != nil {
		return GenerationStats{}, fmt.Errorf("error escribiendo encabezado dataset_para_go: %w", err)
	}

	var auditFile *os.File
	var auditWriter *csv.Writer
	if !cfg.SkipAudit {
		auditPath := filepath.Join(cfg.OutputDir, "dataset_con_auditoria.csv")
		auditFile, err = createCSVWithBOM(auditPath)
		if err != nil {
			return GenerationStats{}, err
		}
		defer auditFile.Close()
		auditWriter = csv.NewWriter(auditFile)
		defer auditWriter.Flush()
		if err := auditWriter.Write(auditCols); err != nil {
			return GenerationStats{}, fmt.Errorf("error escribiendo encabezado dataset_con_auditoria: %w", err)
		}
	}

	stats := GenerationStats{
		Real:      len(result.Records),
		Synthetic: nSynthetic,
		Total:     len(result.Records) + nSynthetic,
		ByMonth:   make(map[string]int),
		ByRule:    make(map[string]int),
	}

	for _, rec := range result.Records {
		if err := goWriter.Write(rec.Fields[:]); err != nil {
			return stats, fmt.Errorf("error escribiendo registro real: %w", err)
		}
		if auditWriter != nil {
			row := auditRow(rec.Fields, rec.ArchivoOrigen, "REAL", "REAL", "REGISTRO_ORIGINAL", rec.MonthKey, rec.Fields[1])
			if err := auditWriter.Write(row); err != nil {
				return stats, fmt.Errorf("error escribiendo auditoría real: %w", err)
			}
		}
	}

	seq := 0
	for _, bucket := range plan.Buckets {
		if bucket.Quota <= 0 || len(bucket.Records) == 0 {
			continue
		}
		for i := 0; i < bucket.Quota; i++ {
			seq++
			base := bucket.Records[rng.Intn(len(bucket.Records))]
			syn := makeSyntheticRecord(base, seq, rng, cfg)

			if err := goWriter.Write(syn.Fields[:]); err != nil {
				return stats, fmt.Errorf("error escribiendo registro sintético: %w", err)
			}
			if auditWriter != nil {
				row := auditRow(syn.Fields, "SINTETICO_GENERADO", "SINTETICO", syn.Tipo, syn.Regla, syn.Month, base.Fields[1])
				if err := auditWriter.Write(row); err != nil {
					return stats, fmt.Errorf("error escribiendo auditoría sintética: %w", err)
				}
			}

			stats.ByMonth[syn.Month]++
			stats.ByRule[syn.Regla]++
			switch syn.Tipo {
			case "REMUESTREO_CONTROLADO":
				stats.Remuestreo++
			case "VARIACION_TEXTO":
				stats.VariacionTexto++
			case "RUIDO_TEXTO":
				stats.RuidoTexto++
			case "ANOMALIA_CONTROLADA":
				stats.Anomalia++
			}
			if strings.TrimSpace(syn.Fields[0]) != strings.TrimSpace(base.Fields[0]) {
				stats.FechaGenerada++
			}

			if seq%cfg.ChunkSize == 0 || seq == nSynthetic {
				goWriter.Flush()
				if err := goWriter.Error(); err != nil {
					return stats, err
				}
				if auditWriter != nil {
					auditWriter.Flush()
					if err := auditWriter.Error(); err != nil {
						return stats, err
					}
				}
				fmt.Printf("Sintéticos escritos: %s/%s\n", formatInt(seq), formatInt(nSynthetic))
			}
		}
	}

	goWriter.Flush()
	if err := goWriter.Error(); err != nil {
		return stats, err
	}
	if auditWriter != nil {
		auditWriter.Flush()
		if err := auditWriter.Error(); err != nil {
			return stats, err
		}
	}

	fmt.Printf("Escritura terminada en %.1f segundos\n", time.Since(start).Seconds())
	return stats, nil
}

func buildGenerationPlan(records []Record, nSynthetic int, preserveMonth bool) GenerationPlan {
	if nSynthetic <= 0 {
		return GenerationPlan{Buckets: []BucketPlan{}, Total: 0}
	}
	if !preserveMonth {
		return GenerationPlan{Buckets: []BucketPlan{{Key: "GLOBAL", Records: records, Quota: nSynthetic}}, Total: nSynthetic}
	}

	byMonth := make(map[string][]Record)
	for _, rec := range records {
		byMonth[rec.MonthKey] = append(byMonth[rec.MonthKey], rec)
	}

	keys := make([]string, 0, len(byMonth))
	for key := range byMonth {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	totalReal := len(records)
	quotas := make(map[string]int)
	remainders := make([]monthRemainder, 0, len(keys))
	assigned := 0

	for _, key := range keys {
		exact := float64(len(byMonth[key])) / float64(totalReal) * float64(nSynthetic)
		floorQuota := int(math.Floor(exact))
		quotas[key] = floorQuota
		assigned += floorQuota
		remainders = append(remainders, monthRemainder{Key: key, Remainder: exact - float64(floorQuota), Count: len(byMonth[key])})
	}

	sort.Slice(remainders, func(i, j int) bool {
		if remainders[i].Remainder == remainders[j].Remainder {
			return remainders[i].Count > remainders[j].Count
		}
		return remainders[i].Remainder > remainders[j].Remainder
	})

	left := nSynthetic - assigned
	for i := 0; i < left; i++ {
		quotas[remainders[i%len(remainders)].Key]++
	}

	buckets := make([]BucketPlan, 0, len(keys))
	for _, key := range keys {
		buckets = append(buckets, BucketPlan{Key: key, Records: byMonth[key], Quota: quotas[key]})
	}
	return GenerationPlan{Buckets: buckets, Total: nSynthetic}
}

type monthRemainder struct {
	Key       string
	Remainder float64
	Count     int
}

func makeSyntheticRecord(base Record, seq int, rng *rand.Rand, cfg Config) SyntheticResult {
	fields := base.Fields
	seqStr := fmt.Sprintf("%09d", seq)
	fields[1] = "SYN-" + seqStr

	numeroBase := strings.TrimSpace(fields[4])
	if numeroBase != "" {
		fields[4] = numeroBase + "-SYN" + seqStr
	} else {
		fields[4] = "SYN" + seqStr
	}
	fields[6] = "synthetic://elperuano/SYN-" + seqStr

	if cfg.RandomizeDateWithinMonth && base.HasDate {
		newDate := randomDateWithinMonth(base.Fecha, rng)
		fields[0] = formatDateLikeOriginal(fields[0], newDate)
	}

	tipo := "REMUESTREO_CONTROLADO"
	regla := "REMUESTREO_CON_ID_UNICO"

	u := rng.Float64()
	if u < cfg.AnomalyRate {
		tipo = "ANOMALIA_CONTROLADA"
		fields, regla = applyAnomaly(fields, seqStr, rng)
	} else if u < cfg.AnomalyRate+cfg.NoiseRate {
		tipo = "RUIDO_TEXTO"
		fields[5], regla = addTextNoise(fields[5], rng)
	} else if u < cfg.AnomalyRate+cfg.NoiseRate+cfg.TextVariantRate {
		tipo = "VARIACION_TEXTO"
		fields[5], regla = varySumilla(fields[5], fields[2], fields[3], rng)
	}

	if cfg.MarkSumilla && strings.TrimSpace(fields[5]) != "" {
		fields[5] = fields[5] + " [registro sintetico]"
	}

	month := base.MonthKey
	if d, ok := parseDate(fields[0]); ok {
		month = d.Format("2006-01")
	}

	return SyntheticResult{Fields: fields, Tipo: tipo, Regla: regla, Month: month}
}

func randomDateWithinMonth(base time.Time, rng *rand.Rand) time.Time {
	year, month, _ := base.Date()
	loc := time.UTC
	firstNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, loc)
	lastDay := firstNextMonth.AddDate(0, 0, -1).Day()
	day := rng.Intn(lastDay) + 1
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

func formatDateLikeOriginal(original string, d time.Time) string {
	s := strings.TrimSpace(original)
	if len(s) == 8 && onlyDigits(s) {
		return d.Format("20060102")
	}
	if strings.Contains(s, "-") && len(s) >= 10 && strings.Index(s, "-") == 4 {
		return d.Format("2006-01-02")
	}
	if strings.Contains(s, "-") {
		return d.Format("02-01-2006")
	}
	return d.Format("02/01/2006")
}

func varySumilla(sumilla string, entidad string, dispositivo string, rng *rand.Rand) (string, string) {
	s := cleanTextValue(sumilla)
	if s == "" {
		return fallbackSumilla(entidad, dispositivo, rng), "VARIACION_DESDE_CAMPOS_CATEGORICOS"
	}

	lower := strings.ToLower(removeAccents(s))
	templates := make([]string, 0)

	if strings.Contains(lower, "designan") || strings.Contains(lower, "designa") {
		templates = append(templates,
			replaceFirstInsensitive(s, "Designan", "Designan temporalmente"),
			replaceFirstInsensitive(s, "Designan", "Encargan funciones a"),
			replaceFirstInsensitive(s, "Designan", "Nombran"),
		)
	}
	if strings.Contains(lower, "aprueban") || strings.Contains(lower, "aprueba") {
		templates = append(templates,
			replaceFirstInsensitive(s, "Aprueban", "Aprueban la actualización de"),
			replaceFirstInsensitive(s, "Aprueban", "Autorizan y aprueban"),
			replaceFirstInsensitive(s, "Aprueba", "Aprueba la modificación de"),
		)
	}
	if strings.Contains(lower, "autorizan") || strings.Contains(lower, "autoriza") {
		templates = append(templates,
			replaceFirstInsensitive(s, "Autorizan", "Autorizan de manera excepcional"),
			replaceFirstInsensitive(s, "Autorizan", "Disponen autorización para"),
			replaceFirstInsensitive(s, "Autoriza", "Autoriza temporalmente"),
		)
	}
	if strings.Contains(lower, "resolucion") || strings.Contains(lower, "resolución") {
		templates = append(templates,
			s+" para fines administrativos",
			"Modifican alcances de "+lowerFirst(s),
		)
	}
	if strings.Contains(lower, "decreto") {
		templates = append(templates,
			s+" en el marco de la normativa vigente",
			"Actualizan disposiciones relacionadas con "+lowerFirst(s),
		)
	}
	if strings.Contains(lower, "renuncia") || strings.Contains(lower, "aceptan") {
		templates = append(templates,
			s+" y encargan funciones temporalmente",
			"Formalizan "+lowerFirst(s),
		)
	}

	if len(templates) == 0 {
		prefixes := []string{
			"Actualizan disposición referida a",
			"Precisan alcances de",
			"Modifican parcialmente",
			"Formalizan procedimiento vinculado a",
			"Disponen acciones administrativas sobre",
		}
		return prefixes[rng.Intn(len(prefixes))] + " " + lowerFirst(s), "VARIACION_GENERICA_SUMILLA"
	}

	candidate := cleanTextValue(templates[rng.Intn(len(templates))])
	candidate = limitWords(candidate, 55)
	return candidate, "VARIACION_SEMANTICA_CONTROLADA"
}

func fallbackSumilla(entidad string, dispositivo string, rng *rand.Rand) string {
	ent := cleanTextValue(entidad)
	disp := cleanTextValue(dispositivo)
	if ent == "" {
		ent = "entidad pública"
	}
	if disp == "" {
		disp = "dispositivo legal"
	}
	templates := []string{
		"Publican " + lowerFirst(disp) + " emitido por " + ent,
		"Formalizan acto administrativo de " + ent,
		"Aprueban disposición institucional de " + ent,
		"Comunican actualización normativa de " + ent,
	}
	return templates[rng.Intn(len(templates))]
}

func addTextNoise(sumilla string, rng *rand.Rand) (string, string) {
	s := cleanTextValue(sumilla)
	if s == "" {
		s = "Resolución administrativa publicada por entidad pública"
	}

	switch rng.Intn(6) {
	case 0:
		return strings.ToUpper(s), "RUIDO_MAYUSCULAS"
	case 1:
		return introduceDoubleSpaces(s, rng), "RUIDO_ESPACIOS_DOBLES"
	case 2:
		return removeSomeAccents(s), "RUIDO_SIN_TILDES"
	case 3:
		return corruptCommonChars(s), "RUIDO_CODIFICACION_SIMULADA"
	case 4:
		return s + "  ", "RUIDO_ESPACIO_FINAL"
	default:
		return abbreviateCommonWords(s), "RUIDO_ABREVIATURAS"
	}
}

func applyAnomaly(fields [8]string, seqStr string, rng *rand.Rand) ([8]string, string) {
	switch rng.Intn(6) {
	case 0:
		fields[5] = "PUBLICACION REPETIDA PUBLICACION REPETIDA PUBLICACION REPETIDA"
		return fields, "ANOMALIA_SUMILLA_DUPLICADA"
	case 1:
		fields[6] = ""
		return fields, "ANOMALIA_LINK_VACIO"
	case 2:
		fields[4] = ""
		return fields, "ANOMALIA_NUMERO_VACIO"
	case 3:
		fields[5] = "S/D"
		return fields, "ANOMALIA_SUMILLA_MUY_CORTA"
	case 4:
		fields[5] = makeLongText(fields[5])
		return fields, "ANOMALIA_TEXTO_LARGO"
	default:
		fields[0] = "99/99/9999"
		fields[6] = "synthetic://elperuano/anomalia-fecha/SYN-" + seqStr
		return fields, "ANOMALIA_FECHA_INVALIDA"
	}
}

func makeLongText(base string) string {
	s := cleanTextValue(base)
	if s == "" {
		s = "Resolución administrativa para prueba de procesamiento masivo de texto"
	}
	parts := make([]string, 0, 8)
	for i := 0; i < 8; i++ {
		parts = append(parts, s)
	}
	return strings.Join(parts, " | ")
}

func cleanTextValue(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	s = multiSpace.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func lowerFirst(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func replaceFirstInsensitive(s string, old string, repl string) string {
	idx := strings.Index(strings.ToLower(s), strings.ToLower(old))
	if idx < 0 {
		return s
	}
	return s[:idx] + repl + s[idx+len(old):]
}

func introduceDoubleSpaces(s string, rng *rand.Rand) string {
	words := strings.Fields(s)
	if len(words) < 2 {
		return s + "  "
	}
	idx := rng.Intn(len(words)-1) + 1
	return strings.Join(words[:idx], " ") + "  " + strings.Join(words[idx:], " ")
}

func removeSomeAccents(s string) string {
	return accentReplacer.Replace(s)
}

func removeAccents(s string) string {
	return accentReplacer.Replace(s)
}

func corruptCommonChars(s string) string {
	return corruptReplacer.Replace(s)
}

func abbreviateCommonWords(s string) string {
	return abbrevReplacer.Replace(s)
}

func limitWords(s string, maxWords int) string {
	words := strings.Fields(s)
	if len(words) <= maxWords {
		return s
	}
	return strings.Join(words[:maxWords], " ")
}

func classifyTextType(sumilla string) string {
	s := strings.ToLower(removeAccents(sumilla))
	checks := []struct {
		Name  string
		Terms []string
	}{
		{"DESIGNACION", []string{"designan", "designa", "nombran", "nombra", "encargan", "encarga"}},
		{"APROBACION", []string{"aprueban", "aprueba", "aprobacion"}},
		{"AUTORIZACION", []string{"autorizan", "autoriza", "autorizacion"}},
		{"RESOLUCION", []string{"resolucion"}},
		{"DECRETO", []string{"decreto"}},
		{"RENUNCIA", []string{"renuncia", "aceptan"}},
		{"CONVENIO", []string{"convenio"}},
		{"ORDENANZA", []string{"ordenanza"}},
		{"SANCION", []string{"sancion", "sancionan", "multa"}},
	}
	for _, c := range checks {
		for _, term := range c.Terms {
			if strings.Contains(s, term) {
				return c.Name
			}
		}
	}
	return "OTROS"
}

func createCSVWithBOM(path string) (*os.File, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear %s: %w", path, err)
	}
	if _, err := file.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		file.Close()
		return nil, fmt.Errorf("no se pudo escribir BOM UTF-8 en %s: %w", path, err)
	}
	return file, nil
}

func auditRow(fields [8]string, archivoOrigen string, origenRegistro string, tipoSintetico string, regla string, mes string, opOrigen string) []string {
	row := make([]string, 0, len(auditCols))
	row = append(row, fields[:]...)
	row = append(row, archivoOrigen, origenRegistro, tipoSintetico, regla, mes, opOrigen)
	return row
}

func buildReports(result ConsolidationResult, stats GenerationStats, outputDir string) error {
	totalFinal := stats.Total
	fechaMin, fechaMax, hasDate := minMaxDates(result.Records)
	entidades := make(map[string]struct{})
	dispositivos := make(map[string]struct{})
	tiposSumilla := make(map[string]int)

	for _, rec := range result.Records {
		if strings.TrimSpace(rec.Fields[2]) != "" {
			entidades[rec.Fields[2]] = struct{}{}
		}
		if strings.TrimSpace(rec.Fields[3]) != "" {
			dispositivos[rec.Fields[3]] = struct{}{}
		}
		tiposSumilla[classifyTextType(rec.Fields[5])]++
	}

	fechaMinStr := ""
	fechaMaxStr := ""
	if hasDate {
		fechaMinStr = fechaMin.Format("2006-01-02")
		fechaMaxStr = fechaMax.Format("2006-01-02")
	}

	reportRows := [][]string{
		{"metrica", "valor"},
		{"registros_totales", strconv.Itoa(totalFinal)},
		{"registros_reales", strconv.Itoa(len(result.Records))},
		{"registros_sinteticos", strconv.Itoa(stats.Synthetic)},
		{"fecha_minima_real", fechaMinStr},
		{"fecha_maxima_real", fechaMaxStr},
		{"entidades_unicas", strconv.Itoa(len(entidades))},
		{"dispositivos_unicos", strconv.Itoa(len(dispositivos))},
		{"registros_antes_deduplicacion", strconv.Itoa(result.TotalBeforeDedup)},
		{"duplicados_exactos_eliminados", strconv.Itoa(result.Duplicates)},
		{"sinteticos_remuestreo_controlado", strconv.Itoa(stats.Remuestreo)},
		{"sinteticos_variacion_texto", strconv.Itoa(stats.VariacionTexto)},
		{"sinteticos_ruido_texto", strconv.Itoa(stats.RuidoTexto)},
		{"sinteticos_anomalia_controlada", strconv.Itoa(stats.Anomalia)},
		{"fechas_sinteticas_regeneradas_dentro_mes", strconv.Itoa(stats.FechaGenerada)},
	}
	if err := writeCSV(filepath.Join(outputDir, "reporte_generacion_sintetica.csv"), reportRows); err != nil {
		return err
	}

	originRows := [][]string{
		{"ORIGEN_REGISTRO", "registros"},
		{"REAL", strconv.Itoa(len(result.Records))},
		{"SINTETICO", strconv.Itoa(stats.Synthetic)},
	}
	if err := writeCSV(filepath.Join(outputDir, "resumen_por_origen.csv"), originRows); err != nil {
		return err
	}

	fileRows := [][]string{{"ARCHIVO_ORIGEN", "registros_reales_unicos"}}
	fileNames := make([]string, 0, len(result.CountsByFile))
	for name := range result.CountsByFile {
		fileNames = append(fileNames, name)
	}
	sort.Strings(fileNames)
	for _, name := range fileNames {
		fileRows = append(fileRows, []string{name, strconv.Itoa(result.CountsByFile[name])})
	}
	if err := writeCSV(filepath.Join(outputDir, "resumen_archivos_reales.csv"), fileRows); err != nil {
		return err
	}

	typeRows := [][]string{
		{"TIPO_SINTETICO", "registros"},
		{"REMUESTREO_CONTROLADO", strconv.Itoa(stats.Remuestreo)},
		{"VARIACION_TEXTO", strconv.Itoa(stats.VariacionTexto)},
		{"RUIDO_TEXTO", strconv.Itoa(stats.RuidoTexto)},
		{"ANOMALIA_CONTROLADA", strconv.Itoa(stats.Anomalia)},
	}
	if err := writeCSV(filepath.Join(outputDir, "resumen_sintetico_por_tipo.csv"), typeRows); err != nil {
		return err
	}

	monthRows := [][]string{{"MES", "registros_sinteticos"}}
	monthKeys := sortedKeys(stats.ByMonth)
	for _, key := range monthKeys {
		monthRows = append(monthRows, []string{key, strconv.Itoa(stats.ByMonth[key])})
	}
	if err := writeCSV(filepath.Join(outputDir, "resumen_sintetico_por_mes.csv"), monthRows); err != nil {
		return err
	}

	ruleRows := [][]string{{"REGLA_GENERACION", "registros_sinteticos"}}
	ruleKeys := sortedKeys(stats.ByRule)
	for _, key := range ruleKeys {
		ruleRows = append(ruleRows, []string{key, strconv.Itoa(stats.ByRule[key])})
	}
	if err := writeCSV(filepath.Join(outputDir, "resumen_reglas_generacion.csv"), ruleRows); err != nil {
		return err
	}

	textTypeRows := [][]string{{"TIPO_TEXTO_SUMILLA_REAL", "registros_reales"}}
	textKeys := sortedKeys(tiposSumilla)
	for _, key := range textKeys {
		textTypeRows = append(textTypeRows, []string{key, strconv.Itoa(tiposSumilla[key])})
	}
	if err := writeCSV(filepath.Join(outputDir, "resumen_tipo_sumilla_real.csv"), textTypeRows); err != nil {
		return err
	}

	return nil
}

func writeCSV(path string, rows [][]string) error {
	file, err := createCSVWithBOM(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	if err := writer.WriteAll(rows); err != nil {
		return fmt.Errorf("error escribiendo %s: %w", path, err)
	}
	return nil
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func minMaxDates(records []Record) (time.Time, time.Time, bool) {
	var minDate time.Time
	var maxDate time.Time
	hasDate := false
	for _, rec := range records {
		date, ok := parseDate(rec.Fields[0])
		if !ok {
			continue
		}
		if !hasDate || date.Before(minDate) {
			minDate = date
		}
		if !hasDate || date.After(maxDate) {
			maxDate = date
		}
		hasDate = true
	}
	return minDate, maxDate, hasDate
}

func parseDate(value string) (time.Time, bool) {
	s := strings.TrimSpace(value)
	if s == "" {
		return time.Time{}, false
	}
	if len(s) == 8 && onlyDigits(s) {
		if t, err := time.Parse("20060102", s); err == nil {
			return t, true
		}
	}

	layouts := []string{
		"02/01/2006",
		"2/1/2006",
		"2006-01-02",
		"02-01-2006",
		"2-1-2006",
		"02/01/06",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func onlyDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func formatInt(n int) string {
	s := strconv.Itoa(n)
	if n < 0 {
		return "-" + formatInt(-n)
	}
	var buf bytes.Buffer
	rem := len(s) % 3
	if rem == 0 {
		rem = 3
	}
	buf.WriteString(s[:rem])
	for i := rem; i < len(s); i += 3 {
		buf.WriteByte(',')
		buf.WriteString(s[i : i+3])
	}
	return buf.String()
}
