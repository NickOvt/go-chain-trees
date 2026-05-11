#!/usr/bin/env python3
from __future__ import annotations

import argparse
import warnings
from collections import defaultdict
from pathlib import Path
from typing import Any

import matplotlib.pyplot as plt
import numpy as np
from matplotlib import MatplotlibDeprecationWarning
from matplotlib import colors as mcolors

from benchmark_visualization import (
    PRIMARY_SERIES_INDEX,
    ROOT,
    SECONDARY_SERIES_INDEX,
    TERTIARY_SERIES_INDEX,
    TREE_COLORS,
    TREE_LABELS,
    TREE_ORDER,
    add_series,
    aggregate_bucket_rows,
    aggregate_rows,
    annotate_lower_is_better,
    build_bucket_drift_summary,
    build_ordered_ratio_rows,
    build_reinsert_pair_rows,
    configure_log_x_axis,
    expand_bucket_rows,
    filter_rows,
    human_count,
    load_rows,
    ns_to_us,
    resolve_input_paths,
    save_figure,
    sorted_rows,
    tree_line_style,
    tree_scatter_style,
    write_csv,
)

FAMILY_ORDER = ["insert_only_build", "build_then_add_new", "build_then_reinsert_existing"]
FAMILY_LABELS = {
    "insert_only_build": "full build",
    "build_then_add_new": "post-build add",
    "build_then_reinsert_existing": "reinsert existing",
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate supplementary benchmark plots.")
    parser.add_argument(
        "inputs",
        nargs="*",
        help="Benchmark CSV files or timestamped benchmark directories.",
    )
    parser.add_argument(
        "--avl",
        default="total_ref_5_avl.csv",
        help="Legacy fallback AVL CSV path, relative to analysis/.",
    )
    parser.add_argument(
        "--smt",
        default="total_ref_5_smt.csv",
        help="Legacy fallback SMT CSV path, relative to analysis/.",
    )
    parser.add_argument(
        "--exclusion-avl",
        default="total_ref_7_exclusion_only_avl.csv",
        help="Fallback AVL exclusion-proof CSV path, relative to analysis/.",
    )
    parser.add_argument(
        "--exclusion-smt",
        default="total_ref_7_exclusion_only_smt.csv",
        help="Fallback SMT exclusion-proof CSV path, relative to analysis/.",
    )
    parser.add_argument(
        "--plots-dir",
        default="plots/less-important",
        help="Directory for generated supplementary plots.",
    )
    parser.add_argument(
        "--tables-dir",
        default="tables/less-important",
        help="Directory for derived supplementary tables.",
    )
    return parser.parse_args()


def throughput_from_ns(avg_ns: float | int) -> float:
    return 1_000_000_000.0 / float(avg_ns)


def safe_per_insert(numerator: float | int, row: dict[str, Any]) -> float:
    insert_count = float(row["measured_insert_scale"])
    return float(numerator) / insert_count if insert_count else float("nan")


def live_alloc_bytes_per_insert(row: dict[str, Any]) -> float:
    return safe_per_insert(float(row["insert_mem_alloc_mb"]) * 1_000_000.0, row)


def total_alloc_bytes_per_insert(row: dict[str, Any]) -> float:
    return safe_per_insert(float(row["insert_total_alloc_mb"]) * 1_000_000.0, row)


def net_live_objects_per_insert(row: dict[str, Any]) -> float:
    return safe_per_insert(float(row["net_live_heap_object_change"]), row)


def proof_size_per_million(row: dict[str, Any], proof_kind: str = "inclusion") -> float:
    scale = float(row["scale_size"])
    return float(row[f"{proof_kind}_proof_size_bytes"]) / (scale / 1_000_000.0) if scale else float("nan")


def marker_sizes(values: list[float | int], *, min_size: float = 45.0, max_size: float = 180.0) -> np.ndarray:
    array = np.asarray(values, dtype=float)
    if array.size == 0:
        return np.asarray([], dtype=float)
    if float(array.max()) == float(array.min()):
        return np.full(array.shape, (min_size + max_size) / 2.0, dtype=float)
    return min_size + (array - array.min()) / (array.max() - array.min()) * (max_size - min_size)


def pair_cross_tree_metric_rows(
    aggregated_rows: list[dict[str, Any]],
    *,
    scenario_family: str,
    metric: str,
) -> list[dict[str, Any]]:
    grouped: defaultdict[tuple[int, int, int, str], dict[str, dict[str, Any]]] = defaultdict(dict)
    for row in aggregated_rows:
        if row["is_ordered_prehashed"] or row["scenario_family"] != scenario_family:
            continue
        sample_key = str(int(row["proof_sample_pct"])) if row["proof_sample_pct"] not in ("", None) else ""
        grouped[
            (
                int(row["scale_size"]),
                int(row["measured_insert_scale"]),
                int(row["final_scale"]),
                sample_key,
            )
        ][row["tree_type"]] = row

    paired: list[dict[str, Any]] = []
    for (scale_size, measured_insert_scale, final_scale, sample_key), bundle in sorted(grouped.items(), key=lambda item: item[0]):
        if not {"avlhashtree", "smt"} <= set(bundle):
            continue
        avl_row = bundle["avlhashtree"]
        smt_row = bundle["smt"]
        paired.append(
            {
                "scenario_family": scenario_family,
                "metric": metric,
                "scale_size": scale_size,
                "scale_label": human_count(scale_size),
                "measured_insert_scale": measured_insert_scale,
                "final_scale": final_scale,
                "sample_pct": sample_key,
                "avl_value": float(avl_row[metric]),
                "smt_value": float(smt_row[metric]),
                "ratio": float(avl_row[metric]) / float(smt_row[metric]),
            }
        )
    return paired


def build_context(input_paths: list[Path]) -> tuple[list[dict[str, Any]], list[dict[str, Any]], list[dict[str, Any]], list[dict[str, Any]]]:
    raw_rows: list[dict[str, Any]] = []
    for path in input_paths:
        raw_rows.extend(load_rows(path))
    aggregated_rows, coverage_rows = aggregate_rows(raw_rows)
    bucket_rows = expand_bucket_rows(raw_rows)
    aggregated_bucket_rows = aggregate_bucket_rows(bucket_rows)
    return raw_rows, aggregated_rows, coverage_rows, aggregated_bucket_rows


def plot_cross_tree_ratio_lines(aggregated_rows: list[dict[str, Any]], plots_dir: Path) -> list[dict[str, Any]]:
    panels = [
        ("insert_only_build", "insert_avg_ns", "AVL/SMT Full-Build Insert Avg", "AVL / SMT ratio"),
        ("insert_only_build", "insert_total_ns", "AVL/SMT Full-Build Insert Total", "AVL / SMT ratio"),
        ("proof_only_after_build", "inclusion_proof_gen_avg_ns", "AVL/SMT Proof Generation", "AVL / SMT ratio"),
        ("proof_only_after_build", "inclusion_proof_verify_avg_ns", "AVL/SMT Proof Verification", "AVL / SMT ratio"),
        ("proof_only_after_build", "inclusion_proof_size_bytes", "AVL/SMT Proof Size", "AVL / SMT ratio"),
        ("build_then_add_new", "insert_avg_ns", "AVL/SMT Post-Build Add Avg", "AVL / SMT ratio"),
        ("build_then_reinsert_existing", "insert_avg_ns", "AVL/SMT Reinsert Avg", "AVL / SMT ratio"),
    ]
    all_rows: list[dict[str, Any]] = []
    fig, axes = plt.subplots(3, 3, figsize=(18, 13))
    axes_flat = list(axes.flat)
    for ax, (family, metric, title, ylabel) in zip(axes_flat, panels):
        rows = pair_cross_tree_metric_rows(aggregated_rows, scenario_family=family, metric=metric)
        all_rows.extend(rows)
        xs = [row["scale_size"] for row in rows]
        ys = [row["ratio"] for row in rows]
        ax.plot(xs, ys, color="#304d63", marker="s", linewidth=2.3, linestyle="-")
        configure_log_x_axis(ax, xs)
        ax.axhline(1.0, color="#444444", linestyle=":", linewidth=1.2)
        ax.set_title(title)
        ax.set_ylabel(ylabel)
        ax.set_xlabel("Scale")
    for ax in axes_flat[len(panels) :]:
        ax.axis("off")
    fig.suptitle("Cross-Tree Ratio Lines", y=1.01, fontsize=15)
    save_figure(fig, plots_dir, "01_cross_tree_ratio_lines")
    return all_rows


def plot_throughput_charts(aggregated_rows: list[dict[str, Any]], plots_dir: Path) -> None:
    fig, axes = plt.subplots(1, 3, figsize=(18, 5.4))

    family_specs = [
        ("insert_only_build", "Full build", PRIMARY_SERIES_INDEX),
        ("build_then_add_new", "Post-build add", SECONDARY_SERIES_INDEX),
        ("build_then_reinsert_existing", "Reinsert", TERTIARY_SERIES_INDEX),
    ]
    all_insert_x = [
        row["scale_size"]
        for row in aggregated_rows
        if row["scenario_family"] in {family for family, _, _ in family_specs} and not row["is_ordered_prehashed"]
    ]
    for family, label, series_index in family_specs:
        rows = filter_rows(aggregated_rows, scenario_family=family, is_ordered_prehashed=False)
        for tree_type in TREE_ORDER:
            series = sorted_rows([row for row in rows if row["tree_type"] == tree_type])
            axes[0].plot(
                [row["scale_size"] for row in series],
                [throughput_from_ns(row["insert_avg_ns"]) for row in series],
                label=f"{TREE_LABELS[tree_type]} {label}",
                **tree_line_style(tree_type, series_index=series_index),
            )
    configure_log_x_axis(axes[0], all_insert_x)
    axes[0].set_title("Insert Throughput")
    axes[0].set_ylabel("Ops / second")
    axes[0].set_xlabel("Scale")

    proof_rows = filter_rows(aggregated_rows, scenario_family="proof_only_after_build", is_ordered_prehashed=False)
    for tree_type in TREE_ORDER:
        series = sorted_rows([row for row in proof_rows if row["tree_type"] == tree_type])
        xs = [row["scale_size"] for row in series]
        axes[1].plot(
            xs,
            [throughput_from_ns(row["inclusion_proof_gen_avg_ns"]) for row in series],
            label=TREE_LABELS[tree_type],
            **tree_line_style(tree_type, series_index=PRIMARY_SERIES_INDEX),
        )
        axes[2].plot(
            xs,
            [throughput_from_ns(row["inclusion_proof_verify_avg_ns"]) for row in series],
            label=TREE_LABELS[tree_type],
            **tree_line_style(tree_type, series_index=PRIMARY_SERIES_INDEX),
        )
    configure_log_x_axis(axes[1], [row["scale_size"] for row in proof_rows])
    configure_log_x_axis(axes[2], [row["scale_size"] for row in proof_rows])
    axes[1].set_title("Proof Generation Throughput")
    axes[2].set_title("Proof Verification Throughput")
    for ax in axes[1:]:
        ax.set_xlabel("Scale")
        ax.set_ylabel("Proofs / second")
    axes[0].legend(loc="upper right", ncol=2)
    axes[1].legend(loc="upper right", ncol=2)
    fig.suptitle("Throughput Views", y=1.02, fontsize=15)
    save_figure(fig, plots_dir, "02_throughput_vs_scale")


def plot_normalized_memory_charts(aggregated_rows: list[dict[str, Any]], plots_dir: Path) -> None:
    fig, axes = plt.subplots(1, 3, figsize=(18, 5.4))
    family_specs = [
        ("insert_only_build", "Full build", PRIMARY_SERIES_INDEX),
        ("build_then_add_new", "Post-build add", SECONDARY_SERIES_INDEX),
    ]
    metrics = [
        (live_alloc_bytes_per_insert, "Live Alloc / Insert", "Bytes per measured insert"),
        (total_alloc_bytes_per_insert, "TotalAlloc / Insert", "Bytes per measured insert"),
        (net_live_objects_per_insert, "Net Live Objects / Insert", "Objects per measured insert"),
    ]
    all_x = [
        row["scale_size"]
        for row in aggregated_rows
        if row["scenario_family"] in {family for family, _, _ in family_specs} and not row["is_ordered_prehashed"]
    ]
    for axis_index, (calculator, title, ylabel) in enumerate(metrics):
        ax = axes[axis_index]
        for family, family_label, series_index in family_specs:
            rows = filter_rows(aggregated_rows, scenario_family=family, is_ordered_prehashed=False)
            for tree_type in TREE_ORDER:
                series = sorted_rows([row for row in rows if row["tree_type"] == tree_type])
                ax.plot(
                    [row["scale_size"] for row in series],
                    [calculator(row) for row in series],
                    label=f"{TREE_LABELS[tree_type]} {family_label}",
                    **tree_line_style(tree_type, series_index=series_index),
                )
        configure_log_x_axis(ax, all_x)
        ax.set_title(title)
        ax.set_ylabel(ylabel)
        ax.set_xlabel("Scale")
        annotate_lower_is_better(ax)
    axes[0].legend(loc="upper right", ncol=2)
    fig.suptitle("Normalized Memory Charts", y=1.02, fontsize=15)
    save_figure(fig, plots_dir, "03_normalized_memory_vs_scale")


def plot_proof_balance_lines(
    aggregated_rows: list[dict[str, Any]],
    plots_dir: Path,
    *,
    scenario_family: str = "proof_only_after_build",
    proof_kind: str = "inclusion",
    title: str = "Proof Balance and Efficiency",
    stem: str = "04_proof_balance_lines",
) -> None:
    proof_rows = filter_rows(aggregated_rows, scenario_family=scenario_family, is_ordered_prehashed=False)
    if not proof_rows:
        return
    fig, axes = plt.subplots(2, 2, figsize=(18, 11))
    balance_ax = axes[0][0]
    gen_efficiency_ax = axes[0][1]
    verify_efficiency_ax = axes[1][0]
    size_density_ax = axes[1][1]
    for tree_type in TREE_ORDER:
        series = sorted_rows([row for row in proof_rows if row["tree_type"] == tree_type])
        xs = [row["scale_size"] for row in series]
        proof_sizes = [float(row[f"{proof_kind}_proof_size_bytes"]) for row in series]
        gen_us = [ns_to_us(row[f"{proof_kind}_proof_gen_avg_ns"]) for row in series]
        verify_us = [ns_to_us(row[f"{proof_kind}_proof_verify_avg_ns"]) for row in series]
        size_per_million = [proof_size_per_million(row, proof_kind) for row in series]
        sizes = marker_sizes(xs, min_size=50.0, max_size=180.0)

        balance_ax.plot(
            xs,
            gen_us,
            label=f"{TREE_LABELS[tree_type]} generation",
            **tree_line_style(tree_type, series_index=PRIMARY_SERIES_INDEX),
        )
        balance_ax.plot(
            xs,
            verify_us,
            label=f"{TREE_LABELS[tree_type]} verification",
            **tree_line_style(tree_type, series_index=SECONDARY_SERIES_INDEX),
        )

        gen_efficiency_ax.plot(
            proof_sizes,
            gen_us,
            **tree_line_style(
                tree_type,
                series_index=PRIMARY_SERIES_INDEX,
                linewidth=1.3,
                markersize=5.5,
                alpha=0.6,
            ),
        )
        gen_efficiency_ax.scatter(
            proof_sizes,
            gen_us,
            s=sizes,
            label=TREE_LABELS[tree_type],
            **tree_scatter_style(tree_type, series_index=PRIMARY_SERIES_INDEX, alpha=0.8),
        )

        verify_efficiency_ax.plot(
            proof_sizes,
            verify_us,
            **tree_line_style(
                tree_type,
                series_index=PRIMARY_SERIES_INDEX,
                linewidth=1.3,
                markersize=5.5,
                alpha=0.6,
            ),
        )
        verify_efficiency_ax.scatter(
            proof_sizes,
            verify_us,
            s=sizes,
            **tree_scatter_style(tree_type, series_index=PRIMARY_SERIES_INDEX, alpha=0.8),
        )

        size_density_ax.plot(
            xs,
            size_per_million,
            **tree_line_style(tree_type, series_index=PRIMARY_SERIES_INDEX),
        )

    configure_log_x_axis(balance_ax, [row["scale_size"] for row in proof_rows])
    balance_ax.set_title("Proof Generation vs Proof Verification")
    balance_ax.set_xlabel("Scale")
    balance_ax.set_ylabel("Latency (µs/proof)")
    annotate_lower_is_better(balance_ax)
    balance_ax.legend(loc="upper left", ncol=1)

    gen_efficiency_ax.set_title("Proof Size vs Generation Time")
    gen_efficiency_ax.set_xlabel("Proof size (bytes)")
    gen_efficiency_ax.set_ylabel("Generation (µs/proof)")
    annotate_lower_is_better(gen_efficiency_ax)
    gen_efficiency_ax.legend(loc="upper left")

    verify_efficiency_ax.set_title("Proof Size vs Verification Time")
    verify_efficiency_ax.set_xlabel("Proof size (bytes)")
    verify_efficiency_ax.set_ylabel("Verification (µs/proof)")
    annotate_lower_is_better(verify_efficiency_ax)

    configure_log_x_axis(size_density_ax, [row["scale_size"] for row in proof_rows])
    size_density_ax.set_title("Proof Size per Million Elements")
    size_density_ax.set_xlabel("Scale")
    size_density_ax.set_ylabel("Bytes / million elements")
    annotate_lower_is_better(size_density_ax)

    fig.subplots_adjust(hspace=0.34, wspace=0.24, top=0.84)
    fig.suptitle(title, y=0.97, fontsize=15)
    save_figure(fig, plots_dir, stem)


def plot_proof_balance_ratio(aggregated_rows: list[dict[str, Any]], plots_dir: Path) -> None:
    proof_rows = filter_rows(aggregated_rows, scenario_family="proof_only_after_build", is_ordered_prehashed=False)
    fig, ax = plt.subplots(figsize=(9.5, 5.3))
    for tree_type in TREE_ORDER:
        series = sorted_rows([row for row in proof_rows if row["tree_type"] == tree_type])
        ax.plot(
            [row["scale_size"] for row in series],
            [float(row["inclusion_proof_gen_avg_ns"]) / float(row["inclusion_proof_verify_avg_ns"]) for row in series],
            label=TREE_LABELS[tree_type],
            **tree_line_style(tree_type, series_index=PRIMARY_SERIES_INDEX),
        )
    configure_log_x_axis(ax, [row["scale_size"] for row in proof_rows])
    ax.axhline(1.0, color="#444444", linestyle=":", linewidth=1.2)
    ax.set_title("Generation / Verification Ratio")
    ax.set_ylabel("Ratio")
    ax.set_xlabel("Scale")
    ax.legend(loc="upper left", ncol=2)
    save_figure(fig, plots_dir, "05_proof_balance_ratio")


def plot_proof_efficiency_charts(
    aggregated_rows: list[dict[str, Any]],
    plots_dir: Path,
    *,
    scenario_family: str = "proof_only_after_build",
    proof_kind: str = "inclusion",
    title: str = "Proof Efficiency Charts",
    stem: str = "06_proof_efficiency",
) -> None:
    proof_rows = filter_rows(aggregated_rows, scenario_family=scenario_family, is_ordered_prehashed=False)
    if not proof_rows:
        return
    fig, axes = plt.subplots(1, 3, figsize=(18, 5.6))
    for tree_type in TREE_ORDER:
        series = sorted_rows([row for row in proof_rows if row["tree_type"] == tree_type])
        sizes = marker_sizes([row["scale_size"] for row in series], min_size=50.0, max_size=180.0)
        proof_sizes = [float(row[f"{proof_kind}_proof_size_bytes"]) for row in series]
        gen_us = [ns_to_us(row[f"{proof_kind}_proof_gen_avg_ns"]) for row in series]
        verify_us = [ns_to_us(row[f"{proof_kind}_proof_verify_avg_ns"]) for row in series]
        size_per_million = [proof_size_per_million(row, proof_kind) for row in series]
        scales = [row["scale_size"] for row in series]
        axes[0].plot(
            proof_sizes,
            gen_us,
            **tree_line_style(
                tree_type,
                series_index=PRIMARY_SERIES_INDEX,
                linewidth=1.3,
                markersize=5.5,
                alpha=0.6,
            ),
        )
        axes[0].scatter(
            proof_sizes,
            gen_us,
            s=sizes,
            label=TREE_LABELS[tree_type],
            **tree_scatter_style(tree_type, series_index=PRIMARY_SERIES_INDEX, alpha=0.8),
        )
        axes[1].plot(
            proof_sizes,
            verify_us,
            **tree_line_style(
                tree_type,
                series_index=PRIMARY_SERIES_INDEX,
                linewidth=1.3,
                markersize=5.5,
                alpha=0.6,
            ),
        )
        axes[1].scatter(
            proof_sizes,
            verify_us,
            s=sizes,
            **tree_scatter_style(tree_type, series_index=PRIMARY_SERIES_INDEX, alpha=0.8),
            label=TREE_LABELS[tree_type],
        )
        axes[2].plot(
            scales,
            size_per_million,
            label=TREE_LABELS[tree_type],
            **tree_line_style(tree_type, series_index=PRIMARY_SERIES_INDEX),
        )
    axes[0].set_title("Proof Size vs Generation Time")
    axes[0].set_xlabel("Proof Size (bytes)")
    axes[0].set_ylabel("Generation (µs/proof)")
    annotate_lower_is_better(axes[0])
    axes[1].set_title("Proof Size vs Verification Time")
    axes[1].set_xlabel("Proof Size (bytes)")
    axes[1].set_ylabel("Verification (µs/proof)")
    annotate_lower_is_better(axes[1])
    configure_log_x_axis(axes[2], [row["scale_size"] for row in proof_rows])
    axes[2].set_title("Proof Size per Million Elements")
    axes[2].set_xlabel("Scale")
    axes[2].set_ylabel("Bytes / million elements")
    annotate_lower_is_better(axes[2])
    axes[0].legend(loc="upper left")
    axes[2].legend(loc="upper left")
    fig.subplots_adjust(top=0.82)
    fig.suptitle(title, y=1.02, fontsize=15)
    save_figure(fig, plots_dir, stem)


def plot_range_error_bands(aggregated_rows: list[dict[str, Any]], plots_dir: Path) -> None:
    panels = [
        ("insert_only_build", "insert_avg_ns", ns_to_us, "Full-Build Insert Latency", "Latency (µs/op)"),
        ("proof_only_after_build", "inclusion_proof_gen_avg_ns", ns_to_us, "Proof Generation", "Latency (µs/proof)"),
        ("proof_only_after_build", "inclusion_proof_verify_avg_ns", ns_to_us, "Proof Verification", "Latency (µs/proof)"),
        ("insert_only_build", "insert_total_alloc_mb", float, "Full-Build TotalAlloc", "MB"),
        ("build_then_add_new", "insert_avg_ns", ns_to_us, "Post-Build Add Latency", "Latency (µs/op)"),
    ]
    fig, axes = plt.subplots(2, 3, figsize=(18, 9))
    axes_flat = list(axes.flat)
    for ax, (family, metric, transform, title, ylabel) in zip(axes_flat, panels):
        rows = filter_rows(aggregated_rows, scenario_family=family, is_ordered_prehashed=False)
        for tree_type in TREE_ORDER:
            series = sorted_rows([row for row in rows if row["tree_type"] == tree_type])
            add_series(
                ax,
                series,
                metric,
                transform,
                TREE_LABELS[tree_type],
                tree_type,
            )
        configure_log_x_axis(ax, [row["scale_size"] for row in rows])
        ax.set_title(title)
        ax.set_xlabel("Scale")
        ax.set_ylabel(ylabel)
        annotate_lower_is_better(ax)
    axes_flat[0].legend(loc="upper left", ncol=2)
    axes_flat[-1].axis("off")
    fig.suptitle("Range/Error-Band Charts", y=1.01, fontsize=15)
    save_figure(fig, plots_dir, "07_range_error_bands")


def build_heatmap_matrix(
    aggregated_rows: list[dict[str, Any]],
    *,
    metric: str,
    tree_type: str,
    scale_order: list[int],
) -> np.ndarray:
    matrix = np.full((len(FAMILY_ORDER), len(scale_order)), np.nan, dtype=float)
    mapping = {}
    for row in aggregated_rows:
        if row["is_ordered_prehashed"] or row["tree_type"] != tree_type:
            continue
        value = row[metric]
        if value in ("", None, 0, 0.0) and metric not in {"insert_avg_ns", "insert_total_alloc_mb"}:
            continue
        mapping[(row["scenario_family"], int(row["scale_size"]))] = float(value) if value not in ("", None) else np.nan
    for family_index, family in enumerate(FAMILY_ORDER):
        for scale_index, scale in enumerate(scale_order):
            matrix[family_index, scale_index] = mapping.get((family, scale), np.nan)
    return matrix


def build_ratio_heatmap_matrix(
    aggregated_rows: list[dict[str, Any]],
    *,
    metric: str,
    scale_order: list[int],
) -> np.ndarray:
    matrix = np.full((len(FAMILY_ORDER), len(scale_order)), np.nan, dtype=float)
    mapping: defaultdict[tuple[str, int], dict[str, float]] = defaultdict(dict)
    for row in aggregated_rows:
        if row["is_ordered_prehashed"]:
            continue
        value = row[metric]
        if value in ("", None, 0, 0.0):
            continue
        mapping[(row["scenario_family"], int(row["scale_size"]))][row["tree_type"]] = float(value)
    for family_index, family in enumerate(FAMILY_ORDER):
        for scale_index, scale in enumerate(scale_order):
            bundle = mapping.get((family, scale), {})
            if {"avlhashtree", "smt"} <= set(bundle):
                matrix[family_index, scale_index] = bundle["avlhashtree"] / bundle["smt"]
    return matrix


def draw_heatmap(
    ax: plt.Axes,
    matrix: np.ndarray,
    *,
    x_labels: list[str],
    title: str,
    cmap: str,
    norm: mcolors.Normalize,
    colorbar_label: str,
) -> None:
    masked = np.ma.masked_invalid(matrix)
    ax.grid(False)
    image = ax.imshow(masked, aspect="auto", cmap=cmap, norm=norm)
    ax.set_title(title)
    ax.set_xticks(range(len(x_labels)))
    ax.set_xticklabels(x_labels, rotation=35, ha="right")
    ax.set_yticks(range(len(FAMILY_ORDER)))
    ax.set_yticklabels([FAMILY_LABELS[family] for family in FAMILY_ORDER])
    with warnings.catch_warnings():
        warnings.filterwarnings(
            "ignore",
            message="Auto-removal of grids by pcolor\\(\\) and pcolormesh\\(\\) is deprecated.*",
            category=MatplotlibDeprecationWarning,
        )
        colorbar = plt.colorbar(image, ax=ax, fraction=0.046, pad=0.04)
    colorbar.set_label(colorbar_label)


def plot_heatmaps(aggregated_rows: list[dict[str, Any]], plots_dir: Path) -> None:
    scale_order = sorted({int(row["scale_size"]) for row in aggregated_rows if not row["is_ordered_prehashed"]})
    x_labels = [human_count(scale) for scale in scale_order]

    avl_insert = build_heatmap_matrix(aggregated_rows, metric="insert_avg_ns", tree_type="avlhashtree", scale_order=scale_order)
    smt_insert = build_heatmap_matrix(aggregated_rows, metric="insert_avg_ns", tree_type="smt", scale_order=scale_order)
    insert_ratio = build_ratio_heatmap_matrix(aggregated_rows, metric="insert_avg_ns", scale_order=scale_order)
    avl_alloc = build_heatmap_matrix(aggregated_rows, metric="insert_total_alloc_mb", tree_type="avlhashtree", scale_order=scale_order)
    smt_alloc = build_heatmap_matrix(aggregated_rows, metric="insert_total_alloc_mb", tree_type="smt", scale_order=scale_order)
    alloc_ratio = build_ratio_heatmap_matrix(aggregated_rows, metric="insert_total_alloc_mb", scale_order=scale_order)

    valid_insert = np.concatenate([arr[~np.isnan(arr)] for arr in [avl_insert, smt_insert] if np.any(~np.isnan(arr))])
    valid_alloc = np.concatenate([arr[~np.isnan(arr)] for arr in [avl_alloc, smt_alloc] if np.any(~np.isnan(arr))])
    insert_ratio_values = insert_ratio[~np.isnan(insert_ratio)]
    alloc_ratio_values = alloc_ratio[~np.isnan(alloc_ratio)]
    if not len(valid_insert) or not len(valid_alloc) or not len(insert_ratio_values) or not len(alloc_ratio_values):
        return

    fig, axes = plt.subplots(2, 3, figsize=(20, 9.5))
    draw_heatmap(
        axes[0][0],
        avl_insert,
        x_labels=x_labels,
        title="AVL Insert Latency Heatmap",
        cmap="YlOrBr",
        norm=mcolors.LogNorm(vmin=float(valid_insert.min()), vmax=float(valid_insert.max())),
        colorbar_label="Avg insert ns",
    )
    annotate_lower_is_better(axes[0][0])
    draw_heatmap(
        axes[0][1],
        smt_insert,
        x_labels=x_labels,
        title="SMT Insert Latency Heatmap",
        cmap="YlOrBr",
        norm=mcolors.LogNorm(vmin=float(valid_insert.min()), vmax=float(valid_insert.max())),
        colorbar_label="Avg insert ns",
    )
    annotate_lower_is_better(axes[0][1])
    draw_heatmap(
        axes[0][2],
        insert_ratio,
        x_labels=x_labels,
        title="AVL/SMT Insert Ratio Heatmap",
        cmap="coolwarm_r",
        norm=mcolors.TwoSlopeNorm(
            vmin=min(0.3, float(insert_ratio_values.min())),
            vcenter=1.0,
            vmax=max(1.2, float(insert_ratio_values.max())),
        ),
        colorbar_label="AVL / SMT ratio",
    )
    draw_heatmap(
        axes[1][0],
        avl_alloc,
        x_labels=x_labels,
        title="AVL TotalAlloc Heatmap",
        cmap="YlGnBu",
        norm=mcolors.LogNorm(vmin=float(valid_alloc.min()), vmax=float(valid_alloc.max())),
        colorbar_label="TotalAlloc MB",
    )
    annotate_lower_is_better(axes[1][0])
    draw_heatmap(
        axes[1][1],
        smt_alloc,
        x_labels=x_labels,
        title="SMT TotalAlloc Heatmap",
        cmap="YlGnBu",
        norm=mcolors.LogNorm(vmin=float(valid_alloc.min()), vmax=float(valid_alloc.max())),
        colorbar_label="TotalAlloc MB",
    )
    annotate_lower_is_better(axes[1][1])
    draw_heatmap(
        axes[1][2],
        alloc_ratio,
        x_labels=x_labels,
        title="AVL/SMT TotalAlloc Ratio Heatmap",
        cmap="coolwarm_r",
        norm=mcolors.TwoSlopeNorm(
            vmin=min(0.3, float(alloc_ratio_values.min())),
            vcenter=1.0,
            vmax=max(1.2, float(alloc_ratio_values.max())),
        ),
        colorbar_label="AVL / SMT ratio",
    )
    fig.suptitle("Heatmaps: Latency, Allocation, and Cross-Tree Ratios", y=1.01, fontsize=15)
    save_figure(fig, plots_dir, "08_heatmaps_latency_memory_ratio")


def plot_drift_summary_bars(bucket_rows: list[dict[str, Any]], plots_dir: Path) -> list[dict[str, Any]]:
    summary_rows = build_bucket_drift_summary(bucket_rows)
    fig, axes = plt.subplots(1, 2, figsize=(18, 5.6), sharey=False)
    panel_specs = [
        ("insert_only_build", False, "Default Full-Build Insert Drift"),
        ("insert_only_build", True, "Ordered Full-Build Insert Drift"),
    ]
    selected_rows: list[dict[str, Any]] = []
    for ax, (family, ordered_flag, title) in zip(axes, panel_specs):
        rows = [
            row
            for row in summary_rows
            if row["scenario_family"] == family and row["is_ordered_prehashed"] == ordered_flag
        ]
        selected_rows.extend(rows)
        scales = sorted({int(row["scale_size"]) for row in rows})
        positions = np.arange(len(scales))
        width = 0.36
        for offset_index, tree_type in enumerate(TREE_ORDER):
            tree_rows = {int(row["scale_size"]): row for row in rows if row["tree_type"] == tree_type}
            ys = [tree_rows[scale]["last_over_first_ratio"] for scale in scales]
            ax.bar(
                positions + (offset_index - 0.5) * width,
                ys,
                width=width,
                color=TREE_COLORS[tree_type],
                label=TREE_LABELS[tree_type],
            )
        ax.axhline(1.0, color="#444444", linestyle=":", linewidth=1.2)
        ax.set_xticks(positions)
        ax.set_xticklabels([human_count(scale) for scale in scales], rotation=35, ha="right")
        ax.set_title(title)
        ax.set_ylabel("Last / first ratio")
        ax.set_xlabel("Scale")
    axes[0].legend(loc="upper left", ncol=2)
    fig.suptitle("Bucket Drift Summary Bars", y=1.02, fontsize=15)
    save_figure(fig, plots_dir, "09_drift_summary_bars")
    return selected_rows


def plot_ordered_input_delta_bars(aggregated_rows: list[dict[str, Any]], plots_dir: Path) -> list[dict[str, Any]]:
    ratio_rows = build_ordered_ratio_rows(aggregated_rows)
    preferred_panels = [
        ("insert_only_build", "insert_avg_ns", "Full-Build Insert Latency Improvement"),
        ("proof_only_after_build", "inclusion_proof_gen_avg_ns", "Proof Generation Improvement"),
        ("proof_only_after_build", "inclusion_proof_verify_avg_ns", "Proof Verification Improvement"),
        ("proof_only_after_build", "inclusion_proof_size_bytes", "Proof Size Improvement"),
        ("insert_only_build", "insert_total_alloc_mb", "Full-Build TotalAlloc Improvement"),
        ("build_then_add_new", "insert_avg_ns", "Post-Build Add Latency Improvement"),
        ("build_then_reinsert_existing", "insert_avg_ns", "Reinsert Latency Improvement"),
    ]
    metric_titles: list[tuple[str, str, str]] = []
    for panel in preferred_panels:
        family, metric, _ = panel
        rows = [row for row in ratio_rows if row["scenario_family"] == family and row["metric"] == metric]
        scale_sets = [
            {int(row["scale_size"]) for row in rows if row["tree_type"] == tree_type}
            for tree_type in TREE_ORDER
        ]
        shared_scales = sorted(set.intersection(*scale_sets)) if all(scale_sets) else []
        if shared_scales:
            metric_titles.append(panel)
        if len(metric_titles) == 4:
            break
    if not metric_titles:
        return []

    fig, axes = plt.subplots(2, 2, figsize=(18, 10))
    selected_rows: list[dict[str, Any]] = []
    axes_flat = list(axes.flat)
    for ax, (family, metric, title) in zip(axes_flat, metric_titles):
        rows = [row for row in ratio_rows if row["scenario_family"] == family and row["metric"] == metric]
        tree_rows_by_type = {
            tree_type: {int(row["scale_size"]): row for row in rows if row["tree_type"] == tree_type}
            for tree_type in TREE_ORDER
        }
        scales = sorted(set.intersection(*(set(tree_rows) for tree_rows in tree_rows_by_type.values())))
        selected_rows.extend([row for row in rows if int(row["scale_size"]) in scales])
        positions = np.arange(len(scales))
        width = 0.36
        for offset_index, tree_type in enumerate(TREE_ORDER):
            tree_rows = tree_rows_by_type[tree_type]
            ys = [(1.0 - tree_rows[scale]["ordered_over_default_ratio"]) * 100.0 for scale in scales]
            ax.bar(
                positions + (offset_index - 0.5) * width,
                ys,
                width=width,
                color=TREE_COLORS[tree_type],
                label=TREE_LABELS[tree_type],
            )
        ax.axhline(0.0, color="#444444", linestyle=":", linewidth=1.2)
        ax.set_xticks(positions)
        ax.set_xticklabels([human_count(scale) for scale in scales], rotation=35, ha="right")
        ax.set_title(title)
        ax.set_ylabel("Percent improvement")
        ax.set_xlabel("Scale")
    for ax in axes_flat[len(metric_titles) :]:
        ax.axis("off")
    axes[0][0].legend(loc="upper right", ncol=2)
    fig.suptitle("Ordered-Input Delta Bars", y=1.01, fontsize=15)
    save_figure(fig, plots_dir, "10_ordered_input_delta_bars")
    return selected_rows


def plot_reinsert_delta_bars(aggregated_rows: list[dict[str, Any]], plots_dir: Path) -> list[dict[str, Any]]:
    pair_rows = build_reinsert_pair_rows(aggregated_rows)
    fig, axes = plt.subplots(1, 2, figsize=(18, 5.4), sharey=True)
    for ax, tree_type in zip(axes, TREE_ORDER):
        tree_rows = [row for row in pair_rows if row["tree_type"] == tree_type]
        positions = np.arange(len(tree_rows))
        savings = [(1.0 - row["reinsert_to_new_ratio"]) * 100.0 for row in tree_rows]
        ax.bar(positions, savings, color=TREE_COLORS[tree_type])
        ax.axhline(0.0, color="#444444", linestyle=":", linewidth=1.2)
        ax.set_xticks(positions)
        ax.set_xticklabels([human_count(row["prebuild_elements"]) for row in tree_rows], rotation=35, ha="right")
        ax.set_title(TREE_LABELS[tree_type])
        ax.set_ylabel("Percent savings")
        ax.set_xlabel("Prebuild scale")
    fig.suptitle("Reinsert Delta Bars", y=1.02, fontsize=15)
    save_figure(fig, plots_dir, "11_reinsert_delta_bars")
    return pair_rows


def plot_postbuild_sensitivity(aggregated_rows: list[dict[str, Any]], plots_dir: Path) -> None:
    add_rows = filter_rows(aggregated_rows, scenario_family="build_then_add_new", is_ordered_prehashed=False)
    fig, axes = plt.subplots(1, 3, figsize=(18, 5.4))
    for tree_type in TREE_ORDER:
        series = sorted_rows([row for row in add_rows if row["tree_type"] == tree_type])
        xs = [row["scale_size"] for row in series]
        axes[0].plot(
            xs,
            [ns_to_us(row["insert_avg_ns"]) for row in series],
            label=TREE_LABELS[tree_type],
            **tree_line_style(tree_type, series_index=PRIMARY_SERIES_INDEX),
        )
        axes[1].plot(
            xs,
            [float(row["insert_total_alloc_mb"]) for row in series],
            label=TREE_LABELS[tree_type],
            **tree_line_style(tree_type, series_index=PRIMARY_SERIES_INDEX),
        )
        axes[2].plot(
            xs,
            [float(row["net_live_heap_object_change"]) for row in series],
            label=TREE_LABELS[tree_type],
            **tree_line_style(tree_type, series_index=PRIMARY_SERIES_INDEX),
        )
    all_x = [row["scale_size"] for row in add_rows]
    for ax in axes:
        configure_log_x_axis(ax, all_x)
        ax.set_xlabel("Prebuild scale")
    axes[0].set_title("Post-Build Avg Insert Latency")
    axes[0].set_ylabel("Latency (µs/op)")
    annotate_lower_is_better(axes[0])
    axes[1].set_title("Post-Build TotalAlloc")
    axes[1].set_ylabel("MB")
    annotate_lower_is_better(axes[1])
    axes[2].set_title("Post-Build Net Live Objects")
    axes[2].set_ylabel("Objects")
    annotate_lower_is_better(axes[2])
    axes[0].legend(loc="upper left", ncol=2)
    fig.suptitle("Post-Build Sensitivity Charts", y=1.02, fontsize=15)
    save_figure(fig, plots_dir, "12_postbuild_sensitivity")


def plot_tradeoff_scatter(aggregated_rows: list[dict[str, Any]], plots_dir: Path) -> None:
    fig, axes = plt.subplots(1, 2, figsize=(16, 6.2))
    panel_specs = [
        ("insert_only_build", "Full-build tradeoff"),
        ("build_then_add_new", "Post-build add tradeoff"),
    ]
    for ax, (family, title) in zip(axes, panel_specs):
        rows = filter_rows(aggregated_rows, scenario_family=family, is_ordered_prehashed=False)
        for tree_type in TREE_ORDER:
            tree_rows = [row for row in rows if row["tree_type"] == tree_type]
            tree_sizes = marker_sizes([row["scale_size"] for row in tree_rows], min_size=40.0, max_size=220.0)
            ax.scatter(
                [ns_to_us(row["insert_avg_ns"]) for row in tree_rows],
                [float(row["insert_total_alloc_mb"]) for row in tree_rows],
                s=tree_sizes,
                label=TREE_LABELS[tree_type],
                **tree_scatter_style(tree_type, series_index=PRIMARY_SERIES_INDEX, alpha=0.75),
            )
        ax.set_title(title)
        ax.set_xlabel("Avg insert latency (µs/op)")
        ax.set_ylabel("TotalAlloc MB")
        ax.set_xscale("log")
        ax.set_yscale("log")
        annotate_lower_is_better(ax)
    axes[0].legend(loc="upper left", ncol=2)
    fig.suptitle("Tradeoff Scatter Plots", y=1.01, fontsize=15)
    save_figure(fig, plots_dir, "13_tradeoff_scatter")


def pareto_mask(x_values: np.ndarray, y_values: np.ndarray) -> np.ndarray:
    mask = np.ones(len(x_values), dtype=bool)
    for index in range(len(x_values)):
        dominates = (
            (x_values <= x_values[index])
            & (y_values <= y_values[index])
            & ((x_values < x_values[index]) | (y_values < y_values[index]))
        )
        dominates[index] = False
        if dominates.any():
            mask[index] = False
    return mask


def build_insert_proof_tradeoff_rows(aggregated_rows: list[dict[str, Any]]) -> list[dict[str, Any]]:
    insert_map = {}
    proof_map = {}
    for row in aggregated_rows:
        if row["is_ordered_prehashed"]:
            continue
        key = (row["tree_type"], int(row["scale_size"]))
        if row["scenario_family"] == "insert_only_build":
            insert_map[key] = row
        if row["scenario_family"] == "proof_only_after_build":
            proof_map[key] = row
    rows: list[dict[str, Any]] = []
    for key in sorted(set(insert_map) & set(proof_map), key=lambda item: (TREE_ORDER.index(item[0]), item[1])):
        insert_row = insert_map[key]
        proof_row = proof_map[key]
        rows.append(
            {
                "tree_type": key[0],
                "scale_size": key[1],
                "scale_label": human_count(key[1]),
                "insert_avg_us": ns_to_us(insert_row["insert_avg_ns"]),
                "proof_size_bytes": float(proof_row["inclusion_proof_size_bytes"]),
            }
        )
    return rows


def draw_pareto_panel(
    ax: plt.Axes,
    rows: list[dict[str, Any]],
    *,
    x_key: str,
    y_key: str,
    title: str,
    x_label: str,
    y_label: str,
    x_log: bool = False,
    y_log: bool = False,
) -> None:
    if not rows:
        return
    x_values = np.asarray([row[x_key] for row in rows], dtype=float)
    y_values = np.asarray([row[y_key] for row in rows], dtype=float)
    mask = pareto_mask(x_values, y_values)
    pareto_rows = [row for row, keep in zip(rows, mask) if keep]
    pareto_rows = sorted(pareto_rows, key=lambda row: row[x_key])
    for tree_type in TREE_ORDER:
        tree_rows = [row for row in rows if row["tree_type"] == tree_type]
        sizes = marker_sizes([row["scale_size"] for row in tree_rows], min_size=50.0, max_size=200.0)
        ax.scatter(
            [row[x_key] for row in tree_rows],
            [row[y_key] for row in tree_rows],
            s=sizes,
            label=TREE_LABELS[tree_type],
            **tree_scatter_style(tree_type, series_index=PRIMARY_SERIES_INDEX, alpha=0.78),
        )
    if pareto_rows:
        ax.plot(
            [row[x_key] for row in pareto_rows],
            [row[y_key] for row in pareto_rows],
            color="#111111",
            linewidth=1.6,
            linestyle="--",
            label="Pareto front",
        )
        ax.scatter(
            [row[x_key] for row in pareto_rows],
            [row[y_key] for row in pareto_rows],
            s=210,
            facecolors="none",
            edgecolors="#111111",
            linewidths=1.5,
        )
    if x_log:
        ax.set_xscale("log")
    if y_log:
        ax.set_yscale("log")
    ax.set_title(title)
    ax.set_xlabel(x_label)
    ax.set_ylabel(y_label)
    annotate_lower_is_better(ax)


def plot_pareto_fronts(aggregated_rows: list[dict[str, Any]], plots_dir: Path) -> None:
    insert_rows = filter_rows(aggregated_rows, scenario_family="insert_only_build", is_ordered_prehashed=False)
    latency_memory_rows = [
        {
            "tree_type": row["tree_type"],
            "scale_size": int(row["scale_size"]),
            "scale_label": row["scale_label"],
            "insert_avg_us": ns_to_us(row["insert_avg_ns"]),
            "total_alloc_mb": float(row["insert_total_alloc_mb"]),
        }
        for row in insert_rows
    ]
    insert_proof_rows = build_insert_proof_tradeoff_rows(aggregated_rows)

    fig, axes = plt.subplots(1, 2, figsize=(16, 6.2))
    draw_pareto_panel(
        axes[0],
        insert_proof_rows,
        x_key="insert_avg_us",
        y_key="proof_size_bytes",
        title="Pareto: Insert Latency vs Proof Size",
        x_label="Insert latency (µs/op)",
        y_label="Proof size (bytes)",
    )
    draw_pareto_panel(
        axes[1],
        latency_memory_rows,
        x_key="insert_avg_us",
        y_key="total_alloc_mb",
        title="Pareto: Insert Latency vs TotalAlloc",
        x_label="Insert latency (µs/op)",
        y_label="TotalAlloc MB",
        x_log=True,
        y_log=True,
    )
    axes[0].legend(loc="upper left", ncol=2)
    fig.suptitle("Pareto Charts", y=1.01, fontsize=15)
    save_figure(fig, plots_dir, "14_pareto_fronts")


def main() -> None:
    args = parse_args()
    input_paths = resolve_input_paths(args)
    if not input_paths:
        raise SystemExit(
            "No benchmark CSV inputs were found. Pass one or more CSV files or timestamped directories when data is available."
        )

    plots_dir = (ROOT / args.plots_dir).resolve()
    tables_dir = (ROOT / args.tables_dir).resolve()
    plots_dir.mkdir(parents=True, exist_ok=True)
    tables_dir.mkdir(parents=True, exist_ok=True)

    raw_rows, aggregated_rows, coverage_rows, aggregated_bucket_rows = build_context(input_paths)

    ratio_rows = plot_cross_tree_ratio_lines(aggregated_rows, plots_dir)
    plot_throughput_charts(aggregated_rows, plots_dir)
    plot_normalized_memory_charts(aggregated_rows, plots_dir)
    plot_proof_balance_lines(aggregated_rows, plots_dir)
    plot_proof_balance_ratio(aggregated_rows, plots_dir)
    plot_proof_efficiency_charts(aggregated_rows, plots_dir)
    plot_proof_balance_lines(
        aggregated_rows,
        plots_dir,
        scenario_family="exclusion_proof_only_after_build",
        proof_kind="exclusion",
        title="Exclusion Proof Balance and Efficiency",
        stem="15_exclusion_proof_balance_lines",
    )
    plot_proof_efficiency_charts(
        aggregated_rows,
        plots_dir,
        scenario_family="exclusion_proof_only_after_build",
        proof_kind="exclusion",
        title="Exclusion Proof Efficiency Charts",
        stem="16_exclusion_proof_efficiency",
    )
    plot_range_error_bands(aggregated_rows, plots_dir)
    plot_heatmaps(aggregated_rows, plots_dir)
    drift_rows = plot_drift_summary_bars(aggregated_bucket_rows, plots_dir)
    ordered_delta_rows = plot_ordered_input_delta_bars(aggregated_rows, plots_dir)
    reinsert_rows = plot_reinsert_delta_bars(aggregated_rows, plots_dir)
    plot_postbuild_sensitivity(aggregated_rows, plots_dir)
    plot_tradeoff_scatter(aggregated_rows, plots_dir)
    plot_pareto_fronts(aggregated_rows, plots_dir)

    write_csv(tables_dir / "cross_tree_ratio_lines.csv", ratio_rows)
    write_csv(tables_dir / "ordered_input_delta.csv", ordered_delta_rows)
    write_csv(tables_dir / "reinsert_delta.csv", reinsert_rows)
    write_csv(tables_dir / "drift_summary.csv", drift_rows)
    write_csv(tables_dir / "coverage_rows.csv", coverage_rows)
    write_csv(tables_dir / "aggregated_bucket_rows.csv", aggregated_bucket_rows)

    plot_count = len(list(plots_dir.glob("*.png")))
    print(f"Generated supplementary plots in {plots_dir}")
    print(f"Supplementary tables written to {tables_dir}")
    print(f"PNG plot count: {plot_count}")
    print(f"Aggregated scenario rows available: {len(aggregated_rows)}")
    print(f"Raw rows loaded: {len(raw_rows)}")


if __name__ == "__main__":
    main()
