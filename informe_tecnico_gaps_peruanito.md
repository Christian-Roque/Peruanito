# Informe técnico de GAPs

## Modelo de IA seleccionado

**Modelo seleccionado:** GPT-5.5 Thinking.

**Justificación:** se selecciona un modelo de razonamiento avanzado porque la revisión exige combinar auditoría estática del repositorio, validación de comandos reales, lectura de código Go, análisis de concurrencia, seguridad, benchmarking, reproducibilidad y revisión de artefactos SPIN/Promela. El análisis se realizó sobre el ZIP subido en esta conversación, sin asumir archivos externos al repositorio.

## Resumen ejecutivo

El repositorio **Peruanito** implementa un pipeline de procesamiento de textos legales en Go, con una versión secuencial y una versión concurrente basada en worker pool, channels, `sync.WaitGroup` y `sync.Mutex`. La funcionalidad principal de `pc2.go` pudo ejecutarse correctamente en una prueba rápida con 5,000 registros y produjo CSV de resultados. También se pudo generar un dataset consolidado sin sintéticos a partir de los CSV incluidos en `datasets/`.

Sin embargo, el repositorio presenta GAPs relevantes para una entrega técnica formal. El problema más importante es que el módulo completo **no compila ni es testeable con `go test ./...`** debido a artefactos de SPIN ubicados en la raíz (`pan.c`, `pan.m`, etc.) y, aun retirándolos en una copia de prueba, por múltiples funciones `main()` dentro del mismo paquete raíz. Además, el README documenta comandos y scripts que no existen realmente, como `./cmd/race_condition_demo`, `scripts/ejecutar_benchmark.ps1` y `scripts/ejecutar_benchmark.sh`.

En concurrencia, la implementación principal evita una carrera evidente sobre el mapa global de deduplicación mediante `sync.Mutex`, y una ejecución rápida con `go run -race pc2.go` no reportó `DATA RACE`. No obstante, existen cuellos de botella de escalabilidad por contención del mutex global, ausencia de pruebas automatizadas de concurrencia, ausencia de cancelación por contexto y falta de una arquitectura por paquetes que permita testear los componentes de forma profesional.

En rendimiento, el uso de `ReadAll` y `os.ReadFile` carga datasets completos en memoria, y `LimpiarTexto` recompila expresiones regulares y recrea reemplazadores por cada texto procesado. Esto afecta throughput y escalabilidad cuando se buscan ejecuciones grandes, como el millón de registros documentado en el README.

En SPIN/Promela, existe un modelo de verificación (`modelo_verificacion_spin.pml`) con señal de cierre, exclusión mutua y aserciones. No se pudo regenerar el verificador porque `spin` no está instalado en el entorno de revisión. Como evidencia parcial, se ejecutó el binario `./pan` ya incluido y reportó `errors: 0`. Esta evidencia debe considerarse válida solo como ejecución del artefacto incluido, no como regeneración independiente desde cero.

## Metodología de revisión

La revisión se realizó sobre el archivo ZIP subido: `Peruanito-main (3).zip`. Se descomprimió el repositorio y se revisaron estructura, README, código Go, scripts Python, datasets y modelos Promela.

Validaciones ejecutadas:

| Validación | Resultado observado |
|---|---|
| Listado de estructura del repositorio | Existen `README.md`, `go.mod`, `pc1.go`, `pc2.go`, `generar_dataset_sintetico_elperuano_enriquecido.go`, `utils/`, `datasets/`, `graficos/`, `scripts/graficos_benchmark.py`, `modelo.pml`, `modelo_verificacion_spin.pml` y artefactos `pan*`. |
| `go version` | `go version go1.23.2 linux/amd64`. El `go.mod` declara `go 1.22`. |
| `gofmt -l *.go utils/*.go` | No devolvió archivos, por lo que el código revisado está formateado con `gofmt`. |
| `go test ./...` | Falló: `C source files not allowed when not using cgo or SWIG: pan.c`. |
| `go test ./...` en copia sin artefactos `pan*` | Falló por múltiples `main`: `main redeclared` entre `pc1.go`, `pc2.go` y `generar_dataset_sintetico_elperuano_enriquecido.go`. |
| Generación rápida de dataset | Ejecutó correctamente con `--target-total 0 --skip-audit`. Se consolidaron 251,540 registros reales. |
| Benchmark rápido `pc2.go` | Ejecutó correctamente con `-runs 2 -workers 1,2,4 -max-records 5000`; todos los resultados comparados salieron `correcto=true`. |
| Race detector sobre `pc2.go` | `go run -race pc2.go ... -max-records 500` terminó sin reportar `WARNING: DATA RACE`. |
| Comando README `go run -race ./cmd/race_condition_demo` | Falló: `directory not found`. |
| Script de gráficos | Ejecutó correctamente, pero mostró advertencia de deprecación de Matplotlib por uso de `labels` en `plt.boxplot`. |
| `spin -V` | No disponible en el entorno: `spin: command not found`. |
| `./pan` incluido | Ejecutó y reportó `errors: 0`, depth 92, 36,539 states stored. |
| `./pan.exe` incluido | Se ejecutó como ELF Linux, pero no terminó dentro del tiempo de revisión; parece inconsistente o generado para otro modelo. |
| Búsqueda simple de secretos | No se encontraron secretos evidentes; los matches fueron falsos positivos por palabras como `tokenización`. |
| Muestra de CSV injection en primeras 1,000 filas por archivo | No se detectaron celdas sospechosas iniciando con `=`, `+`, `-` o `@` en la muestra revisada. No obstante, el código no mitiga el riesgo para datos futuros. |

Limitaciones:

- No se ejecutó el benchmark completo de 100 runs ni con 1,000,000 registros por costo de tiempo y recursos.
- No se pudo regenerar `pan.c` desde Promela porque SPIN no está instalado en el entorno.
- No se realizó auditoría SAST especializada con herramientas como `gosec`, `staticcheck`, `govulncheck` o `pprof`, porque no forman parte obligatoria del entorno usado.
- La revisión de CSV injection fue muestral; el GAP se basa principalmente en ausencia de sanitización en el código de salida.

## Hallazgos priorizados

