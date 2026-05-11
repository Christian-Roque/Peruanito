# Peruanito

Pipeline concurrente en Go para el procesamiento masivo de textos legales del Diario Oficial El Peruano.

## Objetivo

Evaluar experimentalmente el impacto de la concurrencia en un pipeline NLP legal, comparando una versión secuencial contra una versión concurrente basada en goroutines, channels, `sync.WaitGroup` y `sync.Mutex`.

El proyecto mide:

- tiempo de ejecución;
- speedup;
- throughput;
- eficiencia paralela;
- estabilidad en 100 ejecuciones;
- exactitud funcional de los resultados;
- tradeoff entre rendimiento y uso de recursos.

## Pipeline de procesamiento

Cada sumilla legal pasa por:

1. limpieza y normalización;
2. tokenización;
3. eliminación de stopwords;
4. eliminación de duplicados de tokens;
5. deduplicación global de textos procesados.

La versión concurrente usa un patrón **Worker Pool**. La deduplicación global se protege con `sync.Mutex` para evitar condiciones de carrera.

## Generación del dataset sintético

```bash
go run generar_dataset_sintetico_elperuano_enriquecido.go --input-dir ./datasets --output-dir ./salida --target-total 1000000
```

Archivo generado:

```text
salida/dataset_para_go.csv
```

## Benchmark formal PC2

```bash
go run pc2.go -input salida/dataset_para_go.csv -runs 100 -workers 1,2,4,8,16,32,64,100 -tipo-data mixto -out resultados/benchmark_pc2_raw.csv -summary resultados/benchmark_pc2_resumen.csv
```

Para una prueba rápida durante desarrollo:

```bash
go run pc2.go -input salida/dataset_para_go.csv -runs 3 -workers 1,2,4 -max-records 5000
```

## Gráficos

```bash
python scripts/graficos_benchmark.py --raw resultados/benchmark_pc2_raw.csv --summary resultados/benchmark_pc2_resumen.csv --out-dir graficos
```

## Condición de carrera

Demostración intencional para ejecutar con el detector de carreras de Go:

```bash
go run -race ./cmd/race_condition_demo
```

Esta prueba debe mostrar `WARNING: DATA RACE` y sirve para justificar el uso de exclusión mutua en el procesamiento real.

## Scripts automáticos

En Windows PowerShell:

```powershell
./scripts/ejecutar_benchmark.ps1
```

En Linux/macOS/Git Bash:

```bash
./scripts/ejecutar_benchmark.sh
```

## Evidencias sugeridas

Guardar en `evidencias/` las capturas del Administrador de tareas durante las ejecuciones con:

- 1 worker;
- 8 workers;
- 32 workers;
- 100 workers.

Estas capturas permiten explicar el tradeoff entre reducción de tiempo y mayor uso de CPU/memoria.
