#!/usr/bin/env python3
from __future__ import annotations

import argparse
import csv
import os
import re
from collections import defaultdict
from dataclasses import dataclass
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parent
os.environ.setdefault("MPLCONFIGDIR", str(ROOT / ".mplconfig"))
(ROOT / ".mplconfig").mkdir(exist_ok=True)

import matplotlib

matplotlib.use("Agg")
import matplotlib.pyplot as plt
import numpy as np

plt.rcParams.update(
    {
        "figure.facecolor": "#fcfbf8",
        "axes.facecolor": "#fcfbf8",
        "axes.edgecolor": "#3b3b3b",
        "axes.grid": True,
        "grid.alpha": 0.22,
        "grid.color": "#808080",
        "axes.spines.top": False,
        "axes.spines.right": False,
        "font.size": 11,
        "axes.titlesize": 13,
        "axes.labelsize": 11,
        "legend.frameon": False,
        "figure.autolayout": False,
    }
)

TREE_ORDER = ["avlhashtree", "smt"]
TREE_LABELS = {"avlhashtree": "AVLHashTree", "smt": "SMT"}
TREE_COLORS = {"avlhashtree": "#1f5aa6", "smt": "#b5482d"}
PRIMARY_SERIES_INDEX = 0
SECONDARY_SERIES_INDEX = 1
TERTIARY_SERIES_INDEX = 2
TREE_MARKERS = {"avlhashtree": "s", "smt": "s"}
SERIES_MARKERS = {
    PRIMARY_SERIES_INDEX: "s",
    SECONDARY_SERIES_INDEX: "o",
    TERTIARY_SERIES_INDEX: "o",
}
SERIES_LINESTYLES = {
    PRIMARY_SERIES_INDEX: "-",
    SECONDARY_SERIES_INDEX: "--",
    TERTIARY_SERIES_INDEX: ":",
}
ORDER_LABELS = {False: "Random inserts", True: "Ordered inserts"}
ORDER_LINESTYLES = {
    False: SERIES_LINESTYLES[PRIMARY_SERIES_INDEX],
    True: SERIES_LINESTYLES[SECONDARY_SERIES_INDEX],
}

TIMESTAMP_DIR_RE = re.compile(r"\d{2}-\d{2}-\d{4}-\d{2}-\d{2}-\d{2}")
BUCKET_RE = re.compile(
    r"(?P<start>\d+)-(?P<end>\d+)%=(?P<value>n/a|[0-9]+(?:\.[0-9]+)?(?:ns|µs|ms|s))(?:\(n=(?P<count>\d+)\))?"
)
DURATION_RE = re.compile(r"([0-9]+(?:\.[0-9]+)?)(ns|µs|ms|s)$")


@dataclass(frozen=True)
class ColumnSpec:
    raw_name: str
    clean_name: str
    kind: str
    required: bool = True
    default: str = ""


COLUMN_SPECS = [
    ColumnSpec("TreeType", "tree_type", "text"),
    ColumnSpec("Scenario", "raw_scenario", "text"),
    ColumnSpec("PrebuildElements", "prebuild_elements", "int"),
    ColumnSpec("FinalElements", "final_elements", "int"),
    ColumnSpec("InsertElements", "insert_elements", "int"),
    ColumnSpec("InsertTime(ns)", "insert_total_ns", "int"),
    ColumnSpec("AvgPerBlock(ns)", "insert_avg_ns", "int"),
    ColumnSpec("InsertTimeBuckets", "insert_buckets", "text"),
    ColumnSpec("MemAllocMB", "insert_mem_alloc_mb", "float"),
    ColumnSpec("TotalAllocMB", "insert_total_alloc_mb", "float"),
    ColumnSpec("HeapObjects", "insert_heap_objects", "int"),
    ColumnSpec("CreatedHeapObjects", "created_heap_objects", "int", required=False, default="0"),
    ColumnSpec("FreedHeapObjects", "freed_heap_objects", "int", required=False, default="0"),
    ColumnSpec("NetLiveHeapObjectChange", "net_live_heap_object_change", "int", required=False, default="0"),
    ColumnSpec("InclusionProofTotal(ns)", "inclusion_proof_total_ns", "int"),
    ColumnSpec("InclusionProofGen(ns)", "inclusion_proof_gen_avg_ns", "int"),
    ColumnSpec("InclusionProofSize(bytes)", "inclusion_proof_size_bytes", "int"),
    ColumnSpec("InclusionProofVerify(ns)", "inclusion_proof_verify_avg_ns", "int"),
    ColumnSpec("ExclusionProofTotal(ns)", "exclusion_proof_total_ns", "int"),
    ColumnSpec("ExclusionProofGen(ns)", "exclusion_proof_gen_avg_ns", "int"),
    ColumnSpec("ExclusionProofSize(bytes)", "exclusion_proof_size_bytes", "int"),
    ColumnSpec("ExclusionProofVerify(ns)", "exclusion_proof_verify_avg_ns", "int"),
    ColumnSpec("DeleteElements", "delete_elements", "int"),
    ColumnSpec("DeleteTime(ns)", "delete_total_ns", "int"),
    ColumnSpec("AvgDeletePerBlock(ns)", "delete_avg_ns", "int"),
    ColumnSpec("DeleteMemAllocMB", "delete_mem_alloc_mb", "float"),
    ColumnSpec("DeleteTotalAllocMB", "delete_total_alloc_mb", "float"),
    ColumnSpec("DeleteHeapObjects", "delete_heap_objects", "int"),
    ColumnSpec("DeletesCreatedHeapObjects", "deletes_created_heap_objects", "int", required=False, default="0"),
    ColumnSpec("DeletesFreedHeapObjects", "deletes_freed_heap_objects", "int", required=False, default="0"),
    ColumnSpec("DeletesNetLiveHeapObjectChange", "deletes_net_live_heap_object_change", "int", required=False, default="0"),
]

LEGACY_IGNORED_COLUMNS = {
    "InclusionProofGenBuckets",
    "ExclusionProofGenBuckets",
}

INT_FIELDS = {spec.clean_name for spec in COLUMN_SPECS if spec.kind == "int"}
FLOAT_FIELDS = {spec.clean_name for spec in COLUMN_SPECS if spec.kind == "float"}
TEXT_FIELDS = {spec.clean_name for spec in COLUMN_SPECS if spec.kind == "text"}

NUMERIC_METRICS = [
    "insert_total_ns",
    "insert_avg_ns",
    "insert_mem_alloc_mb",
    "insert_total_alloc_mb",
    "insert_heap_objects",
    "created_heap_objects",
    "freed_heap_objects",
    "net_live_heap_object_change",
    "inclusion_proof_total_ns",
    "inclusion_proof_gen_avg_ns",
    "inclusion_proof_size_bytes",
    "inclusion_proof_verify_avg_ns",
    "exclusion_proof_total_ns",
    "exclusion_proof_gen_avg_ns",
    "exclusion_proof_size_bytes",
    "exclusion_proof_verify_avg_ns",
    "delete_total_ns",
    "delete_avg_ns",
    "delete_mem_alloc_mb",
    "delete_total_alloc_mb",
    "delete_heap_objects",
    "deletes_created_heap_objects",
    "deletes_freed_heap_objects",
    "deletes_net_live_heap_object_change",
]

SCENARIO_PATTERNS = [
    ("insert_only_build", re.compile(r"insert_only(?:_build)?_(\d+[km])$")),
    ("proof_only_after_build", re.compile(r"proof_only_after_(\d+[km])_build_sample_(\d+)pct$")),
    ("exclusion_proof_only_after_build", re.compile(r"exclusion_proof_only_after_(\d+[km])_build_sample_(\d+)pct$")),
    ("build_then_add_new", re.compile(r"build_(\d+[km])_then_add_(\d+[km])_new$")),
    ("build_then_reinsert_existing", re.compile(r"build_(\d+[km])_then_reinsert_(\d+[km])_existing$")),
    ("legacy_build", re.compile(r"build_(\d+[km])$")),
]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate benchmark plots and summary tables.")
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
    parser.add_argument("--plots-dir", default="plots", help="Directory for generated plots.")
    parser.add_argument("--tables-dir", default="tables", help="Directory for generated tables.")
    parser.add_argument(
        "--report",
        default="benchmark_analysis.md",
        help="Path to the generated markdown analysis report.",
    )
    return parser.parse_args()


def resolve_user_path(raw_path: str) -> Path:
    candidate = Path(raw_path)
    return candidate if candidate.is_absolute() else (ROOT / candidate).resolve()


def relative_display_path(path: Path) -> str:
    try:
        return str(path.relative_to(ROOT))
    except ValueError:
        return str(path)


def parse_size_token(token: str | None) -> int | None:
    if token is None:
        return None
    match = re.fullmatch(r"(\d+)([km])", token)
    if not match:
        return None
    value = int(match.group(1))
    suffix = match.group(2)
    return value * (1_000 if suffix == "k" else 1_000_000)