| ID | Severidad | Categoría | GAP | Evidencia | Impacto | Recomendación |
|---|---|---|---|---|---|---|
| G-001 | Crítico | Estructura / Build | El módulo completo no es testeable con `go test ./...` por artefactos SPIN en la raíz. | `go test ./...` falla con `C source files not allowed when not using cgo or SWIG: pan.c`. En la raíz existen `pan.c`, `pan.m`, `pan.h`, `pan.p`, `pan.t`, `pan.b`, `pan`, `pan.exe`. | Bloquea pruebas automatizadas, CI/CD y validación estándar de Go. | Mover artefactos SPIN a `verification/spin/generated/` o excluirlos del módulo Go. Agregar reglas en `.gitignore` para `pan*` generados. |
| G-002 | Crítico | Estructura / Build | Aun retirando artefactos SPIN, el paquete raíz no compila como módulo por múltiples `main()`. | En copia sin `pan*`, `go test ./...` falla: `main redeclared` entre `pc1.go`, `pc2.go` y `generar_dataset_sintetico_elperuano_enriquecido.go`. | Impide pruebas unitarias globales, empaquetado profesional y CI. | Reorganizar en `cmd/pc1/main.go`, `cmd/pc2/main.go`, `cmd/generador/main.go` y mover lógica reusable a paquetes internos (`internal/pipeline`, `internal/benchmark`, `internal/dataset`). |
| G-003 | Alto | Documentación / Reproducibilidad | El README documenta comandos y scripts que no existen. | README indica `go run -race ./cmd/race_condition_demo`, `./scripts/ejecutar_benchmark.ps1` y `./scripts/ejecutar_benchmark.sh`; esos archivos/carpetas no existen. El comando de race falla con `directory not found`. | Un evaluador no puede reproducir lo documentado; resta confiabilidad al informe experimental. | Actualizar README o crear los scripts/carpetas faltantes. Validar todos los comandos en una máquina limpia antes de entregar. |
| G-004 | Alto | Pruebas | No existen pruebas unitarias ni benchmarks nativos de Go. | No hay archivos `*_test.go`; `go test -race ./utils` devuelve `[no test files]`. | No hay garantía automática de equivalencia funcional, regresión, race safety o rendimiento. | Crear tests para limpieza, tokenización, stopwords, deduplicación y equivalencia secuencial/concurrente. Agregar `BenchmarkProcesarSecuencial` y `BenchmarkProcesarConcurrente`. |
| G-005 | Alto | Rendimiento | Carga completa de archivos en memoria usando `ReadAll` y `os.ReadFile`. | `pc2.go` usa `reader.ReadAll()`; `pc1.go` usa `reader.ReadAll()`; el generador usa `os.ReadFile(path)`. | Alto consumo de memoria, menor escalabilidad y riesgo de fallos con datasets grandes o ejecución de 1,000,000 registros. | Migrar a lectura streaming con `csv.Reader.Read()` por lotes, procesamiento por chunks y medición separada de I/O vs CPU. |
| G-006 | Alto | Rendimiento / Código | `LimpiarTexto` recompila regex y recrea replacer por cada registro. | En `utils/limpieza.go`, `strings.NewReplacer(...)` y `regexp.MustCompile(...)` están dentro de `LimpiarTexto`. | Sobrecosto significativo en pipelines masivos; reduce throughput y puede distorsionar benchmarks de concurrencia. | Declarar `var accentReplacer = strings.NewReplacer(...)` y `var nonAlphaNumRe = regexp.MustCompile(...)` a nivel de paquete. |
| G-007 | Alto | Concurrencia / Escalabilidad | La deduplicación global usa un único `sync.Mutex`, generando contención a medida que aumentan workers. | `ProcesarConcurrenteStats` protege `vistosTexto` con un mutex global alrededor de check/insert. | El speedup puede degradarse en 32, 64 o 100 workers; el benchmark mide más contención que paralelismo real. | Evaluar particionamiento por hash/sharding de mapas, `sync.Map` con benchmark comparativo, o deduplicación por etapas. |
| G-008 | Medio | Concurrencia | No existe cancelación por contexto ni propagación estructurada de errores en el worker pool. | `ProcesarConcurrenteStats` no recibe `context.Context`; los workers solo terminan por cierre de canal. | Si en el futuro se agrega I/O, errores de parsing o procesamiento externo, no habrá forma limpia de cancelar trabajos. | Incorporar `context.Context`, canal de errores y patrón de cancelación temprana. |
| G-009 | Medio | Seguridad | Riesgo potencial de CSV injection en archivos generados. | Los CSV de salida escriben campos provenientes de datasets/sumillas con `csv.Writer`, pero no neutralizan celdas que empiecen por `=`, `+`, `-` o `@`. La muestra revisada no detectó casos, pero el código no lo previene. | Si los CSV se abren en Excel, datos maliciosos futuros podrían interpretarse como fórmulas. | Sanitizar celdas antes de escribir CSV para consumo en Excel, por ejemplo anteponiendo `'` o espacio controlado a prefijos peligrosos. |
| G-010 | Medio | Seguridad / Repositorio | El repositorio incluye datasets y binarios/artefactos generados. | Carpeta `datasets/` pesa ~85 MB; existen `pan`, `pan.exe`, `pan.c` y gráficos generados. `.gitignore` tiene `*.csv`, pero los CSV están presentes en el ZIP. | Aumenta tamaño del repositorio, dificulta revisión, puede exponer datos sin clasificación y rompe builds Go por archivos C/Objective-C. | Separar datos y artefactos en releases, storage externo o Git LFS. Mantener en Git solo muestras pequeñas y reproducibles. |
| G-011 | Medio | Benchmarking | El benchmark es manual y no integra herramientas estándar de Go. | `pc2.go` implementa medición propia con `time.Now()`, CSV raw y summary; no hay `testing.B`, `benchstat`, perfiles CPU/memoria ni control de `GOMAXPROCS`. | Las conclusiones de rendimiento pueden ser menos defendibles ante revisión técnica. | Agregar benchmarks nativos, `benchstat`, `pprof`, registro de hardware, versión Go, `GOMAXPROCS`, tamaño exacto de entrada y warm-up. |
| G-012 | Medio | Reproducibilidad experimental | Los gráficos incluidos no tienen sus CSV fuente dentro de `resultados/`. | Existen imágenes en `graficos/`, pero no se observó carpeta `resultados/` con `benchmark_pc2_raw.csv` y `benchmark_pc2_resumen.csv` en el ZIP. | No se puede auditar si los gráficos corresponden al benchmark documentado. | Versionar CSV resumen pequeños o documentar hash/ruta de artefactos. Regenerar gráficos desde CSV trazables. |
| G-013 | Medio | SPIN / Promela | La evidencia SPIN no es completamente reproducible desde cero en el entorno revisado. | `spin -V` no está disponible; solo se pudo ejecutar el `./pan` incluido, que reportó `errors: 0`. | Un auditor externo no puede verificar que `pan.c` proviene exactamente del `.pml` actual. | Documentar instalación de SPIN, comandos exactos y salida esperada. Generar `pan.c` dentro de un script reproducible. |
| G-014 | Medio | SPIN / Modelo formal | El modelo formal es pequeño y abstrae demasiados aspectos del Go real. | `modelo_verificacion_spin.pml` fija `N_TAREAS=4`, `N_WORKERS=2`, buffer `[2]`, y modela solo exclusión mutua/consumo total. | La verificación no cubre duplicados, variabilidad de workers, cierre real de channels en Go, errores, ni carga grande. | Crear variantes parametrizadas: límites de buffer, más workers, duplicados, tareas vacías, productor lento/rápido y propiedades LTL. |
| G-015 | Bajo | SPIN / Limpieza de repo | Existe `modelo.pml` como modelo demo sin señales de cierre para workers/consumidor. | `modelo.pml` tiene workers y consumidor con bucles que no reciben señal FIN ni condición de salida. | Puede confundir al evaluador si se toma como evidencia formal. | Marcarlo como modelo exploratorio o moverlo a `verification/spin/demos/`; dejar como evidencia principal solo el modelo verificable. |
| G-016 | Bajo | Código | Uso de `panic` y rutas rígidas en `pc1.go`. | `obtenerArchivosCSV()` hace `panic(err)` si no existe `datasets`; `numWorkers := 4` está hardcodeado. | Menor robustez y menor reutilización. | Convertir `pc1` a CLI con flags y manejo de errores retornables. |
| G-017 | Bajo | Python / Gráficos | Script de gráficos usa argumento deprecado de Matplotlib. | Ejecución mostró `MatplotlibDeprecationWarning`: `labels` fue renombrado a `tick_labels`. | Puede fallar en futuras versiones de Matplotlib. | Cambiar `plt.boxplot(data, labels=labels, ...)` por `plt.boxplot(data, tick_labels=labels, ...)` según versión soportada. |

