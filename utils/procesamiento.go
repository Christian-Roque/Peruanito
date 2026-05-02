package utils

import (
	"sync"
)

// 🔹 Procesamiento SECUENCIAL
func ProcesarSecuencial(records [][]string, indiceSumilla int) int {

	vistosTexto := make(map[string]bool)
	total := 0

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

		// Pipeline NLP
		limpio := LimpiarTexto(sumilla)

		if vistosTexto[limpio] {
			continue
		}
		vistosTexto[limpio] = true

		tokens := Tokenizar(limpio)
		sinStopwords := RemoverStopwords(tokens)
		final := RemoverDuplicadosTokens(sinStopwords)

		if len(final) > 0 {
			total++
		}
	}

	return total
}

// 🔹 Procesamiento CONCURRENTE (Worker Pool)
func ProcesarConcurrente(records [][]string, indiceSumilla int, numWorkers int) int {

	tareas := make(chan string, 1000)
	resultados := make(chan []string, 1000)

	var wg sync.WaitGroup

	// 🔹 Workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for texto := range tareas {

				// Pipeline NLP
				limpio := LimpiarTexto(texto)
				tokens := Tokenizar(limpio)
				sinStopwords := RemoverStopwords(tokens)
				final := RemoverDuplicadosTokens(sinStopwords)

				if len(final) > 0 {
					resultados <- final
				}
			}
		}()
	}

	// 🔹 Productor
	go func() {
		for i, row := range records {
			if i == 0 {
				continue
			}

			if len(row) <= indiceSumilla {
				continue
			}

			sumilla := row[indiceSumilla]
			if sumilla != "" {
				tareas <- sumilla
			}
		}
		close(tareas)
	}()

	// 🔹 Cerrar resultados cuando terminen los workers
	go func() {
		wg.Wait()
		close(resultados)
	}()

	total := 0

	for range resultados {
		total++
	}

	return total
}
