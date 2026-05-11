import argparse
from pathlib import Path

import matplotlib.pyplot as plt
import pandas as pd


def save_line(df, x, y, title, ylabel, output_path):
    plt.figure(figsize=(9, 5))
    plt.plot(df[x], df[y], marker="o")
    plt.title(title)
    plt.xlabel("Número de workers / goroutines")
    plt.ylabel(ylabel)
    plt.grid(True, alpha=0.3)
    plt.tight_layout()
    plt.savefig(output_path, dpi=300)
    plt.close()


def main():
    parser = argparse.ArgumentParser(description="Genera gráficos del benchmark PC2")
    parser.add_argument("--raw", default="resultados/benchmark_pc2_raw.csv")
    parser.add_argument("--summary", default="resultados/benchmark_pc2_resumen.csv")
    parser.add_argument("--out-dir", default="graficos")
    args = parser.parse_args()

    raw_path = Path(args.raw)
    summary_path = Path(args.summary)
    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)

    raw = pd.read_csv(raw_path)
    summary = pd.read_csv(summary_path)

    conc = summary[summary["modo"].str.contains("concurrente", case=False, na=False)].copy()
    conc = conc.sort_values("workers")

    save_line(
        conc,
        "workers",
        "media_recortada_seg",
        "Tiempo promedio recortado según número de workers",
        "Tiempo promedio recortado (segundos)",
        out_dir / "01_tiempo_promedio_workers.png",
    )

    save_line(
        conc,
        "workers",
        "speedup",
        "Speedup frente a la versión secuencial",
        "Speedup",
        out_dir / "02_speedup_workers.png",
    )

    save_line(
        conc,
        "workers",
        "throughput_promedio",
        "Throughput del procesamiento concurrente",
        "Registros válidos procesados por segundo",
        out_dir / "03_throughput_workers.png",
    )

    save_line(
        conc,
        "workers",
        "eficiencia",
        "Eficiencia paralela según número de workers",
        "Eficiencia = speedup / workers",
        out_dir / "04_eficiencia_workers.png",
    )

    raw_conc = raw[raw["modo"].str.contains("concurrente", case=False, na=False)].copy()
    raw_conc["workers"] = raw_conc["workers"].astype(str)
    ordered = [str(w) for w in sorted(raw["workers"].unique())]
    data = [raw_conc.loc[raw_conc["workers"] == w, "total_seconds"].values for w in ordered if w in set(raw_conc["workers"])]
    labels = [w for w in ordered if w in set(raw_conc["workers"])]

    plt.figure(figsize=(10, 5))
    plt.boxplot(data, tick_labels=labels, showmeans=True)
    plt.title("Distribución de tiempos en las ejecuciones concurrentes")
    plt.xlabel("Workers / goroutines")
    plt.ylabel("Tiempo (segundos)")
    plt.grid(True, axis="y", alpha=0.3)
    plt.tight_layout()
    plt.savefig(out_dir / "05_boxplot_tiempos_workers.png", dpi=300)
    plt.close()

    exactitud = raw.groupby(["modo", "workers"], as_index=False)["resultado_correcto"].mean()
    exactitud = exactitud[exactitud["modo"].str.contains("concurrente", case=False, na=False)].sort_values("workers")
    save_line(
        exactitud,
        "workers",
        "resultado_correcto",
        "Tasa de resultados correctos por configuración concurrente",
        "Proporción de ejecuciones correctas",
        out_dir / "06_resultados_correctos_workers.png",
    )

    print(f"Gráficos generados en: {out_dir.resolve()}")


if __name__ == "__main__":
    main()