## Análisis por dimensión

### Calidad de código

El código tiene una separación inicial aceptable en `utils/` para limpieza, tokenización, stopwords, deduplicación y procesamiento. Esto facilita entender el pipeline base. También se observa consistencia de formato, ya que `gofmt -l` no reportó archivos pendientes de formateo.

El principal GAP de calidad es arquitectónico: hay tres programas ejecutables (`pc1.go`, `pc2.go` y `generar_dataset_sintetico_elperuano_enriquecido.go`) en el mismo paquete raíz `main`. Esta estructura permite ejecutar archivos individuales con `go run pc2.go`, pero no permite compilar/testear el módulo completo de forma estándar. Para un proyecto profesional en Go, los ejecutables deberían ubicarse en `cmd/<nombre>/main.go`, y la lógica reusable debería ir en paquetes separados.

También existe duplicidad funcional entre `pc1.go` y `utils/procesamiento.go`: ambos implementan una forma de procesamiento concurrente, deduplicación y conteo. `pc1.go` parece más exploratorio, con workers hardcodeados y salida por consola, mientras que `pc2.go` es el benchmark formal. Se recomienda dejar `pc1.go` como demo en una carpeta separada o retirarlo de la ruta principal del módulo.

El manejo de errores es mejor en `pc2.go` y el generador que en `pc1.go`, pero aún puede mejorar. `pc1.go` usa `panic` si no existe `datasets/`, mientras que una CLI debería retornar mensajes controlados y códigos de salida consistentes. Asimismo, varias funciones podrían volverse testeables si no dependieran directamente de rutas o de impresión por consola.