def format_scale(value: int | float | None) -> str:
    if value is None:
        return "n/a"
    rounded = int(round(float(value)))
    if rounded >= 1_000_000 and rounded % 1_000_000 == 0:
        return f"{rounded // 1_000_000}m"
    if rounded >= 1_000 and rounded % 1_000 == 0:
        return f"{rounded // 1_000}k"
    return str(rounded)


def human_count(value: int | float | None) -> str:
    if value is None:
        return "n/a"
    rounded = int(round(float(value)))
    if rounded >= 1_000_000:
        if rounded % 1_000_000 == 0:
            return f"{rounded // 1_000_000}M"
        return f"{rounded / 1_000_000:.1f}M"
    if rounded >= 1_000:
        if rounded % 1_000 == 0:
            return f"{rounded // 1_000}K"
        return f"{rounded / 1_000:.1f}K"
    return str(rounded)


def cast_metric_value(metric: str, value: float) -> int | float:
    if metric in INT_FIELDS:
        return int(round(value))
    return float(value)


def to_number(kind: str, raw_value: str) -> int | float:
    if raw_value == "":
        raw_value = "0"
    if kind == "int":
        return int(float(raw_value))
    if kind == "float":
        return float(raw_value)
    raise ValueError(f"Unexpected numeric kind: {kind}")


def parse_duration_to_ns(value: str) -> float | None:
    if value == "n/a":
        return None
    match = DURATION_RE.fullmatch(value)
    if not match:
        return None
    amount = float(match.group(1))
    unit = match.group(2)
    multipliers = {"ns": 1.0, "µs": 1_000.0, "ms": 1_000_000.0, "s": 1_000_000_000.0}
    return amount * multipliers[unit]


def ns_to_us(value: float | int) -> float:
    return float(value) / 1_000.0


def ns_to_ms(value: float | int) -> float:
    return float(value) / 1_000_000.0


def ns_to_seconds(value: float | int) -> float:
    return float(value) / 1_000_000_000.0


def quantiles(values: list[float | int]) -> tuple[float, float, float]:
    array = np.asarray(values, dtype=float)
    return float(np.median(array)), float(np.percentile(array, 10)), float(np.percentile(array, 90))


def parse_bucket_string(bucket_string: str) -> list[dict[str, Any]]:
    entries: list[dict[str, Any]] = []
    if not bucket_string:
        return entries
    for segment in bucket_string.split("|"):
        match = BUCKET_RE.fullmatch(segment)
        if not match:
            continue
        duration_ns = parse_duration_to_ns(match.group("value"))
        sample_count = int(match.group("count")) if match.group("count") else 0
        start_pct = int(match.group("start"))
        end_pct = int(match.group("end"))
        entries.append(
            {
                "bucket_start_pct": start_pct,
                "bucket_end_pct": end_pct,
                "bucket_mid_pct": (start_pct + end_pct) / 2.0,
                "avg_duration_ns": duration_ns,
                "sample_count": sample_count,
            }
        )
    return entries


def discover_timestamp_csvs(root: Path) -> list[Path]:
    csv_paths: list[Path] = []
    for candidate in sorted(root.iterdir()):
        if not candidate.is_dir() or not TIMESTAMP_DIR_RE.fullmatch(candidate.name):
            continue
        csv_path = candidate / f"{candidate.name}.csv"
        if csv_path.exists():
            csv_paths.append(csv_path.resolve())
    return csv_paths


def expand_input_path(path: Path) -> list[Path]:
    if path.is_file():
        return [path.resolve()]
    if not path.exists():
        raise FileNotFoundError(f"Input path does not exist: {path}")
    if TIMESTAMP_DIR_RE.fullmatch(path.name):
        csv_path = path / f"{path.name}.csv"
        if csv_path.exists():
            return [csv_path.resolve()]
    discovered = discover_timestamp_csvs(path)
    if discovered:
        return discovered
    csv_paths = sorted(candidate.resolve() for candidate in path.glob("*.csv"))
    if csv_paths:
        return csv_paths
    raise FileNotFoundError(f"No benchmark CSVs found under: {path}")


def resolve_input_paths(args: argparse.Namespace) -> list[Path]:
    if args.inputs:
        expanded: list[Path] = []
        for raw_path in args.inputs:
            expanded.extend(expand_input_path(resolve_user_path(raw_path)))
        return sorted(dict.fromkeys(expanded))

    timestamp_csvs = discover_timestamp_csvs(ROOT)
    if timestamp_csvs:
        return timestamp_csvs

    legacy_paths = []
    for raw_path in [args.avl, args.smt, args.exclusion_avl, args.exclusion_smt]:
        path = resolve_user_path(raw_path)
        if path.exists():
            legacy_paths.append(path)
    return sorted(dict.fromkeys(legacy_paths))


def extract_run_id(csv_path: Path) -> str:
    if TIMESTAMP_DIR_RE.fullmatch(csv_path.parent.name) and csv_path.stem == csv_path.parent.name:
        return csv_path.parent.name
    return csv_path.stem


def validate_header(csv_path: Path, fieldnames: list[str] | None) -> tuple[str, list[str]]:
    if not fieldnames:
        raise ValueError(f"{csv_path} has no CSV header.")
    header_set = set(fieldnames)
    missing_required = [spec.raw_name for spec in COLUMN_SPECS if spec.required and spec.raw_name not in header_set]
    if missing_required:
        raise ValueError(
            f"{csv_path} is missing required benchmark columns: {', '.join(missing_required)}"
        )
    missing_optional = [spec.raw_name for spec in COLUMN_SPECS if not spec.required and spec.raw_name not in header_set]
    schema_variant = "current" if not missing_optional else "legacy_without_object_churn_columns"
    return schema_variant, missing_optional


