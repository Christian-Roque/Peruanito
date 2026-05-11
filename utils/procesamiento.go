package utils

import (
	"strings"
	"sync"
)

// ProcessingStats resume el resultado funcional del pipeline.
// Se usa para validar que la versión secuencial y la concurrente procesen
// la misma cantidad de registros únicos y no solo para medir tiempo.
type ProcessingStats struct {
	TotalFilas         int
	FilasValidas       int
	FilasInvalidas     int
	ProcesadosUnicos   int
	DuplicadosGlobales int
	VaciosPostPipeline int
}

const (
	resultadoProcesado = iota
	resultadoDuplicado
	resultadoVacio
)

// ProcesarSecuencial mantiene la firma original para compatibilidad.
func ProcesarSecuencial(records [][]string, indiceSumilla int) int {
	return ProcesarSecuencialStats(records, indiceSumilla).ProcesadosUnicos
}

// ProcesarConcurrente mantiene la firma original para compatibilidad.
func ProcesarConcurrente(records [][]string, indiceSumilla int, numWorkers int) int {
	return ProcesarConcurrenteStats(records, indiceSumilla, numWorkers).ProcesadosUnicos
}

// ProcesarSecuencialStats ejecuta el pipeline NLP de forma lineal.
// Incluye deduplicación global para que la comparación con la versión
// concurrente sea funcionalmente equivalente.
func ProcesarSecuencialStats(records [][]string, indiceSumilla int) ProcessingStats {
	stats := ProcessingStats{}
	vistosTexto := make(map[string]bool)

	for i, row := range records {
		if i == 0 {
			continue
		}
		stats.TotalFilas++

		if len(row) <= indiceSumilla || strings.TrimSpace(row[indiceSumilla]) == "" {
			stats.FilasInvalidas++
			continue
		}
		stats.FilasValidas++

		limpio := LimpiarTexto(row[indiceSumilla])
		if vistosTexto[limpio] {
			stats.DuplicadosGlobales++
			continue
		}
		vistosTexto[limpio] = true

		tokens := Tokenizar(limpio)
		sinStopwords := RemoverStopwords(tokens)
		final := RemoverDuplicadosTokens(sinStopwords)

		if len(final) > 0 {
			stats.ProcesadosUnicos++
		} else {
			stats.VaciosPostPipeline++
		}
	}

	return stats
}

// ProcesarConcurrenteStats ejecuta el mismo pipeline con patrón Worker Pool.
// El mapa vistosTexto es un recurso compartido y se protege con sync.Mutex
// para evitar condiciones de carrera durante la deduplicación global.
func ProcesarConcurrenteStats(records [][]string, indiceSumilla int, numWorkers int) ProcessingStats {
	if numWorkers < 1 {
		numWorkers = 1
	}

	stats := ProcessingStats{}
	tareas := make(chan string, 4096)
	resultados := make(chan int, 4096)

	var wg sync.WaitGroup
	var mutex sync.Mutex
	vistosTexto := make(map[string]bool)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for texto := range tareas {
				limpio := LimpiarTexto(texto)

				mutex.Lock()
				if vistosTexto[limpio] {
					mutex.Unlock()
					resultados <- resultadoDuplicado
					continue
				}
				vistosTexto[limpio] = true
				mutex.Unlock()

				tokens := Tokenizar(limpio)
				sinStopwords := RemoverStopwords(tokens)
				final := RemoverDuplicadosTokens(sinStopwords)

				if len(final) > 0 {
					resultados <- resultadoProcesado
				} else {
					resultados <- resultadoVacio
				}
			}
		}()
	}

	go func() {
		for i, row := range records {
			if i == 0 {
				continue
			}
			stats.TotalFilas++

			if len(row) <= indiceSumilla || strings.TrimSpace(row[indiceSumilla]) == "" {
				stats.FilasInvalidas++
				continue
			}

			stats.FilasValidas++
			tareas <- row[indiceSumilla]
		}
		close(tareas)
	}()

	go func() {
		wg.Wait()
		close(resultados)
	}()

	for r := range resultados {
		switch r {
		case resultadoProcesado:
			stats.ProcesadosUnicos++
		case resultadoDuplicado:
			stats.DuplicadosGlobales++
		case resultadoVacio:
			stats.VaciosPostPipeline++
		}
	}

	return stats
}