### Seguridad

No se encontraron secretos evidentes mediante una búsqueda simple de patrones como `password`, `secret`, `api_key`, `token`, claves privadas o tokens de GitHub. El módulo Go no usa dependencias externas, lo cual reduce superficie de ataque en la parte Go. En Python sí se usan `pandas` y `matplotlib`, pero no hay archivo de versiones fijadas.

El GAP principal de seguridad está en la gestión de datos y artefactos. El ZIP contiene una carpeta `datasets/` de aproximadamente 85 MB, gráficos generados y binarios/artefactos SPIN. Aunque los datos parecen corresponder a textos legales públicos, no hay una clasificación explícita de datos, licencia, fuente exacta, hash, ni criterio de publicación. En un repositorio público o académico, lo recomendable es incluir una muestra pequeña y documentar cómo obtener o regenerar el dataset completo.

También existe riesgo potencial de CSV injection. La revisión muestral no encontró celdas sospechosas iniciando con `=`, `+`, `-` o `@` en las primeras 1,000 filas por archivo, pero el código no implementa sanitización. Como los CSV generados pueden abrirse en Excel, cualquier dato futuro que empiece con esos caracteres podría interpretarse como fórmula.

La validación de rutas es básica. Para uso local por CLI esto no es crítico, pero si el programa se expusiera como servicio o se integrara a una interfaz, habría que limitar rutas de entrada/salida, normalizar paths y evitar sobrescritura accidental de ubicaciones no previstas.

### Concurrencia

La implementación principal de `ProcesarConcurrenteStats` usa un patrón worker pool razonable: canal de tareas, canal de resultados, `WaitGroup`, cierre de `tareas` y cierre de `resultados` después de finalizar workers. La deduplicación global se protege con `sync.Mutex`, lo que evita una carrera directa sobre el mapa compartido `vistosTexto`. En una prueba rápida con `go run -race pc2.go`, el race detector no reportó carreras.

El diseño también compara los resultados concurrentes contra un baseline secuencial mediante `statsIguales`, lo cual es una buena práctica porque no mide solo tiempo, sino también equivalencia funcional básica.

El GAP principal está en escalabilidad. Todos los workers compiten por el mismo mutex global para revisar e insertar en `vistosTexto`. Para pocos workers puede ser suficiente, pero con 32, 64 o 100 workers la contención puede limitar el speedup. En un pipeline NLP masivo, convendría evaluar particionamiento por hash, mapas segmentados, reducción local por worker y fusión final, o una estrategia de deduplicación por lotes.

Otro GAP es la ausencia de cancelación. El pipeline no recibe `context.Context`; si en el futuro se agrega lectura remota, escritura externa, parsing costoso o errores internos, no habrá mecanismo estándar para detener productores y workers de forma segura.

`pc1.go` tiene una implementación más débil: usa un solo mutex tanto para deduplicación como para contadores e impresión, ampliando la sección crítica; además usa `numWorkers := 4` hardcodeado y no está integrado al benchmark formal.

### Rendimiento

El benchmark de `pc2.go` separa la carga de CSV del procesamiento, porque lee todos los registros antes de iniciar las mediciones. Esto es válido si se declara que se está midiendo el procesamiento en memoria, no el pipeline end-to-end. Sin embargo, esta decisión debe explicitarse en el informe final porque el throughput no incluye I/O.

El mayor problema de rendimiento es la carga completa de datos en memoria. `pc2.go` y `pc1.go` usan `reader.ReadAll()`, mientras que el generador usa `os.ReadFile()` antes de parsear cada CSV. Esto puede funcionar para los 251,540 registros reales observados, pero no es la mejor opción si se busca procesar datasets de 1,000,000 registros o más. Un enfoque streaming con chunks permitiría menor memoria y mejor control de backpressure.

`LimpiarTexto` recompila una expresión regular y reconstruye un `strings.NewReplacer` por cada texto. Dado que esta función se invoca por registro, el costo acumulado puede ser alto y afecta directamente la medición del pipeline. Estos objetos deberían declararse una sola vez a nivel de paquete.

