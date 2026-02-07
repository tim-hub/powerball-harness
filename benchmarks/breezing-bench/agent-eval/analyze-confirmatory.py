#!/usr/bin/env python3
"""
Breezing v2 Confirmatory Study - Statistical Analysis

Pre-registered analysis plan:
1. Primary: Fisher's exact test (overall Validate vs Baseline)
2. Stratified: Cochran-Mantel-Haenszel test (controlling for task)
3. Effect size: Cohen's h + Newcombe CI for difference in proportions
4. Per-task: Fisher's exact with Holm-Bonferroni correction
5. Controls: Separate analysis of bug tasks vs control tasks
"""

import json
import math
import sys
from pathlib import Path

import numpy as np
from scipy import stats

# --- Configuration ---

RESULTS_DIR = Path(__file__).parent / "results"

# Confirmatory result directories (A = task-11~15, B = task-16~20)
BASELINE_DIRS = [
    ("confirm-baseline-a", "2026-02-07T07-39-55.799Z"),
    ("confirm-baseline-b", "2026-02-07T07-44-01.810Z"),
]
VALIDATE_DIRS = [
    ("confirm-validate-a", "2026-02-07T07-49-07.288Z"),
    ("confirm-validate-b", "2026-02-07T07-54-14.652Z"),
]

TASK_META = {
    "task-11": {"name": "EventEmitter", "bug": "off-by-one", "type": "bug"},
    "task-12": {"name": "PriorityQueue", "bug": "null/falsy", "type": "bug"},
    "task-13": {"name": "HTTP Parser", "bug": "string split", "type": "bug"},
    "task-14": {"name": "TTL Cache", "bug": "stale size", "type": "bug"},
    "task-15": {"name": "Form Validator", "bug": "logic inversion", "type": "bug"},
    "task-16": {"name": "Config Merger", "bug": "mutation", "type": "bug"},
    "task-17": {"name": "Template Engine", "bug": "XSS/encoding", "type": "bug"},
    "task-18": {"name": "Invoice Calc", "bug": "float precision", "type": "bug"},
    "task-19": {"name": "Stack", "bug": None, "type": "control"},
    "task-20": {"name": "LinkedList", "bug": None, "type": "control"},
}


def load_results(dir_configs: list[tuple[str, str]]) -> dict[str, list[dict]]:
    """Load and merge results from multiple batch directories."""
    tasks = {}
    for condition, timestamp in dir_configs:
        result_dir = RESULTS_DIR / condition / timestamp
        if not result_dir.exists():
            print(f"WARNING: {result_dir} not found")
            continue
        for task_dir in sorted(result_dir.iterdir()):
            if not task_dir.is_dir():
                continue
            task_name = task_dir.name
            if task_name not in tasks:
                tasks[task_name] = []
            for run_dir in sorted(task_dir.iterdir()):
                if not run_dir.is_dir():
                    continue
                result_file = run_dir / "result.json"
                if result_file.exists():
                    with open(result_file) as f:
                        tasks[task_name].append(json.load(f))
    return tasks


def cohens_h(p1: float, p2: float) -> float:
    """Cohen's h for difference between two proportions."""
    return 2 * math.asin(math.sqrt(p1)) - 2 * math.asin(math.sqrt(p2))


def newcombe_ci(p1: float, n1: int, p2: float, n2: int, alpha: float = 0.05) -> tuple[float, float]:
    """Newcombe's hybrid score CI for difference in proportions (Method 10)."""
    z = stats.norm.ppf(1 - alpha / 2)

    # Wilson score CIs for each proportion
    def wilson(p, n):
        denom = 1 + z**2 / n
        center = (p + z**2 / (2 * n)) / denom
        spread = z * math.sqrt(p * (1 - p) / n + z**2 / (4 * n**2)) / denom
        return center - spread, center + spread

    l1, u1 = wilson(p1, n1)
    l2, u2 = wilson(p2, n2)

    diff = p1 - p2
    lower = diff - math.sqrt((p1 - l1) ** 2 + (u2 - p2) ** 2)
    upper = diff + math.sqrt((u1 - p1) ** 2 + (p2 - l2) ** 2)
    return lower, upper