def parse_scenario(row: dict[str, Any]) -> dict[str, Any]:
    raw_scenario = row["raw_scenario"]
    ordered_prefix = "prehashed_ordered_"
    is_ordered_prehashed = raw_scenario.startswith(ordered_prefix)
    working = raw_scenario[len(ordered_prefix) :] if is_ordered_prehashed else raw_scenario

    scenario_family = "unknown"
    scenario_family_detailed = "unknown"
    declared_primary_size: int | None = None
    declared_secondary_size: int | None = None
    proof_sample_pct: float | None = None
    issue_notes: list[str] = []

    for family, pattern in SCENARIO_PATTERNS:
        match = pattern.fullmatch(working)
        if not match:
            continue
        scenario_family = family
        declared_primary_size = parse_size_token(match.group(1))
        if family in {"proof_only_after_build", "exclusion_proof_only_after_build"}:
            proof_sample_pct = float(match.group(2))
        if family in {"build_then_add_new", "build_then_reinsert_existing"}:
            declared_secondary_size = parse_size_token(match.group(2))
        break

    if scenario_family == "unknown":
        issue_notes.append("scenario_name_parse_failed")

    if scenario_family == "insert_only_build":
        scenario_family_detailed = "prehashed_ordered_insert_only" if is_ordered_prehashed else "insert_only_build"
        scale_size = row["final_elements"]
        expected_primary_size = row["final_elements"]
        expected_secondary_size = None
    elif scenario_family == "proof_only_after_build":
        scenario_family_detailed = "ordered_prehashed_proof_only" if is_ordered_prehashed else "proof_only_after_build"
        scale_size = row["final_elements"]
        expected_primary_size = row["final_elements"]
        expected_secondary_size = None
    elif scenario_family == "exclusion_proof_only_after_build":
        scenario_family_detailed = (
            "ordered_prehashed_exclusion_proof_only" if is_ordered_prehashed else "exclusion_proof_only_after_build"
        )
        scale_size = row["final_elements"]
        expected_primary_size = row["final_elements"]
        expected_secondary_size = None
    elif scenario_family == "build_then_add_new":
        scenario_family_detailed = "ordered_prehashed_postbuild_add" if is_ordered_prehashed else "build_then_add_new"
        scale_size = row["prebuild_elements"]
        expected_primary_size = row["prebuild_elements"]
        expected_secondary_size = row["insert_elements"]
    elif scenario_family == "build_then_reinsert_existing":
        scenario_family_detailed = "ordered_prehashed_reinsert" if is_ordered_prehashed else "build_then_reinsert_existing"
        scale_size = row["prebuild_elements"]
        expected_primary_size = row["prebuild_elements"]
        expected_secondary_size = row["insert_elements"]
    elif scenario_family == "legacy_build":
        scenario_family_detailed = "legacy_build"
        scale_size = row["final_elements"]
        expected_primary_size = row["final_elements"]
        expected_secondary_size = None
        issue_notes.append("legacy_build_row")
    else:
        scale_size = row["final_elements"]
        expected_primary_size = None
        expected_secondary_size = None

    if declared_primary_size is not None and expected_primary_size is not None and declared_primary_size != expected_primary_size:
        issue_notes.append(
            f"declared_primary_size_{format_scale(declared_primary_size)}_mismatch_numeric_{format_scale(expected_primary_size)}"
        )
    if (
        declared_secondary_size is not None
        and expected_secondary_size is not None
        and declared_secondary_size != expected_secondary_size
    ):
        issue_notes.append(
            f"declared_secondary_size_{format_scale(declared_secondary_size)}_mismatch_numeric_{format_scale(expected_secondary_size)}"
        )
    if scenario_family in {"proof_only_after_build", "exclusion_proof_only_after_build"}:
        if row["prebuild_elements"] != row["final_elements"]:
            issue_notes.append("proof_only_prebuild_final_mismatch")
        if row["insert_elements"] != 0 and row["insert_total_ns"] > 0:
            issue_notes.append("proof_only_has_measured_insert_metrics")

    canonical_prefix = "prehashed_ordered_" if is_ordered_prehashed else ""
    if scenario_family == "insert_only_build":
        if is_ordered_prehashed:
            canonical_scenario = f"prehashed_ordered_insert_only_{format_scale(row['final_elements'])}"
        else:
            canonical_scenario = f"insert_only_build_{format_scale(row['final_elements'])}"
    elif scenario_family == "proof_only_after_build":
        pct_label = (
            f"{int(proof_sample_pct)}pct"
            if proof_sample_pct is not None and proof_sample_pct.is_integer()
            else "sample_unknown"
        )
        if is_ordered_prehashed:
            canonical_scenario = f"prehashed_ordered_proof_only_after_{format_scale(row['final_elements'])}_build_sample_{pct_label}"
        else:
            canonical_scenario = f"proof_only_after_{format_scale(row['final_elements'])}_build_sample_{pct_label}"
    elif scenario_family == "exclusion_proof_only_after_build":
        pct_label = (
            f"{int(proof_sample_pct)}pct"
            if proof_sample_pct is not None and proof_sample_pct.is_integer()
            else "sample_unknown"
        )
        if is_ordered_prehashed:
            canonical_scenario = (
                f"prehashed_ordered_exclusion_proof_only_after_{format_scale(row['final_elements'])}_build_sample_{pct_label}"
            )
        else:
            canonical_scenario = f"exclusion_proof_only_after_{format_scale(row['final_elements'])}_build_sample_{pct_label}"
    elif scenario_family == "build_then_add_new":
        canonical_scenario = (
            f"{canonical_prefix}build_{format_scale(row['prebuild_elements'])}_then_add_"
            f"{format_scale(row['insert_elements'])}_new"
        )
    elif scenario_family == "build_then_reinsert_existing":
        canonical_scenario = (
            f"{canonical_prefix}build_{format_scale(row['prebuild_elements'])}_then_reinsert_"
            f"{format_scale(row['insert_elements'])}_existing"
        )
    elif scenario_family == "legacy_build":
        canonical_scenario = f"build_{format_scale(row['final_elements'])}"
    else:
        canonical_scenario = raw_scenario

    if raw_scenario != canonical_scenario:
        issue_notes.append("raw_scenario_normalized_from_numeric_columns")

    return {
        "scenario_family": scenario_family,
        "scenario_family_detailed": scenario_family_detailed,
        "is_ordered_prehashed": is_ordered_prehashed,
        "is_prehashed_ordered": is_ordered_prehashed,
        "is_insert_only": scenario_family == "insert_only_build",
        "is_proof_only": scenario_family in {"proof_only_after_build", "exclusion_proof_only_after_build"},
        "is_exclusion_proof_only": scenario_family == "exclusion_proof_only_after_build",
        "is_postbuild_add": scenario_family == "build_then_add_new",
        "is_postbuild_add_new": scenario_family == "build_then_add_new",
        "is_reinsert_existing": scenario_family == "build_then_reinsert_existing",
        "proof_sample_pct": proof_sample_pct,
        "sample_pct": proof_sample_pct if proof_sample_pct is not None else "",
        "prebuild_scale": row["prebuild_elements"],
        "measured_insert_scale": row["insert_elements"],
        "final_scale": row["final_elements"],
        "scale_size": scale_size,
        "declared_primary_size": declared_primary_size,
        "declared_secondary_size": declared_secondary_size,
        "canonical_scenario": canonical_scenario,
        "raw_scenario_matches_canonical": raw_scenario == canonical_scenario,
        "scenario_issue": "; ".join(issue_notes),
        "scale_label": human_count(scale_size),
    }


def load_rows(csv_path: Path) -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []
    repeat_counter: defaultdict[str, int] = defaultdict(int)

    with csv_path.open("r", newline="", encoding="utf-8") as handle:
        reader = csv.DictReader(handle)
        schema_variant, missing_optional = validate_header(csv_path, reader.fieldnames)
        for row_index, raw_row in enumerate(reader, start=1):
            clean_row: dict[str, Any] = {
                "source_file": relative_display_path(csv_path),
                "run_id": extract_run_id(csv_path),
                "row_number": row_index,
                "schema_variant": schema_variant,
                "schema_missing_optional_columns": "|".join(missing_optional),
            }
            for spec in COLUMN_SPECS:
                raw_value = raw_row.get(spec.raw_name, spec.default)
                if spec.kind == "text":
                    clean_row[spec.clean_name] = raw_value
                else:
                    clean_row[spec.clean_name] = to_number(spec.kind, raw_value)
            repeat_counter[clean_row["raw_scenario"]] += 1
            clean_row["repeat_index_within_raw_scenario"] = repeat_counter[clean_row["raw_scenario"]]
            clean_row.update(parse_scenario(clean_row))
            rows.append(clean_row)
    return rows


def aggregate_rows(rows: list[dict[str, Any]]) -> tuple[list[dict[str, Any]], list[dict[str, Any]]]:
    grouped: defaultdict[tuple[str, str], list[dict[str, Any]]] = defaultdict(list)
    for row in rows:
        grouped[(row["tree_type"], row["canonical_scenario"])].append(row)

    aggregated_rows: list[dict[str, Any]] = []
    coverage_rows: list[dict[str, Any]] = []
    passthrough_fields = [
        "tree_type",
        "scenario_family",
        "scenario_family_detailed",
        "canonical_scenario",
        "is_ordered_prehashed",
        "is_prehashed_ordered",
        "is_insert_only",
        "is_proof_only",
        "is_exclusion_proof_only",
        "is_postbuild_add",
        "is_postbuild_add_new",
        "is_reinsert_existing",
        "proof_sample_pct",
        "sample_pct",
        "prebuild_elements",
        "final_elements",
        "insert_elements",
        "prebuild_scale",
        "measured_insert_scale",
        "final_scale",
        "scale_size",
        "scale_label",
    ]

    sort_key = lambda item: (
        TREE_ORDER.index(item[0][0]) if item[0][0] in TREE_ORDER else len(TREE_ORDER),
        item[1][0]["scale_size"],
        item[1][0]["canonical_scenario"],
    )

    for (_, _), group_rows in sorted(grouped.items(), key=sort_key):
        first = group_rows[0]
        raw_scenarios = sorted({row["raw_scenario"] for row in group_rows})
        run_ids = sorted({row["run_id"] for row in group_rows})
        notes = {note for row in group_rows for note in row["scenario_issue"].split("; ") if note}
        schema_variants = sorted({row["schema_variant"] for row in group_rows})
        missing_optional = sorted(
            {
                chunk
                for row in group_rows
                for chunk in row["schema_missing_optional_columns"].split("|")
                if chunk
            }
        )
        if len(raw_scenarios) > 1:
            notes.add("multiple_raw_scenarios_collapsed")
        if len(run_ids) != len(group_rows):
            notes.add("multiple_rows_within_run")
        if missing_optional:
            notes.add("legacy_schema_missing_optional_columns")

        aggregated: dict[str, Any] = {field: first[field] for field in passthrough_fields}
        aggregated["raw_scenarios"] = "|".join(raw_scenarios)
        aggregated["run_ids"] = "|".join(run_ids)
        aggregated["run_count"] = len(group_rows)
        aggregated["coverage_note"] = "; ".join(sorted(notes))
        aggregated["schema_variants"] = "|".join(schema_variants)
        aggregated["missing_optional_columns"] = "|".join(missing_optional)
        aggregated["source_files"] = "|".join(sorted({row["source_file"] for row in group_rows}))

        for metric in NUMERIC_METRICS:
            values = [float(row[metric]) for row in group_rows if row[metric] is not None]
            if not values:
                aggregated[metric] = ""
                aggregated[f"{metric}_p10"] = ""
                aggregated[f"{metric}_p90"] = ""
                continue
            median_value, p10_value, p90_value = quantiles(values)
            aggregated[metric] = cast_metric_value(metric, median_value)
            aggregated[f"{metric}_p10"] = cast_metric_value(metric, p10_value if len(values) > 1 else median_value)
            aggregated[f"{metric}_p90"] = cast_metric_value(metric, p90_value if len(values) > 1 else median_value)

        aggregated_rows.append(aggregated)
        coverage_rows.append(
            {
                "tree_type": first["tree_type"],
                "canonical_scenario": first["canonical_scenario"],
                "scenario_family": first["scenario_family"],
                "scenario_family_detailed": first["scenario_family_detailed"],
                "is_ordered_prehashed": first["is_ordered_prehashed"],
                "scale_label": first["scale_label"],
                "run_count": len(group_rows),
                "run_ids": "|".join(run_ids),
                "raw_scenarios": "|".join(raw_scenarios),
                "schema_variants": "|".join(schema_variants),
                "missing_optional_columns": "|".join(missing_optional),
                "coverage_note": "; ".join(sorted(notes)),
            }
        )
    return aggregated_rows, coverage_rows


