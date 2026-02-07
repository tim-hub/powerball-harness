#!/usr/bin/env python3
"""
Breezing v2 Benchmark - Statistical Analysis

Compares Breezing vs Vanilla conditions using:
- Pass rates per task and overall
- Welch's t-test for significance
- Hedges' g for effect size
- Fisher's exact test for pass/fail counts
"""

import json
import os
import sys
from pathlib import Path
from scipy import stats
import numpy as np

# --- Configuration ---

RESULTS_DIR = Path(__file__).parent / "results"

# Automatically find the latest result directories
def find_latest_result(condition: str) -> Path | None:
    cond_dir = RESULTS_DIR / condition
    if not cond_dir.exists():
        return None
    timestamps = sorted(cond_dir.iterdir(), key=lambda p: p.name)
    return timestamps[-1] if timestamps else None


def load_results(result_dir: Path) -> dict[str, list[dict]]:
    """Load all run results grouped by task."""
    tasks = {}
    for task_dir in sorted(result_dir.iterdir()):
        if not task_dir.is_dir():
            continue
        task_name = task_dir.name
        runs = []
        for run_dir in sorted(task_dir.iterdir()):
            if not run_dir.is_dir():
                continue
            result_file = run_dir / "result.json"
            if result_file.exists():
                with open(result_file) as f:
                    runs.append(json.load(f))
        if runs:
            tasks[task_name] = runs
    return tasks


def hedges_g(x: np.ndarray, y: np.ndarray) -> float:
    """Calculate Hedges' g (bias-corrected effect size)."""
    nx, ny = len(x), len(y)
    if nx < 2 or ny < 2:
        return float("nan")

    pooled_std = np.sqrt(
        ((nx - 1) * np.var(x, ddof=1) + (ny - 1) * np.var(y, ddof=1))
        / (nx + ny - 2)
    )

    if pooled_std == 0:
        return float("inf") if np.mean(x) != np.mean(y) else 0.0

    d = (np.mean(x) - np.mean(y)) / pooled_std

    # Bias correction factor
    correction = 1 - 3 / (4 * (nx + ny) - 9)
    return d * correction