El benchmark manual produce métricas útiles: tiempo, speedup, throughput, eficiencia y exactitud. No obstante, para defender resultados técnicamente conviene complementarlo con `testing.B`, `benchstat`, `pprof`, medición de asignaciones (`allocs/op`), memoria, CPU y configuración de `GOMAXPROCS`.

### Pruebas y reproducibilidad

Actualmente no hay archivos `*_test.go`. Esto significa que no existen pruebas unitarias para componentes básicos como limpieza, tokenización, eliminación de stopwords, deduplicación, parsing de workers o equivalencia secuencial/concurrente.

La reproducibilidad está parcialmente lograda: el generador tiene semilla configurable (`--seed`, default 2026), y `pc2.go` permite configurar runs, workers, input, output, summary, tipo de data y límite de registros. Además, el script Python pudo generar gráficos a partir de CSV raw y summary.

No obstante, el repositorio no incluye los CSV fuente de los gráficos dentro de una carpeta `resultados/`. Por tanto, los PNG existentes en `graficos/` no son auditables directamente desde el ZIP. Para una entrega formal, cada gráfico debería poder regenerarse desde CSV trazables o incluir hash de los resultados.

El README promete scripts automáticos para Windows y Linux/macOS, pero esos scripts no existen. También promete una demostración de carrera con `cmd/race_condition_demo`, pero esa carpeta no existe. Esto debe corregirse porque afecta la confianza en toda la documentación experimental.

No se observa CI/CD. Un pipeline mínimo debería ejecutar `gofmt`, `go vet`, `go test ./...`, `go test -race`, generación rápida de dataset sintético, benchmark reducido y validación de SPIN si el entorno lo permite.

### Verificación formal SPIN

El repositorio contiene dos modelos Promela: `modelo.pml` y `modelo_verificacion_spin.pml`. El segundo es el más apropiado para evidencia formal porque modela tareas, workers, una señal de fin (`SENAL_FIN`), sección crítica, aserción de exclusión mutua (`assert(en_critica == 1)`) y una aserción final de consumo total (`assert(total_consumidos == N_TAREAS)`).

No se pudo ejecutar `spin -a modelo_verificacion_spin.pml` porque SPIN no está instalado en el entorno de revisión. Sin embargo, sí se ejecutó el binario `./pan` incluido y reportó:

```text
Spin Version 6.5.2
Full statespace search
assertion violations +
invalid end states +
depth reached 92, errors: 0
36539 states, stored
59661 transitions
```

Esto es una evidencia positiva, pero limitada. Al no poder regenerar `pan.c`, no se puede garantizar desde el entorno que el binario `./pan` provenga exactamente del `.pml` actual. Además, existe `pan.exe`, que en este entorno se identificó también como ELF Linux y no terminó dentro del tiempo de revisión, lo que sugiere que puede estar desactualizado o corresponder a otro modelo.

El modelo formal actual es pequeño: `N_TAREAS=4`, `N_WORKERS=2` y buffers de tamaño 2. Es útil como demostración de ausencia de deadlock y exclusión mutua en un caso reducido, pero no prueba exhaustivamente el comportamiento del programa Go real. No modela duplicados, filas inválidas, textos vacíos, distintos niveles de workers, errores, cancelación, ni el cierre real de canales de Go.

`modelo.pml` debe tratarse como modelo exploratorio porque no tiene señales de cierre para workers ni consumidor. Si se deja junto al modelo formal, puede confundir al evaluador.

### Documentación y estructura del repositorio

El README explica bien el objetivo general del proyecto y el pipeline conceptual. También documenta comandos de generación, benchmark y gráficos. Sin embargo, no todos los comandos coinciden con archivos reales del repositorio.

Estructura observada:

```text
README.md
go.mod
pc1.go
pc2.go
generar_dataset_sintetico_elperuano_enriquecido.go
utils/
scripts/graficos_benchmark.py
datasets/
graficos/
modelo.pml
modelo_verificacion_spin.pml
pan, pan.exe, pan.c, pan.h, pan.m, pan.p, pan.t, pan.b
```

Elementos documentados pero no encontrados:

```text
cmd/race_condition_demo/
scripts/ejecutar_benchmark.ps1
scripts/ejecutar_benchmark.sh
resultados/benchmark_pc2_raw.csv
resultados/benchmark_pc2_resumen.csv
evidencias/
```