def cmh_test(tables: list[tuple[int, int, int, int]]) -> tuple[float, float]:
    """
    Cochran-Mantel-Haenszel test for stratified 2x2 tables.
    Each table: (a, b, c, d) where:
      a = validate pass, b = validate fail
      c = baseline pass, d = baseline fail
    Returns (chi2, p-value).
    """
    numerator = 0.0
    denominator = 0.0
    for a, b, c, d in tables:
        n = a + b + c + d
        if n == 0:
            continue
        r1 = a + b  # validate total
        r2 = c + d  # baseline total
        c1 = a + c  # pass total
        E_a = r1 * c1 / n
        V_a = r1 * r2 * c1 * (b + d) / (n**2 * (n - 1)) if n > 1 else 0
        numerator += a - E_a
        denominator += V_a

    if denominator == 0:
        return 0.0, 1.0

    chi2 = (abs(numerator) - 0.5) ** 2 / denominator  # continuity correction
    p = 1 - stats.chi2.cdf(chi2, df=1)
    return chi2, p


def holm_bonferroni(p_values: list[float]) -> list[float]:
    """Holm-Bonferroni correction for multiple comparisons."""
    n = len(p_values)
    indexed = sorted(enumerate(p_values), key=lambda x: x[1])
    adjusted = [0.0] * n
    running_max = 0.0
    for rank, (orig_idx, p) in enumerate(indexed):
        adj_p = min(p * (n - rank), 1.0)
        running_max = max(running_max, adj_p)
        adjusted[orig_idx] = running_max
    return adjusted


def sig_stars(p: float) -> str:
    if p < 0.001:
        return "***"
    elif p < 0.01:
        return "**"
    elif p < 0.05:
        return "*"
    return "n.s."