def expand_bucket_rows(rows: list[dict[str, Any]]) -> list[dict[str, Any]]:
    expanded: list[dict[str, Any]] = []
    for row in rows:
        for bucket in parse_bucket_string(row["insert_buckets"]):
            expanded.append(
                {
                    "tree_type": row["tree_type"],
                    "raw_scenario": row["raw_scenario"],
                    "canonical_scenario": row["canonical_scenario"],
                    "scenario_family": row["scenario_family"],
                    "scenario_family_detailed": row["scenario_family_detailed"],
                    "is_ordered_prehashed": row["is_ordered_prehashed"],
                    "run_id": row["run_id"],
                    "source_file": row["source_file"],
                    "scale_size": row["scale_size"],
                    "scale_label": row["scale_label"],
                    "final_scale": row["final_scale"],
                    "prebuild_scale": row["prebuild_scale"],
                    "measured_insert_scale": row["measured_insert_scale"],
                    "bucket_start_pct": bucket["bucket_start_pct"],
                    "bucket_end_pct": bucket["bucket_end_pct"],
                    "bucket_mid_pct": bucket["bucket_mid_pct"],
                    "avg_duration_ns": bucket["avg_duration_ns"],
                    "sample_count": bucket["sample_count"],
                }
            )
    return expanded


def aggregate_bucket_rows(bucket_rows: list[dict[str, Any]]) -> list[dict[str, Any]]:
    grouped: defaultdict[tuple[Any, ...], list[dict[str, Any]]] = defaultdict(list)
    for row in bucket_rows:
        key = (
            row["tree_type"],
            row["canonical_scenario"],
            row["bucket_start_pct"],
            row["bucket_end_pct"],
        )
        grouped[key].append(row)

    aggregated: list[dict[str, Any]] = []
    sort_key = lambda item: (
        TREE_ORDER.index(item[0][0]) if item[0][0] in TREE_ORDER else len(TREE_ORDER),
        item[1][0]["scale_size"],
        item[0][2],
    )

    for (_, _, _, _), group_rows in sorted(grouped.items(), key=sort_key):
        first = group_rows[0]
        duration_values = [row["avg_duration_ns"] for row in group_rows if row["avg_duration_ns"] is not None]
        count_values = [row["sample_count"] for row in group_rows]
        duration_median, duration_p10, duration_p90 = quantiles(duration_values)
        count_median, count_p10, count_p90 = quantiles(count_values)
        aggregated.append(
            {
                "tree_type": first["tree_type"],
                "canonical_scenario": first["canonical_scenario"],
                "scenario_family": first["scenario_family"],
                "scenario_family_detailed": first["scenario_family_detailed"],
                "is_ordered_prehashed": first["is_ordered_prehashed"],
                "scale_size": first["scale_size"],
                "scale_label": first["scale_label"],
                "final_scale": first["final_scale"],
                "prebuild_scale": first["prebuild_scale"],
                "measured_insert_scale": first["measured_insert_scale"],
                "bucket_start_pct": first["bucket_start_pct"],
                "bucket_end_pct": first["bucket_end_pct"],
                "bucket_mid_pct": first["bucket_mid_pct"],
                "run_count": len(group_rows),
                "avg_duration_ns": float(duration_median),
                "avg_duration_ns_p10": float(duration_p10 if len(duration_values) > 1 else duration_median),
                "avg_duration_ns_p90": float(duration_p90 if len(duration_values) > 1 else duration_median),
                "sample_count": int(round(count_median)),
                "sample_count_p10": int(round(count_p10 if len(count_values) > 1 else count_median)),
                "sample_count_p90": int(round(count_p90 if len(count_values) > 1 else count_median)),
            }
        )
    return aggregated