`.gitignore` contiene `salida/` y `*.csv`, pero el ZIP contiene múltiples CSV dentro de `datasets/`. Esto puede deberse a que fueron agregados antes de la regla o a que el ZIP no representa exactamente el estado del control de versiones, pero como repositorio entregable genera inconsistencia.

## Plan de remediación

| Prioridad | Acción | Detalle | Resultado esperado |
|---|---|---|---|
| 1 | Reestructurar el módulo Go | Crear `cmd/pc1/main.go`, `cmd/pc2/main.go`, `cmd/generador/main.go`; mover lógica a `internal/pipeline`, `internal/dataset`, `internal/benchmark`. | `go test ./...` debe compilar sin conflictos de `main`. |
| 2 | Separar artefactos SPIN del módulo Go | Mover `pan*` a `verification/spin/generated/` o eliminarlos del repo; dejar scripts para regenerarlos. | El build Go no se rompe por `pan.c` o `pan.m`. |
| 3 | Corregir README | Eliminar comandos inexistentes o crear `cmd/race_condition_demo`, `scripts/ejecutar_benchmark.ps1` y `.sh`. | Documentación reproducible. |
| 4 | Crear pruebas unitarias | Tests para `LimpiarTexto`, `Tokenizar`, `RemoverStopwords`, deduplicación, `parseWorkers`, equivalencia secuencial/concurrente. | Cobertura mínima y regresión controlada. |
| 5 | Agregar race tests y CI | Ejecutar `go test ./...`, `go test -race ./...`, `go vet`, `gofmt`, benchmark reducido y generación rápida de gráficos. | Validación automática antes de cada entrega. |
| 6 | Optimizar limpieza de texto | Mover regex y replacer a variables globales del paquete. | Menor overhead por registro y mayor throughput. |
| 7 | Migrar lectura a streaming/chunks | Reemplazar `ReadAll` y `os.ReadFile` donde aplique. | Menor memoria y mejor escalabilidad. |
| 8 | Reducir contención concurrente | Probar sharded maps, reducción local por worker y fusión final. | Mejor speedup en muchos workers. |
| 9 | Formalizar benchmarking | Incorporar `testing.B`, `benchstat`, `pprof`, registro de hardware, versión Go, `GOMAXPROCS` y tamaño de dataset. | Resultados defendibles ante auditoría. |
| 10 | Fortalecer SPIN | Documentar instalación, comando `spin -a`, compilación `gcc -o pan pan.c`, ejecución `./pan`; añadir modelos con duplicados, buffers variables y más workers. | Evidencia formal reproducible. |
| 11 | Mitigar CSV injection | Sanitizar campos al exportar CSV para Excel. | Menor riesgo al abrir resultados en hojas de cálculo. |
| 12 | Limpiar datos/artefactos del repo | Usar muestras pequeñas, Git LFS o almacenamiento externo para datasets grandes; documentar fuente y licencia. | Repositorio más liviano, reproducible y seguro. |

## Conclusión

El proyecto tiene una base funcional valiosa: implementa un pipeline NLP legal, incluye una versión secuencial y otra concurrente, mide métricas experimentales relevantes y contiene un modelo Promela con evidencia parcial de `errors: 0` mediante el `pan` incluido. La ejecución rápida de `pc2.go` fue exitosa y el race detector no reportó carreras en el escenario probado.

No obstante, para considerarlo una entrega técnicamente sólida, el repositorio necesita remediaciones importantes. Los GAPs críticos están en la estructura del proyecto: el módulo completo no pasa `go test ./...`, hay múltiples `main()` en el mismo paquete y los artefactos SPIN rompen el build Go. Además, la documentación no coincide completamente con los archivos reales.

La recomendación principal es reorganizar el repositorio, separar ejecutables y librerías, limpiar artefactos generados, crear pruebas automatizadas y convertir la verificación SPIN en un flujo reproducible. Con esas correcciones, el proyecto pasaría de ser una demostración funcional a un entregable defendible en términos de calidad de código, concurrencia, seguridad, rendimiento, reproducibilidad y verificación formal.