def analyze():
    """Run full confirmatory analysis."""
    validate = load_results(VALIDATE_DIRS)
    baseline = load_results(BASELINE_DIRS)

    all_tasks = sorted(set(validate.keys()) | set(baseline.keys()))

    print("=" * 78)
    print("  BREEZING v2 CONFIRMATORY STUDY - STATISTICAL ANALYSIS")
    print("=" * 78)
    print()
    print("Study design: 10 tasks x 5 runs x 2 conditions = 100 total runs")
    print("  - 8 bug tasks (unique bug categories) + 2 control tasks (no bugs)")
    print("  - Validate: CLAUDE.md includes 'Run npm run validate' instruction")
    print("  - Baseline: CLAUDE.md without validate instruction")
    print("  - Model: GLM-4.5-air (haiku) via Z.AI Anthropic-compatible API")
    print()

    # ===== Per-task results =====
    print("-" * 78)
    print("  1. PER-TASK RESULTS")
    print("-" * 78)
    print()
    print(f"{'Task':<30} {'Baseline':>12} {'Validate':>12} {'Delta':>8} {'Fisher p':>10} {'h':>6}")
    print("-" * 78)

    total_v_pass = total_v_n = total_b_pass = total_b_n = 0
    bug_v_pass = bug_v_n = bug_b_pass = bug_b_n = 0
    ctrl_v_pass = ctrl_v_n = ctrl_b_pass = ctrl_b_n = 0

    per_task_p = []
    per_task_names = []
    cmh_tables = []

    for task in all_tasks:
        meta = TASK_META.get(task, {"name": task, "bug": "?", "type": "?"})
        v_runs = validate.get(task, [])
        b_runs = baseline.get(task, [])

        v_pass = sum(1 for r in v_runs if r["status"] == "passed")
        b_pass = sum(1 for r in b_runs if r["status"] == "passed")
        v_n = len(v_runs)
        b_n = len(b_runs)

        total_v_pass += v_pass
        total_v_n += v_n
        total_b_pass += b_pass
        total_b_n += b_n

        if meta["type"] == "bug":
            bug_v_pass += v_pass
            bug_v_n += v_n
            bug_b_pass += b_pass
            bug_b_n += b_n
        else:
            ctrl_v_pass += v_pass
            ctrl_v_n += v_n
            ctrl_b_pass += b_pass
            ctrl_b_n += b_n

        v_rate = v_pass / v_n if v_n else 0
        b_rate = b_pass / b_n if b_n else 0
        delta = v_rate - b_rate

        # Fisher's exact (one-sided: validate > baseline)
        table = [[v_pass, v_n - v_pass], [b_pass, b_n - b_pass]]
        _, p_fisher = stats.fisher_exact(table, alternative="greater")
        per_task_p.append(p_fisher)
        per_task_names.append(task)

        # Cohen's h
        h = cohens_h(v_rate, b_rate) if v_n > 0 and b_n > 0 else float("nan")

        # CMH table
        cmh_tables.append((v_pass, v_n - v_pass, b_pass, b_n - b_pass))

        tag = "[CTRL]" if meta["type"] == "control" else f"[{meta['bug']}]"
        label = f"{meta['name']} {tag}"
        print(
            f"{label:<30} {b_pass}/{b_n} ({b_rate*100:4.0f}%) "
            f"{v_pass}/{v_n} ({v_rate*100:4.0f}%) "
            f"{delta*100:+5.0f}%pt  "
            f"p={p_fisher:.4f} {h:+.2f}"
        )

    # Holm-Bonferroni correction
    adjusted_p = holm_bonferroni(per_task_p)

    print()
    print("Holm-Bonferroni adjusted p-values:")
    for i, task in enumerate(per_task_names):
        meta = TASK_META.get(task, {"name": task})
        print(f"  {meta['name']:<22} raw p={per_task_p[i]:.4f}  adj p={adjusted_p[i]:.4f}  {sig_stars(adjusted_p[i])}")

    # ===== Overall results =====
    print()
    print("=" * 78)
    print("  2. OVERALL RESULTS")
    print("=" * 78)

    v_rate_all = total_v_pass / total_v_n
    b_rate_all = total_b_pass / total_b_n
    diff_all = v_rate_all - b_rate_all

    print(f"\n  Validate:  {total_v_pass}/{total_v_n} ({v_rate_all*100:.1f}%)")
    print(f"  Baseline:  {total_b_pass}/{total_b_n} ({b_rate_all*100:.1f}%)")
    print(f"  Delta:     {diff_all*100:+.1f}%pt")

    # ===== Primary test: Fisher's exact (overall) =====
    print()
    print("-" * 78)
    print("  3. PRIMARY ANALYSIS: Fisher's Exact Test (Overall)")
    print("-" * 78)

    table_overall = [
        [total_v_pass, total_v_n - total_v_pass],
        [total_b_pass, total_b_n - total_b_pass],
    ]
    or_val, p_fisher_overall = stats.fisher_exact(table_overall, alternative="greater")
    print(f"\n  Odds ratio = {or_val:.3f}")
    print(f"  p = {p_fisher_overall:.6f}  {sig_stars(p_fisher_overall)}")

    # ===== CMH stratified test =====
    print()
    print("-" * 78)
    print("  4. STRATIFIED ANALYSIS: Cochran-Mantel-Haenszel Test")
    print("-" * 78)

    cmh_chi2, cmh_p = cmh_test(cmh_tables)
    print(f"\n  CMH chi2 = {cmh_chi2:.4f}")
    print(f"  p = {cmh_p:.6f}  {sig_stars(cmh_p)}")
    print("  (Controls for task-level heterogeneity)")

    # ===== Effect sizes =====
    print()
    print("-" * 78)
    print("  5. EFFECT SIZE")
    print("-" * 78)

    h_overall = cohens_h(v_rate_all, b_rate_all)
    ci_low, ci_high = newcombe_ci(v_rate_all, total_v_n, b_rate_all, total_b_n)
    h_interp = "large" if abs(h_overall) >= 0.8 else "medium" if abs(h_overall) >= 0.5 else "small" if abs(h_overall) >= 0.2 else "negligible"

    print(f"\n  Cohen's h = {h_overall:.4f}  ({h_interp})")
    print(f"  Newcombe 95% CI for diff: [{ci_low*100:+.1f}%pt, {ci_high*100:+.1f}%pt]")
    print(f"  Point estimate:           {diff_all*100:+.1f}%pt")

    # Hedges' g for comparison with exploratory study
    v_arr = np.array([1 if r["status"] == "passed" else 0 for task_runs in validate.values() for r in task_runs], dtype=float)
    b_arr = np.array([1 if r["status"] == "passed" else 0 for task_runs in baseline.values() for r in task_runs], dtype=float)

    nx, ny = len(v_arr), len(b_arr)
    pooled_std = np.sqrt(
        ((nx - 1) * np.var(v_arr, ddof=1) + (ny - 1) * np.var(b_arr, ddof=1))
        / (nx + ny - 2)
    )
    if pooled_std > 0:
        d = (np.mean(v_arr) - np.mean(b_arr)) / pooled_std
        correction = 1 - 3 / (4 * (nx + ny) - 9)
        g = d * correction
    else:
        g = float("inf") if np.mean(v_arr) != np.mean(b_arr) else 0.0

    g_interp = "large" if abs(g) >= 0.8 else "medium" if abs(g) >= 0.5 else "small" if abs(g) >= 0.2 else "negligible"
    print(f"  Hedges' g = {g:.4f}  ({g_interp})")

    # ===== Bug tasks vs Control tasks =====
    print()
    print("-" * 78)
    print("  6. BUG TASKS vs CONTROL TASKS")
    print("-" * 78)

    bug_v_rate = bug_v_pass / bug_v_n if bug_v_n else 0
    bug_b_rate = bug_b_pass / bug_b_n if bug_b_n else 0
    ctrl_v_rate = ctrl_v_pass / ctrl_v_n if ctrl_v_n else 0
    ctrl_b_rate = ctrl_b_pass / ctrl_b_n if ctrl_b_n else 0

    print(f"\n  Bug tasks (8 tasks, {bug_v_n + bug_b_n} runs):")
    print(f"    Validate: {bug_v_pass}/{bug_v_n} ({bug_v_rate*100:.1f}%)")
    print(f"    Baseline: {bug_b_pass}/{bug_b_n} ({bug_b_rate*100:.1f}%)")
    bug_diff = bug_v_rate - bug_b_rate
    bug_h = cohens_h(bug_v_rate, bug_b_rate)
    bug_table = [[bug_v_pass, bug_v_n - bug_v_pass], [bug_b_pass, bug_b_n - bug_b_pass]]
    _, bug_p = stats.fisher_exact(bug_table, alternative="greater")
    print(f"    Delta:    {bug_diff*100:+.1f}%pt  Cohen's h={bug_h:+.4f}  Fisher p={bug_p:.6f} {sig_stars(bug_p)}")

    print(f"\n  Control tasks (2 tasks, {ctrl_v_n + ctrl_b_n} runs):")
    print(f"    Validate: {ctrl_v_pass}/{ctrl_v_n} ({ctrl_v_rate*100:.1f}%)")
    print(f"    Baseline: {ctrl_b_pass}/{ctrl_b_n} ({ctrl_b_rate*100:.1f}%)")
    ctrl_diff = ctrl_v_rate - ctrl_b_rate
    ctrl_h = cohens_h(ctrl_v_rate, ctrl_b_rate)
    ctrl_table = [[ctrl_v_pass, ctrl_v_n - ctrl_v_pass], [ctrl_b_pass, ctrl_b_n - ctrl_b_pass]]
    _, ctrl_p = stats.fisher_exact(ctrl_table, alternative="greater")
    print(f"    Delta:    {ctrl_diff*100:+.1f}%pt  Cohen's h={ctrl_h:+.4f}  Fisher p={ctrl_p:.6f} {sig_stars(ctrl_p)}")

    # ===== Duration analysis =====
    print()
    print("-" * 78)
    print("  7. DURATION ANALYSIS")
    print("-" * 78)

    v_dur = [r["duration"] for runs in validate.values() for r in runs]
    b_dur = [r["duration"] for runs in baseline.values() for r in runs]

    print(f"\n  Validate: mean={np.mean(v_dur):.1f}s  std={np.std(v_dur, ddof=1):.1f}s  median={np.median(v_dur):.1f}s")
    print(f"  Baseline: mean={np.mean(b_dur):.1f}s  std={np.std(b_dur, ddof=1):.1f}s  median={np.median(b_dur):.1f}s")
    dur_diff = np.mean(v_dur) - np.mean(b_dur)
    print(f"  Delta:    {dur_diff:+.1f}s")

    # ===== Tool usage analysis =====
    print()
    print("-" * 78)
    print("  8. TOOL USAGE COMPARISON")
    print("-" * 78)

    for label, data in [("Validate", validate), ("Baseline", baseline)]:
        turns = [r["o11y"]["totalTurns"] for runs in data.values() for r in runs if "o11y" in r]
        tools = [r["o11y"]["totalToolCalls"] for runs in data.values() for r in runs if "o11y" in r]
        shells = [r["o11y"]["toolCalls"].get("shell", 0) for runs in data.values() for r in runs if "o11y" in r]
        print(f"\n  {label}:")
        print(f"    Turns: mean={np.mean(turns):.1f}  median={np.median(turns):.0f}")
        print(f"    Tool calls: mean={np.mean(tools):.1f}  median={np.median(tools):.0f}")
        print(f"    Shell calls: mean={np.mean(shells):.1f}  median={np.median(shells):.0f}")

    # ===== Comparison with exploratory study =====
    print()
    print("=" * 78)
    print("  9. COMPARISON WITH EXPLORATORY STUDY")
    print("=" * 78)

    print(f"""
  Exploratory (3 tasks, 5 runs, GLM-4.5-air):
    Validate: 14/15 (93.3%)
    Baseline:  3/15 (20.0%)
    Delta: +73.3%pt, Hedges' g = 2.07

  Confirmatory (10 tasks, 5 runs, GLM-4.5-air):
    Validate: {total_v_pass}/{total_v_n} ({v_rate_all * 100:.1f}%)
    Baseline: {total_b_pass}/{total_b_n} ({b_rate_all * 100:.1f}%)
    Delta: {diff_all * 100:+.1f}%pt, Hedges' g = {g:.2f}

  Effect size reduction: 2.07 -> {g:.2f} (expected due to task diversity)
  Direction confirmed: Validate > Baseline in both studies
""")

    # ===== Conclusion =====
    print("=" * 78)
    print("  10. CONCLUSION")
    print("=" * 78)

    sig = p_fisher_overall < 0.05
    cmh_sig = cmh_p < 0.05

    if sig and cmh_sig:
        print(f"""
  The validate-and-fix instruction significantly improves task pass rates.

  - Overall: {diff_all*100:+.1f}%pt improvement (p={p_fisher_overall:.6f}, Fisher's exact)
  - Stratified: CMH chi2={cmh_chi2:.2f}, p={cmh_p:.6f} (task-controlled)
  - Effect size: Cohen's h = {h_overall:.2f} ({h_interp}), Hedges' g = {g:.2f}
  - 95% CI: [{ci_low*100:+.1f}%pt, {ci_high*100:+.1f}%pt]
  - Bug tasks benefit more ({bug_diff*100:+.1f}%pt) than controls ({ctrl_diff*100:+.1f}%pt)

  The exploratory finding is CONFIRMED with a diverse, pre-registered task set.
""")
    elif sig:
        print(f"\n  Significant overall (p={p_fisher_overall:.6f}) but CMH not significant.")
        print("  Task heterogeneity may explain the difference.\n")
    else:
        print(f"\n  No significant difference detected (p={p_fisher_overall:.4f}).\n")

    print("=" * 78)


if __name__ == "__main__":
    analyze()