def write_csv(path: Path, rows: list[dict[str, Any]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    if not rows:
        path.write_text("", encoding="utf-8")
        return
    fieldnames = list(rows[0].keys())
    with path.open("w", newline="", encoding="utf-8") as handle:
        writer = csv.DictWriter(handle, fieldnames=fieldnames)
        writer.writeheader()
        for row in rows:
            writer.writerow(row)


def prepare_output_dirs(plots_dir: Path, tables_dir: Path) -> None:
    plots_dir.mkdir(parents=True, exist_ok=True)
    tables_dir.mkdir(parents=True, exist_ok=True)


def configure_log_x_axis(ax: plt.Axes, x_values: list[float | int]) -> None:
    ticks = sorted({int(round(float(value))) for value in x_values if value})
    if not ticks:
        return
    ax.set_xscale("log")
    ax.set_yscale("linear")
    ax.set_xticks(ticks)
    ax.set_xticklabels([human_count(value) for value in ticks], rotation=35, ha="right")


def series_marker(series_index: int = PRIMARY_SERIES_INDEX) -> str:
    return SERIES_MARKERS.get(series_index, SERIES_MARKERS[SECONDARY_SERIES_INDEX])


def series_linestyle(series_index: int = PRIMARY_SERIES_INDEX) -> str:
    return SERIES_LINESTYLES.get(series_index, SERIES_LINESTYLES[SECONDARY_SERIES_INDEX])


def tree_line_style(
    tree_type: str,
    *,
    series_index: int = PRIMARY_SERIES_INDEX,
    linewidth: float = 2.2,
    markersize: float = 5.5,
    alpha: float | None = None,
) -> dict[str, Any]:
    style: dict[str, Any] = {
        "color": TREE_COLORS[tree_type],
        "marker": series_marker(series_index),
        "linestyle": series_linestyle(series_index),
        "linewidth": linewidth,
        "markersize": markersize,
    }
    if alpha is not None:
        style["alpha"] = alpha
    return style


def tree_scatter_style(
    tree_type: str,
    *,
    series_index: int = PRIMARY_SERIES_INDEX,
    alpha: float | None = None,
) -> dict[str, Any]:
    style: dict[str, Any] = {
        "color": TREE_COLORS[tree_type],
        "marker": series_marker(series_index),
    }
    if alpha is not None:
        style["alpha"] = alpha
    return style


def add_series(
    ax: plt.Axes,
    rows: list[dict[str, Any]],
    metric: str,
    unit_transform,
    label: str,
    tree_type: str,
    *,
    series_index: int = PRIMARY_SERIES_INDEX,
) -> None:
    if not rows:
        return
    xs = [row["scale_size"] for row in rows]
    ys = [unit_transform(row[metric]) for row in rows]
    ax.plot(xs, ys, label=label, **tree_line_style(tree_type, series_index=series_index))
    if any(int(row.get("run_count", 1)) > 1 for row in rows):
        y_p10 = [unit_transform(row[f"{metric}_p10"]) for row in rows]
        y_p90 = [unit_transform(row[f"{metric}_p90"]) for row in rows]
        ax.fill_between(xs, y_p10, y_p90, color=TREE_COLORS[tree_type], alpha=0.12)


def annotate_lower_is_better(ax: plt.Axes) -> None:
    xlabel = ax.get_xlabel()
    suffix = "(Lower is better)"
    if suffix in xlabel:
        ax.set_xlabel(xlabel.replace(suffix, "").strip())
    top_axis = ax.secondary_xaxis("top")
    top_axis.set_xlabel(suffix)
    top_axis.set_xticks([])
    top_axis.tick_params(axis="x", which="both", length=0, labeltop=False)
    top_axis.spines["top"].set_visible(False)


def save_figure(fig: plt.Figure, plots_dir: Path, stem: str) -> None:
    fig.savefig(plots_dir / f"{stem}.png", dpi=220, bbox_inches="tight")
    fig.savefig(plots_dir / f"{stem}.svg", bbox_inches="tight")
    plt.close(fig)


def filter_rows(
    rows: list[dict[str, Any]],
    *,
    scenario_family: str | None = None,
    is_ordered_prehashed: bool | None = None,
    tree_type: str | None = None,
) -> list[dict[str, Any]]:
    filtered = rows
    if scenario_family is not None:
        filtered = [row for row in filtered if row["scenario_family"] == scenario_family]
    if is_ordered_prehashed is not None:
        filtered = [row for row in filtered if row["is_ordered_prehashed"] == is_ordered_prehashed]
    if tree_type is not None:
        filtered = [row for row in filtered if row["tree_type"] == tree_type]
    return filtered


def sorted_rows(rows: list[dict[str, Any]]) -> list[dict[str, Any]]:
    return sorted(rows, key=lambda row: (row["scale_size"], row["measured_insert_scale"], row["final_scale"]))


def make_insert_latency_plot(aggregated_rows: list[dict[str, Any]], plots_dir: Path) -> None:
    rows = filter_rows(aggregated_rows, scenario_family="insert_only_build")
    if not rows:
        return

    fig, axes = plt.subplots(2, 2, figsize=(16, 10), sharex=True, sharey=True)
    all_x = [row["scale_size"] for row in rows]
    panel_specs = [
        (
            "SMT: Ordered vs Random Inserts",
            [
                ("smt", False, ORDER_LABELS[False]),
                ("smt", True, ORDER_LABELS[True]),
            ],
        ),
        (
            "AVLHashTree: Ordered vs Random Inserts",
            [
                ("avlhashtree", False, ORDER_LABELS[False]),
                ("avlhashtree", True, ORDER_LABELS[True]),
            ],
        ),
        (
            "Random Inserts: SMT vs AVLHashTree",
            [
                ("smt", False, TREE_LABELS["smt"]),
                ("avlhashtree", False, TREE_LABELS["avlhashtree"]),
            ],
        ),
        (
            "Ordered Inserts: SMT vs AVLHashTree",
            [
                ("smt", True, TREE_LABELS["smt"]),
                ("avlhashtree", True, TREE_LABELS["avlhashtree"]),
            ],
        ),
    ]

    for ax, (title, series_specs) in zip(list(axes.flat), panel_specs):
        for tree_type, is_ordered_prehashed, label in series_specs:
            series = sorted_rows(
                [
                    row
                    for row in rows
                    if row["tree_type"] == tree_type and row["is_ordered_prehashed"] == is_ordered_prehashed
                ]
            )
            add_series(
                ax,
                series,
                "insert_avg_ns",
                ns_to_us,
                label,
                tree_type,
                series_index=SECONDARY_SERIES_INDEX if is_ordered_prehashed else PRIMARY_SERIES_INDEX,
            )
        configure_log_x_axis(ax, all_x)
        ax.set_title(title)
        ax.set_xlabel("Final elements")
        ax.set_ylabel("Average Insert Latency (µs/op)")
        annotate_lower_is_better(ax)
        ax.legend(loc="upper left")

    fig.suptitle("Insert Latency vs Scale", y=1.01, fontsize=15)
    save_figure(fig, plots_dir, "01_insert_avg_latency_vs_scale")


def make_insert_total_plot(aggregated_rows: list[dict[str, Any]], plots_dir: Path) -> None:
    rows = filter_rows(aggregated_rows, scenario_family="insert_only_build", is_ordered_prehashed=False)
    fig, ax = plt.subplots(figsize=(8.6, 5.4))
    all_x = [row["scale_size"] for row in rows]
    for tree_type in TREE_ORDER:
        series = sorted_rows([row for row in rows if row["tree_type"] == tree_type])
        add_series(
            ax,
            series,
            "insert_total_ns",
            ns_to_seconds,
            TREE_LABELS[tree_type],
            tree_type,
        )
    configure_log_x_axis(ax, all_x)
    ax.set_title("Total Insert Time vs Final Scale")
    ax.set_ylabel("Total Insert Time (s)")
    ax.set_xlabel("Final elements")
    annotate_lower_is_better(ax)
    ax.legend(loc="upper left", ncol=2)
    save_figure(fig, plots_dir, "02_insert_total_time_vs_scale")


def make_proof_metric_plot(
    aggregated_rows: list[dict[str, Any]],
    plots_dir: Path,
    *,
    metric: str,
    ylabel: str,
    stem: str,
    title: str,
    transform,
) -> None:
    rows = filter_rows(aggregated_rows, scenario_family="proof_only_after_build", is_ordered_prehashed=False)
    fig, ax = plt.subplots(figsize=(8.6, 5.4))
    all_x = [row["scale_size"] for row in rows]
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
    configure_log_x_axis(ax, all_x)
    ax.set_title(title)
    ax.set_ylabel(ylabel)
    ax.set_xlabel("Final elements")
    annotate_lower_is_better(ax)
    ax.legend(loc="upper left", ncol=2)
    save_figure(fig, plots_dir, stem)


def make_proof_scaling_plot(
    aggregated_rows: list[dict[str, Any]],
    plots_dir: Path,
    *,
    scenario_family: str = "proof_only_after_build",
    proof_kind: str = "inclusion",
    title: str = "Inclusion Proof Scaling (Proof-Only Scenarios)",
    stem: str = "03_inclusion_proof_scaling",
) -> None:
    rows = filter_rows(aggregated_rows, scenario_family=scenario_family, is_ordered_prehashed=False)
    if not rows:
        return
    fig, axes = plt.subplots(1, 3, figsize=(18, 5.4))
    all_x = [row["scale_size"] for row in rows]
    panels = [
        (f"{proof_kind}_proof_gen_avg_ns", "Proof Generation (µs)", ns_to_us),
        (f"{proof_kind}_proof_verify_avg_ns", "Proof Verification (µs)", ns_to_us),
        (f"{proof_kind}_proof_size_bytes", "Proof Size (bytes)", float),
    ]
    for ax, (metric, ylabel, transform) in zip(axes, panels):
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
        configure_log_x_axis(ax, all_x)
        ax.set_xlabel("Final elements")
        ax.set_ylabel(ylabel)
        ax.set_title(ylabel.replace(" (", "\n("))
        annotate_lower_is_better(ax)
    axes[0].legend(loc="upper left", ncol=2)
    fig.subplots_adjust(top=0.82)
    fig.suptitle(title, y=1.02, fontsize=15)
    save_figure(fig, plots_dir, stem)


def make_memory_plot(aggregated_rows: list[dict[str, Any]], plots_dir: Path) -> None:
    rows = filter_rows(aggregated_rows, scenario_family="insert_only_build", is_ordered_prehashed=False)
    fig, axes = plt.subplots(2, 3, figsize=(18, 9), sharex=True)
    panels = [
        ("insert_mem_alloc_mb", "Alloc delta (MB)"),
        ("insert_total_alloc_mb", "TotalAlloc delta (MB)"),
        ("created_heap_objects", "Created heap objects"),
        ("freed_heap_objects", "Freed heap objects"),
        ("net_live_heap_object_change", "Net live heap object change"),
        ("insert_heap_objects", "Heap objects after phase"),
    ]
    all_x = [row["scale_size"] for row in rows]
    for ax, (metric, ylabel) in zip(list(axes.flat), panels):
        for tree_type in TREE_ORDER:
            series = sorted_rows([row for row in rows if row["tree_type"] == tree_type])
            add_series(
                ax,
                series,
                metric,
                float,
                TREE_LABELS[tree_type],
                tree_type,
            )
        configure_log_x_axis(ax, all_x)
        ax.set_ylabel(ylabel)
        ax.set_xlabel("Final elements")
        annotate_lower_is_better(ax)
    axes[0][0].legend(loc="upper left", ncol=2)
    fig.suptitle("Insert-Phase Memory and Heap-Object Metrics", y=1.01, fontsize=15)
    save_figure(fig, plots_dir, "04_memory_vs_scale")


def build_reinsert_pair_rows(aggregated_rows: list[dict[str, Any]], *, is_ordered_prehashed: bool = False) -> list[dict[str, Any]]:
    paired: defaultdict[tuple[str, int, int], dict[str, dict[str, Any]]] = defaultdict(dict)
    for row in aggregated_rows:
        if row["is_ordered_prehashed"] != is_ordered_prehashed:
            continue
        pair_key = (row["tree_type"], int(row["prebuild_scale"]), int(row["measured_insert_scale"]))
        if row["scenario_family"] == "build_then_add_new":
            paired[pair_key]["new"] = row
        if row["scenario_family"] == "build_then_reinsert_existing":
            paired[pair_key]["reinsert"] = row

    pair_rows: list[dict[str, Any]] = []
    for (tree_type, prebuild, insert_count), bundle in sorted(
        paired.items(),
        key=lambda item: (
            TREE_ORDER.index(item[0][0]) if item[0][0] in TREE_ORDER else len(TREE_ORDER),
            item[0][1],
        ),
    ):
        if "new" not in bundle or "reinsert" not in bundle:
            continue
        new_row = bundle["new"]
        reinsert_row = bundle["reinsert"]
        pair_rows.append(
            {
                "tree_type": tree_type,
                "is_ordered_prehashed": is_ordered_prehashed,
                "prebuild_elements": prebuild,
                "insert_elements": insert_count,
                "scale_size": prebuild,
                "scale_label": human_count(prebuild),
                "new_insert_avg_ns": new_row["insert_avg_ns"],
                "reinsert_avg_ns": reinsert_row["insert_avg_ns"],
                "reinsert_to_new_ratio": float(reinsert_row["insert_avg_ns"]) / float(new_row["insert_avg_ns"]),
                "new_insert_total_alloc_mb": new_row["insert_total_alloc_mb"],
                "reinsert_total_alloc_mb": reinsert_row["insert_total_alloc_mb"],
                "new_net_live_heap_object_change": new_row["net_live_heap_object_change"],
                "reinsert_net_live_heap_object_change": reinsert_row["net_live_heap_object_change"],
            }
        )
    return pair_rows


def make_reinsert_plot(aggregated_rows: list[dict[str, Any]], plots_dir: Path) -> list[dict[str, Any]]:
    pair_rows = build_reinsert_pair_rows(aggregated_rows, is_ordered_prehashed=False)
    fig, axes = plt.subplots(1, 3, figsize=(18, 5.4))
    for tree_type in TREE_ORDER:
        tree_rows = [row for row in pair_rows if row["tree_type"] == tree_type]
        xs = [row["prebuild_elements"] for row in tree_rows]
        axes[0].plot(
            xs,
            [ns_to_us(row["new_insert_avg_ns"]) for row in tree_rows],
            label=f"{TREE_LABELS[tree_type]} add new",
            **tree_line_style(tree_type, series_index=PRIMARY_SERIES_INDEX),
        )
        axes[0].plot(
            xs,
            [ns_to_us(row["reinsert_avg_ns"]) for row in tree_rows],
            label=f"{TREE_LABELS[tree_type]} update",
            **tree_line_style(tree_type, series_index=SECONDARY_SERIES_INDEX),
        )
        axes[1].plot(
            xs,
            [row["reinsert_to_new_ratio"] for row in tree_rows],
            label=TREE_LABELS[tree_type],
            **tree_line_style(tree_type, series_index=PRIMARY_SERIES_INDEX),
        )
        axes[2].plot(
            xs,
            [float(row["new_insert_total_alloc_mb"]) for row in tree_rows],
            label=f"{TREE_LABELS[tree_type]} add new",
            **tree_line_style(tree_type, series_index=PRIMARY_SERIES_INDEX),
        )
        axes[2].plot(
            xs,
            [float(row["reinsert_total_alloc_mb"]) for row in tree_rows],
            label=f"{TREE_LABELS[tree_type]} update",
            **tree_line_style(tree_type, series_index=SECONDARY_SERIES_INDEX),
        )

    common_x = [row["prebuild_elements"] for row in pair_rows]
    for ax in axes:
        configure_log_x_axis(ax, common_x)
        ax.set_xlabel("Prebuilt elements")
    axes[0].set_title("Average latency")
    axes[0].set_ylabel("Latency (µs/op)")
    annotate_lower_is_better(axes[0])
    axes[1].set_title("Update / add new ratio")
    axes[1].set_ylabel("Ratio")
    axes[1].axhline(1.0, color="#444444", linestyle=":", linewidth=1.2)
    axes[2].set_title("TotalAlloc delta")
    axes[2].set_ylabel("MB")
    annotate_lower_is_better(axes[2])
    axes[0].legend(loc="upper left", ncol=2)
    fig.suptitle("Existing Element Updates vs Post Build Add New", y=1.02, fontsize=15)
    save_figure(fig, plots_dir, "05_update_vs_new")
    return pair_rows


def scenario_pair_key(row: dict[str, Any]) -> tuple[Any, ...]:
    sample_key = int(row["proof_sample_pct"]) if row["proof_sample_pct"] not in ("", None) else ""
    return (
        row["tree_type"],
        row["scenario_family"],
        int(row["prebuild_scale"]),
        int(row["measured_insert_scale"]),
        int(row["final_scale"]),
        sample_key,
    )


def build_ordered_ratio_rows(aggregated_rows: list[dict[str, Any]]) -> list[dict[str, Any]]:
    default_rows: dict[tuple[Any, ...], dict[str, Any]] = {}
    ordered_rows: dict[tuple[Any, ...], dict[str, Any]] = {}
    metric_sets = {
        "insert_only_build": ["insert_avg_ns", "insert_total_ns", "insert_total_alloc_mb"],
        "proof_only_after_build": [
            "inclusion_proof_gen_avg_ns",
            "inclusion_proof_verify_avg_ns",
            "inclusion_proof_size_bytes",
        ],
        "build_then_add_new": ["insert_avg_ns", "insert_total_alloc_mb"],
        "build_then_reinsert_existing": ["insert_avg_ns", "insert_total_alloc_mb"],
    }

    for row in aggregated_rows:
        if row["scenario_family"] not in metric_sets:
            continue
        pair_key = scenario_pair_key(row)
        target = ordered_rows if row["is_ordered_prehashed"] else default_rows
        target[pair_key] = row

    ratio_rows: list[dict[str, Any]] = []
    for key, default_row in sorted(
        default_rows.items(),
        key=lambda item: (
            TREE_ORDER.index(item[0][0]) if item[0][0] in TREE_ORDER else len(TREE_ORDER),
            item[0][1],
            item[0][2],
            item[0][3],
        ),
    ):
        if key not in ordered_rows:
            continue
        ordered_row = ordered_rows[key]
        for metric in metric_sets[default_row["scenario_family"]]:
            ratio_rows.append(
                {
                    "tree_type": default_row["tree_type"],
                    "scenario_family": default_row["scenario_family"],
                    "scale_size": default_row["scale_size"],
                    "scale_label": default_row["scale_label"],
                    "prebuild_scale": default_row["prebuild_scale"],
                    "measured_insert_scale": default_row["measured_insert_scale"],
                    "final_scale": default_row["final_scale"],
                    "metric": metric,
                    "default_value": default_row[metric],
                    "ordered_value": ordered_row[metric],
                    "ordered_over_default_ratio": float(ordered_row[metric]) / float(default_row[metric]),
                }
            )
    return ratio_rows


def make_ordered_plot(aggregated_rows: list[dict[str, Any]], plots_dir: Path) -> list[dict[str, Any]]:
    ratio_rows = build_ordered_ratio_rows(aggregated_rows)
    panels = [
        ("insert_only_build", "insert_avg_ns", "Full-build insert latency"),
        ("proof_only_after_build", "inclusion_proof_gen_avg_ns", "Proof generation"),
        ("proof_only_after_build", "inclusion_proof_verify_avg_ns", "Proof verification"),
    ]
    available_panels = [
        panel
        for panel in panels
        if any(row["scenario_family"] == panel[0] and row["metric"] == panel[1] for row in ratio_rows)
    ]
    if not available_panels:
        return ratio_rows

    fig, axes = plt.subplots(1, len(available_panels), figsize=(6.2 * len(available_panels), 5.4))
    if len(available_panels) == 1:
        axes = [axes]
    for ax, (family, metric, title) in zip(axes, available_panels):
        panel_rows = [row for row in ratio_rows if row["scenario_family"] == family and row["metric"] == metric]
        all_x = [row["scale_size"] for row in panel_rows]
        for tree_type in TREE_ORDER:
            series = sorted([row for row in panel_rows if row["tree_type"] == tree_type], key=lambda row: row["scale_size"])
            ax.plot(
                [row["scale_size"] for row in series],
                [row["ordered_over_default_ratio"] for row in series],
                label=TREE_LABELS[tree_type],
                **tree_line_style(tree_type, series_index=PRIMARY_SERIES_INDEX),
            )
        configure_log_x_axis(ax, all_x)
        ax.axhline(1.0, color="#444444", linestyle=":", linewidth=1.2)
        ax.set_title(title)
        ax.set_ylabel("Ordered / default ratio")
        ax.set_xlabel("Scale")
    axes[0].legend(loc="upper right", ncol=2)
    fig.suptitle("Ordered-Prehashed vs Default Arrival Order", y=1.02, fontsize=15)
    save_figure(fig, plots_dir, "06_ordered_vs_default")
    return ratio_rows


def nearest_available(values: list[int], targets: list[int]) -> list[int]:
    if not values:
        return []
    selected: list[int] = []
    unique_values = sorted(set(values))
    for target in targets:
        choice = min(unique_values, key=lambda value: abs(value - target))
        if choice not in selected:
            selected.append(choice)
    return selected


def make_bucket_plot(
    bucket_rows: list[dict[str, Any]],
    plots_dir: Path,
    *,
    scenario_family: str,
    targets: list[int],
    stem: str,
    title: str,
    ylabel: str,
) -> list[dict[str, Any]]:
    scoped = [
        row
        for row in bucket_rows
        if row["scenario_family"] == scenario_family and not row["is_ordered_prehashed"]
    ]
    selected_scales = nearest_available([int(row["scale_size"]) for row in scoped], targets)
    if not selected_scales:
        return []

    fig, axes = plt.subplots(1, len(selected_scales), figsize=(5.8 * len(selected_scales), 5.1), sharey=True)
    if len(selected_scales) == 1:
        axes = [axes]
    for ax, scale in zip(axes, selected_scales):
        scale_rows = [row for row in scoped if int(row["scale_size"]) == scale]
        for tree_type in TREE_ORDER:
            series = sorted([row for row in scale_rows if row["tree_type"] == tree_type], key=lambda row: row["bucket_mid_pct"])
            xs = [row["bucket_mid_pct"] for row in series]
            ys = [ns_to_us(row["avg_duration_ns"]) for row in series]
            ax.plot(
                xs,
                ys,
                label=TREE_LABELS[tree_type],
                **tree_line_style(tree_type, series_index=PRIMARY_SERIES_INDEX),
            )
            if any(int(row.get("run_count", 1)) > 1 for row in series):
                low = [ns_to_us(row["avg_duration_ns_p10"]) for row in series]
                high = [ns_to_us(row["avg_duration_ns_p90"]) for row in series]
                ax.fill_between(xs, low, high, color=TREE_COLORS[tree_type], alpha=0.12)
        ax.set_title(human_count(scale))
        ax.set_xlabel("Progress bucket midpoint (%)")
        ax.set_xlim(5, 95)
        ax.set_xticks([5, 25, 45, 65, 85])
        ax.set_ylabel(ylabel)
        annotate_lower_is_better(ax)
    axes[0].legend(loc="upper left", ncol=2)
    fig.suptitle(title, y=1.02, fontsize=15)
    save_figure(fig, plots_dir, stem)
    return [row for row in scoped if int(row["scale_size"]) in selected_scales]


def pair_cross_tree_rows(
    aggregated_rows: list[dict[str, Any]],
    *,
    scenario_family: str,
    metric: str,
    is_ordered_prehashed: bool = False,
) -> list[dict[str, Any]]:
    grouped: defaultdict[tuple[Any, ...], dict[str, dict[str, Any]]] = defaultdict(dict)
    for row in aggregated_rows:
        if row["scenario_family"] != scenario_family or row["is_ordered_prehashed"] != is_ordered_prehashed:
            continue
        sample_key = int(row["proof_sample_pct"]) if row["proof_sample_pct"] not in ("", None) else ""
        group_key = (
            int(row["scale_size"]),
            int(row["measured_insert_scale"]),
            int(row["final_scale"]),
            sample_key,
        )
        grouped[group_key][row["tree_type"]] = row

    comparison_rows: list[dict[str, Any]] = []
    for (scale_size, measured_insert_scale, final_scale, sample_key), bundle in sorted(grouped.items(), key=lambda item: item[0]):
        if not {"avlhashtree", "smt"} <= set(bundle):
            continue
        avl_row = bundle["avlhashtree"]
        smt_row = bundle["smt"]
        avl_value = float(avl_row[metric])
        smt_value = float(smt_row[metric])
        comparison_rows.append(
            {
                "scenario_family": scenario_family,
                "metric": metric,
                "scale_size": scale_size,
                "scale_label": human_count(scale_size),
                "measured_insert_scale": measured_insert_scale,
                "final_scale": final_scale,
                "proof_sample_pct": sample_key,
                "avl_value": avl_value,
                "smt_value": smt_value,
                "avl_over_smt_ratio": avl_value / smt_value,
                "leader": "avlhashtree" if avl_value < smt_value else ("smt" if smt_value < avl_value else "tie"),
            }
        )
    return comparison_rows


def build_cross_tree_summary(aggregated_rows: list[dict[str, Any]]) -> tuple[list[dict[str, Any]], list[dict[str, Any]]]:
    summary_rows: list[dict[str, Any]] = []
    crossover_rows: list[dict[str, Any]] = []
    requests = [
        ("insert_only_build", "insert_avg_ns"),
        ("insert_only_build", "insert_total_ns"),
        ("insert_only_build", "insert_mem_alloc_mb"),
        ("insert_only_build", "insert_total_alloc_mb"),
        ("proof_only_after_build", "inclusion_proof_gen_avg_ns"),
        ("proof_only_after_build", "inclusion_proof_verify_avg_ns"),
        ("proof_only_after_build", "inclusion_proof_size_bytes"),
        ("exclusion_proof_only_after_build", "exclusion_proof_gen_avg_ns"),
        ("exclusion_proof_only_after_build", "exclusion_proof_verify_avg_ns"),
        ("exclusion_proof_only_after_build", "exclusion_proof_size_bytes"),
        ("build_then_add_new", "insert_avg_ns"),
        ("build_then_reinsert_existing", "insert_avg_ns"),
    ]
    for family, metric in requests:
        rows = pair_cross_tree_rows(aggregated_rows, scenario_family=family, metric=metric)
        summary_rows.extend(rows)
        leaders = [row["leader"] for row in rows]
        ratios = [row["avl_over_smt_ratio"] for row in rows]
        changes: list[str] = []
        for index in range(1, len(leaders)):
            if leaders[index] != leaders[index - 1]:
                changes.append(rows[index]["scale_label"])
        crossover_rows.append(
            {
                "scenario_family": family,
                "metric": metric,
                "points_compared": len(rows),
                "avl_better_points": sum(1 for leader in leaders if leader == "avlhashtree"),
                "smt_better_points": sum(1 for leader in leaders if leader == "smt"),
                "tie_points": sum(1 for leader in leaders if leader == "tie"),
                "crossover_scales": "|".join(changes) if changes else "none",
                "median_avl_over_smt_ratio": float(np.median(np.asarray(ratios, dtype=float))) if ratios else "",
                "best_avl_ratio": min(ratios) if ratios else "",
                "worst_avl_ratio": max(ratios) if ratios else "",
            }
        )
    return summary_rows, crossover_rows


def build_bucket_drift_summary(bucket_rows: list[dict[str, Any]]) -> list[dict[str, Any]]:
    grouped: defaultdict[tuple[str, str, bool, int], list[dict[str, Any]]] = defaultdict(list)
    for row in bucket_rows:
        grouped[
            (
                row["tree_type"],
                row["scenario_family"],
                row["is_ordered_prehashed"],
                int(row["scale_size"]),
            )
        ].append(row)

    summary: list[dict[str, Any]] = []
    for (tree_type, family, ordered, scale_size), rows in sorted(
        grouped.items(),
        key=lambda item: (
            TREE_ORDER.index(item[0][0]) if item[0][0] in TREE_ORDER else len(TREE_ORDER),
            item[0][1],
            item[0][3],
        ),
    ):
        ordered_rows = sorted(rows, key=lambda row: row["bucket_mid_pct"])
        if not ordered_rows:
            continue
        first = ordered_rows[0]
        last = ordered_rows[-1]
        first_duration = first["avg_duration_ns"]
        last_duration = last["avg_duration_ns"]
        summary.append(
            {
                "tree_type": tree_type,
                "scenario_family": family,
                "is_ordered_prehashed": ordered,
                "scale_size": scale_size,
                "scale_label": human_count(scale_size),
                "first_bucket_ns": first_duration,
                "last_bucket_ns": last_duration,
                "last_over_first_ratio": float(last_duration) / float(first_duration) if first_duration else "",
            }
        )
    return summary


def ratio_range(rows: list[dict[str, Any]], key: str) -> tuple[float, float] | None:
    if not rows:
        return None
    values = np.asarray([row[key] for row in rows], dtype=float)
    return float(np.min(values)), float(np.max(values))


def median_ratio(rows: list[dict[str, Any]], key: str) -> float | None:
    if not rows:
        return None
    return float(np.median(np.asarray([row[key] for row in rows], dtype=float)))


def format_ratio_span(rows: list[dict[str, Any]], key: str, precision: int = 3) -> str:
    ratio_values = ratio_range(rows, key)
    if ratio_values is None:
        return "n/a"
    low, high = ratio_values
    return f"{low:.{precision}f}x to {high:.{precision}f}x"


def build_report(
    raw_rows: list[dict[str, Any]],
    aggregated_rows: list[dict[str, Any]],
    coverage_rows: list[dict[str, Any]],
    cross_tree_rows: list[dict[str, Any]],
    crossover_rows: list[dict[str, Any]],
    reinsert_rows: list[dict[str, Any]],
    ordered_rows: list[dict[str, Any]],
    bucket_drift_rows: list[dict[str, Any]],
    plots_dir: Path,
    input_paths: list[Path],
) -> str:
    insert_default = [
        row for row in cross_tree_rows if row["scenario_family"] == "insert_only_build" and row["metric"] == "insert_avg_ns"
    ]
    proof_gen = [
        row
        for row in cross_tree_rows
        if row["scenario_family"] == "proof_only_after_build" and row["metric"] == "inclusion_proof_gen_avg_ns"
    ]
    proof_verify = [
        row
        for row in cross_tree_rows
        if row["scenario_family"] == "proof_only_after_build" and row["metric"] == "inclusion_proof_verify_avg_ns"
    ]
    proof_size = [
        row
        for row in cross_tree_rows
        if row["scenario_family"] == "proof_only_after_build" and row["metric"] == "inclusion_proof_size_bytes"
    ]
    exclusion_proof_gen = [
        row
        for row in cross_tree_rows
        if row["scenario_family"] == "exclusion_proof_only_after_build" and row["metric"] == "exclusion_proof_gen_avg_ns"
    ]
    exclusion_proof_verify = [
        row
        for row in cross_tree_rows
        if row["scenario_family"] == "exclusion_proof_only_after_build" and row["metric"] == "exclusion_proof_verify_avg_ns"
    ]
    exclusion_proof_size = [
        row
        for row in cross_tree_rows
        if row["scenario_family"] == "exclusion_proof_only_after_build" and row["metric"] == "exclusion_proof_size_bytes"
    ]
    ordered_insert = [
        row for row in ordered_rows if row["scenario_family"] == "insert_only_build" and row["metric"] == "insert_avg_ns"
    ]
    insert_drift = [
        row
        for row in bucket_drift_rows
        if row["scenario_family"] == "insert_only_build" and not row["is_ordered_prehashed"]
    ]

    avl_insert_drift = [row for row in insert_drift if row["tree_type"] == "avlhashtree"]
    smt_insert_drift = [row for row in insert_drift if row["tree_type"] == "smt"]

    schema_notes = [
        row for row in coverage_rows if row["coverage_note"] or row["missing_optional_columns"] or row["schema_variants"] != "current"
    ]
    note_lines = [
        f"- {row['tree_type']} `{row['canonical_scenario']}`: {row['coverage_note'] or row['schema_variants']}"
        for row in schema_notes[:12]
    ]
    if not note_lines:
        note_lines = ["- None."]

    exclusion_nonzero = sum(1 for row in raw_rows if row["exclusion_proof_total_ns"] > 0 or row["delete_total_ns"] > 0)
    build_insert_crossovers = next(
        (
            row
            for row in crossover_rows
            if row["scenario_family"] == "insert_only_build" and row["metric"] == "insert_avg_ns"
        ),
        None,
    )

    lines = [
        "# Benchmark Visualization Summary",
        "",
        "## Dataset",
        f"- Inputs: {', '.join(f'`{relative_display_path(path)}`' for path in input_paths)}.",
        f"- Raw rows loaded: {len(raw_rows)}.",
        f"- Aggregated scenario rows: {len(aggregated_rows)}.",
        f"- Distinct run IDs: {len({row['run_id'] for row in raw_rows})}.",
        f"- Rows with non-zero exclusion/delete totals: {exclusion_nonzero}.",
        "",
        "## Interpretation Notes",
        "- Insert totals, proof totals, and delete totals are preserved in the tables as totals.",
        "- Cross-scenario proof comparisons in the charts use average proof generation time, average proof verification time, and average proof size.",
        "- Heap object charts use created, freed, and net live object counts when those columns are available.",
        "- Exclusion-proof metrics are charted when exclusion-only benchmark CSVs are included.",
        "- Delete metrics are preserved in the tables when present, but they are not charted by the main report.",
        "- Insert buckets are plotted only for measured insert phases because the current CSV schema no longer exposes proof bucket columns.",
        "- Current committed suites actively expose ordered-prehashed full-build rows, while ordered proof-only and ordered post-build rows are treated as historical-only if they appear in older CSVs.",
        "",
        "## Coverage Notes",
        *note_lines,
        "",
        "## Main Findings",
        f"- Default full-build insert latency AVL/SMT ratio range: {format_ratio_span(insert_default, 'avl_over_smt_ratio')}.",
        f"- Proof generation AVL/SMT ratio range: {format_ratio_span(proof_gen, 'avl_over_smt_ratio')}.",
        f"- Proof verification AVL/SMT ratio range: {format_ratio_span(proof_verify, 'avl_over_smt_ratio')}.",
        f"- Proof size AVL/SMT ratio range: {format_ratio_span(proof_size, 'avl_over_smt_ratio', precision=2)}.",
        f"- Exclusion proof generation AVL/SMT ratio range: {format_ratio_span(exclusion_proof_gen, 'avl_over_smt_ratio')}.",
        f"- Exclusion proof verification AVL/SMT ratio range: {format_ratio_span(exclusion_proof_verify, 'avl_over_smt_ratio')}.",
        f"- Exclusion proof size AVL/SMT ratio range: {format_ratio_span(exclusion_proof_size, 'avl_over_smt_ratio', precision=2)}.",
        f"- Ordered/default full-build insert ratio range: {format_ratio_span(ordered_insert, 'ordered_over_default_ratio')}.",
        f"- Update/add new latency ratio range: {format_ratio_span(reinsert_rows, 'reinsert_to_new_ratio')}.",
        f"- Median default insert-bucket drift ratio: AVLHashTree={median_ratio(avl_insert_drift, 'last_over_first_ratio') or float('nan'):.3f}x, SMT={median_ratio(smt_insert_drift, 'last_over_first_ratio') or float('nan'):.3f}x.",
        "",
        "## Generated Outputs",
        f"- Plots: `{relative_display_path(plots_dir)}`",
        "- Tables: `tables/`",
        "- Key plot stems:",
        "  - `01_insert_avg_latency_vs_scale`",
        "  - `02_insert_total_time_vs_scale`",
        "  - `03_inclusion_proof_scaling`",
        "  - `04_memory_vs_scale`",
        "  - `05_update_vs_new`",
        "  - `06_ordered_vs_default`",
        "  - `07_insert_bucket_progression`",
        "  - `09_inclusion_proof_generation_vs_scale`",
        "  - `10_inclusion_proof_verification_vs_scale`",
        "  - `11_inclusion_proof_size_vs_scale`",
        "  - `12_exclusion_proof_scaling`",
        "",
        "## Summary Table Note",
        f"- Insert-latency crossover points: `{build_insert_crossovers['crossover_scales'] if build_insert_crossovers else 'n/a'}`.",
    ]
    return "\n".join(lines)


def main() -> None:
    args = parse_args()
    input_paths = resolve_input_paths(args)
    if not input_paths:
        raise SystemExit(
            "No benchmark CSV inputs were found. Pass one or more CSV files or timestamped directories when data is available."
        )

    plots_dir = (ROOT / args.plots_dir).resolve()
    tables_dir = (ROOT / args.tables_dir).resolve()
    report_path = (ROOT / args.report).resolve()
    prepare_output_dirs(plots_dir, tables_dir)

    raw_rows: list[dict[str, Any]] = []
    for csv_path in input_paths:
        raw_rows.extend(load_rows(csv_path))

    aggregated_rows, coverage_rows = aggregate_rows(raw_rows)
    bucket_rows = expand_bucket_rows(raw_rows)
    aggregated_bucket_rows = aggregate_bucket_rows(bucket_rows)

    make_insert_latency_plot(aggregated_rows, plots_dir)
    make_insert_total_plot(aggregated_rows, plots_dir)
    make_proof_scaling_plot(aggregated_rows, plots_dir)
    make_proof_scaling_plot(
        aggregated_rows,
        plots_dir,
        scenario_family="exclusion_proof_only_after_build",
        proof_kind="exclusion",
        title="Exclusion Proof Scaling (Proof-Only Scenarios)",
        stem="12_exclusion_proof_scaling",
    )
    make_memory_plot(aggregated_rows, plots_dir)
    reinsert_rows = make_reinsert_plot(aggregated_rows, plots_dir)
    ordered_rows = make_ordered_plot(aggregated_rows, plots_dir)
    make_bucket_plot(
        aggregated_bucket_rows,
        plots_dir,
        scenario_family="insert_only_build",
        targets=[100_000, 1_000_000, 25_000_000],
        stem="07_insert_bucket_progression",
        title="Insert Progress Buckets for Representative Full Builds",
        ylabel="Average bucket latency (µs/op)",
    )
    make_proof_metric_plot(
        aggregated_rows,
        plots_dir,
        metric="inclusion_proof_gen_avg_ns",
        ylabel="Average proof generation time (µs/proof)",
        stem="09_inclusion_proof_generation_vs_scale",
        title="Inclusion Proof Generation vs Final Scale",
        transform=ns_to_us,
    )
    make_proof_metric_plot(
        aggregated_rows,
        plots_dir,
        metric="inclusion_proof_verify_avg_ns",
        ylabel="Average proof verification time (µs/proof)",
        stem="10_inclusion_proof_verification_vs_scale",
        title="Inclusion Proof Verification vs Final Scale",
        transform=ns_to_us,
    )
    make_proof_metric_plot(
        aggregated_rows,
        plots_dir,
        metric="inclusion_proof_size_bytes",
        ylabel="Average public proof size (bytes)",
        stem="11_inclusion_proof_size_vs_scale",
        title="Inclusion Proof Size vs Final Scale",
        transform=float,
    )

    cross_tree_rows, crossover_rows = build_cross_tree_summary(aggregated_rows)
    bucket_drift_rows = build_bucket_drift_summary(aggregated_bucket_rows)

    write_csv(tables_dir / "benchmark_rows_cleaned.csv", raw_rows)
    write_csv(tables_dir / "benchmark_aggregated.csv", aggregated_rows)
    write_csv(tables_dir / "coverage_notes.csv", coverage_rows)
    write_csv(tables_dir / "bucket_rows_long.csv", bucket_rows)
    write_csv(tables_dir / "bucket_aggregated.csv", aggregated_bucket_rows)
    write_csv(tables_dir / "cross_tree_summary.csv", cross_tree_rows)
    write_csv(tables_dir / "crossover_points.csv", crossover_rows)
    write_csv(tables_dir / "ordered_vs_default_ratio.csv", ordered_rows)
    write_csv(tables_dir / "reinsert_vs_new_ratio.csv", reinsert_rows)
    write_csv(tables_dir / "bucket_drift_summary.csv", bucket_drift_rows)

    report_text = build_report(
        raw_rows,
        aggregated_rows,
        coverage_rows,
        cross_tree_rows,
        crossover_rows,
        reinsert_rows,
        ordered_rows,
        bucket_drift_rows,
        plots_dir,
        input_paths,
    )
    report_path.write_text(report_text + "\n", encoding="utf-8")

    print(f"Loaded {len(raw_rows)} raw benchmark rows from {len(input_paths)} CSV file(s).")
    print(f"Aggregated rows written to {tables_dir}.")
    print(f"Plots written to {plots_dir}.")
    print(f"Report written to {report_path}.")


if __name__ == "__main__":
    main()