def analyze(breezing_dir: Path, vanilla_dir: Path):
    """Run full statistical analysis."""
    breezing = load_results(breezing_dir)
    vanilla = load_results(vanilla_dir)

    all_tasks = sorted(set(breezing.keys()) | set(vanilla.keys()))

    print("=" * 70)
    print("Breezing v2 Benchmark - Statistical Analysis Report")
    print("=" * 70)
    print(f"\nBreezing results: {breezing_dir.name}")
    print(f"Vanilla results:  {vanilla_dir.name}")
    print()

    # --- Per-task analysis ---
    print("-" * 70)
    print("Per-Task Results")
    print("-" * 70)

    total_b_pass = 0
    total_b_total = 0
    total_v_pass = 0
    total_v_total = 0

    all_b_scores = []
    all_v_scores = []

    for task in all_tasks:
        b_runs = breezing.get(task, [])
        v_runs = vanilla.get(task, [])

        b_pass = sum(1 for r in b_runs if r["status"] == "passed")
        v_pass = sum(1 for r in v_runs if r["status"] == "passed")
        b_total = len(b_runs)
        v_total = len(v_runs)

        total_b_pass += b_pass
        total_b_total += b_total
        total_v_pass += v_pass
        total_v_total += v_total

        b_scores = [1 if r["status"] == "passed" else 0 for r in b_runs]
        v_scores = [1 if r["status"] == "passed" else 0 for r in v_runs]
        all_b_scores.extend(b_scores)
        all_v_scores.extend(v_scores)

        b_rate = b_pass / b_total * 100 if b_total > 0 else 0
        v_rate = v_pass / v_total * 100 if v_total > 0 else 0
        diff = b_rate - v_rate

        b_dur = [r["duration"] for r in b_runs]
        v_dur = [r["duration"] for r in v_runs]

        print(f"\n{task}:")
        print(f"  Breezing: {b_pass}/{b_total} ({b_rate:.0f}%)  mean={np.mean(b_dur):.1f}s")
        print(f"  Vanilla:  {v_pass}/{v_total} ({v_rate:.0f}%)  mean={np.mean(v_dur):.1f}s")
        print(f"  Diff:     +{diff:.0f}%pt")

        # Fisher's exact test per task
        if b_total > 0 and v_total > 0:
            table = [[b_pass, b_total - b_pass], [v_pass, v_total - v_pass]]
            _, p_fisher = stats.fisher_exact(table, alternative="greater")
            print(f"  Fisher's exact p={p_fisher:.4f}")

    # --- Overall analysis ---
    print("\n" + "=" * 70)
    print("Overall Results")
    print("=" * 70)

    b_rate_overall = total_b_pass / total_b_total * 100
    v_rate_overall = total_v_pass / total_v_total * 100

    print(f"\nBreezing: {total_b_pass}/{total_b_total} ({b_rate_overall:.1f}%)")
    print(f"Vanilla:  {total_v_pass}/{total_v_total} ({v_rate_overall:.1f}%)")
    print(f"Diff:     +{b_rate_overall - v_rate_overall:.1f}%pt")

    # Convert to numpy arrays for statistical tests
    b_arr = np.array(all_b_scores, dtype=float)
    v_arr = np.array(all_v_scores, dtype=float)

    # Welch's t-test (unequal variance)
    t_stat, p_welch = stats.ttest_ind(b_arr, v_arr, equal_var=False, alternative="greater")

    # Hedges' g
    g = hedges_g(b_arr, v_arr)

    # Fisher's exact test (overall)
    table_overall = [
        [total_b_pass, total_b_total - total_b_pass],
        [total_v_pass, total_v_total - total_v_pass],
    ]
    _, p_fisher_overall = stats.fisher_exact(table_overall, alternative="greater")

    # Chi-squared test
    chi2, p_chi2, _, _ = stats.chi2_contingency(table_overall)

    print("\n" + "-" * 70)
    print("Statistical Tests")
    print("-" * 70)

    print(f"\nWelch's t-test:")
    print(f"  t = {t_stat:.4f}")
    print(f"  p = {p_welch:.6f}  {'*** (p<0.001)' if p_welch < 0.001 else '** (p<0.01)' if p_welch < 0.01 else '* (p<0.05)' if p_welch < 0.05 else 'n.s.'}")

    print(f"\nFisher's exact test:")
    print(f"  p = {p_fisher_overall:.6f}  {'*** (p<0.001)' if p_fisher_overall < 0.001 else '** (p<0.01)' if p_fisher_overall < 0.01 else '* (p<0.05)' if p_fisher_overall < 0.05 else 'n.s.'}")

    print(f"\nChi-squared test:")
    print(f"  chi2 = {chi2:.4f}")
    print(f"  p = {p_chi2:.6f}  {'*** (p<0.001)' if p_chi2 < 0.001 else '** (p<0.01)' if p_chi2 < 0.01 else '* (p<0.05)' if p_chi2 < 0.05 else 'n.s.'}")

    print(f"\nEffect size (Hedges' g):")
    print(f"  g = {g:.4f}", end="")
    if abs(g) >= 0.8:
        print("  (large)")
    elif abs(g) >= 0.5:
        print("  (medium)")
    elif abs(g) >= 0.2:
        print("  (small)")
    else:
        print("  (negligible)")

    # --- 95% CI for difference in proportions ---
    p1 = total_b_pass / total_b_total
    p2 = total_v_pass / total_v_total
    diff_prop = p1 - p2
    se_diff = np.sqrt(p1 * (1 - p1) / total_b_total + p2 * (1 - p2) / total_v_total)
    ci_low = diff_prop - 1.96 * se_diff
    ci_high = diff_prop + 1.96 * se_diff

    print(f"\n95% CI for difference in pass rates:")
    print(f"  {diff_prop*100:.1f}%pt [{ci_low*100:.1f}%pt, {ci_high*100:.1f}%pt]")

    # --- Duration analysis ---
    print("\n" + "-" * 70)
    print("Duration Analysis")
    print("-" * 70)

    b_durations = [r["duration"] for task_runs in breezing.values() for r in task_runs]
    v_durations = [r["duration"] for task_runs in vanilla.values() for r in task_runs]

    print(f"\nBreezing: mean={np.mean(b_durations):.1f}s  std={np.std(b_durations, ddof=1):.1f}s  median={np.median(b_durations):.1f}s")
    print(f"Vanilla:  mean={np.mean(v_durations):.1f}s  std={np.std(v_durations, ddof=1):.1f}s  median={np.median(v_durations):.1f}s")

    # --- Conclusion ---
    print("\n" + "=" * 70)
    print("Conclusion")
    print("=" * 70)

    sig = p_welch < 0.05 or p_fisher_overall < 0.05
    large_effect = abs(g) >= 0.8

    if sig and large_effect:
        print(f"\nBreezing condition is SIGNIFICANTLY better than Vanilla")
        print(f"with a LARGE effect size (g={g:.2f}).")
        print(f"Pass rate improvement: +{diff_prop*100:.1f}%pt")
        print(f"The validate-and-fix cycle provides substantial value.")
    elif sig:
        print(f"\nBreezing condition is significantly better than Vanilla")
        print(f"with a {'medium' if abs(g) >= 0.5 else 'small'} effect size (g={g:.2f}).")
    else:
        print(f"\nNo statistically significant difference detected (p={p_welch:.4f}).")

    print()


if __name__ == "__main__":
    # Allow overriding result dirs via CLI args
    if len(sys.argv) == 3:
        breezing_dir = Path(sys.argv[1])
        vanilla_dir = Path(sys.argv[2])
    else:
        breezing_dir = find_latest_result("glm-breezing")
        vanilla_dir = find_latest_result("glm-vanilla")

    if not breezing_dir or not vanilla_dir:
        print("Error: Could not find result directories.")
        print(f"  Breezing: {breezing_dir}")
        print(f"  Vanilla:  {vanilla_dir}")
        sys.exit(1)

    if not breezing_dir.exists() or not vanilla_dir.exists():
        print(f"Error: Result directory not found.")
        print(f"  Breezing: {breezing_dir} (exists={breezing_dir.exists()})")
        print(f"  Vanilla:  {vanilla_dir} (exists={vanilla_dir.exists()})")
        sys.exit(1)

    analyze(breezing_dir, vanilla_dir)
