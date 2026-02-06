#!/usr/bin/env python3
"""
agent-eval 結果 → Breezing analyzer 形式への変換スクリプト

agent-eval の結果ディレクトリ構造:
  results/{experiment}/{timestamp}/
    ├── experiment.json
    └── {eval-name}/
        ├── summary.json
        └── run-{N}/
            ├── result.json
            └── transcript.jsonl

変換先 (analyzer.py が期待する形式):
  converted/{timestamp}/
    └── result.json (各実行ごとに1ファイル)
"""

import json
from pathlib import Path
from typing import Any, Dict, List


def convert_experiment(
    experiment_dir: Path, condition: str
) -> List[Dict[str, Any]]:
    """1つの experiment ディレクトリを変換"""
    results = []

    for eval_dir in sorted(experiment_dir.iterdir()):
        if not eval_dir.is_dir():
            continue
        # experiment.json はスキップ
        if eval_dir.name.startswith(".") or eval_dir.name == "experiment.json":
            continue

        task_id = eval_dir.name  # e.g., "task-01"

        for run_dir in sorted(eval_dir.iterdir()):
            if not run_dir.is_dir() or not run_dir.name.startswith("run-"):
                continue

            result_file = run_dir / "result.json"
            if not result_file.exists():
                continue

            with open(result_file, "r", encoding="utf-8") as f:
                agent_result = json.load(f)

            # agent-eval result → analyzer format
            run_number = int(run_dir.name.replace("run-", ""))
            status = agent_result.get("status", "fail")
            duration = agent_result.get("duration", 0)

            # score: pass=1.0, fail=0.0 (agent-eval は pass/fail のみ)
            # より精密な scoring は EVAL.ts 内のアサーションに依存
            score = 1.0 if status == "pass" else 0.0

            converted = {
                "task_id": task_id,
                "condition": condition,
                "iteration": run_number,
                "grading": {
                    "primary": {
                        "score": score,
                        "passed": 1 if status == "pass" else 0,
                        "total": 1,
                        "failed": 0 if status == "pass" else 1,
                        "details": [],
                    },
                    "secondary": {
                        "typecheck": {"success": status == "pass", "error_count": 0},
                    },
                },
                "execution": {
                    "duration_seconds": duration / 1000 if duration > 1000 else duration,
                    "status": "completed" if status == "pass" else "failed",
                },
                "metadata": {
                    "source": "agent-eval",
                    "experiment": condition,
                },
            }
            results.append(converted)

    return results


def main():
    import argparse

    parser = argparse.ArgumentParser(
        description="Convert agent-eval results to analyzer format"
    )
    parser.add_argument(
        "--vanilla-dir",
        type=Path,
        help="Path to vanilla experiment results (e.g., results/vanilla/2026-01-01T...)",
    )
    parser.add_argument(
        "--breezing-dir",
        type=Path,
        help="Path to breezing experiment results (e.g., results/breezing/2026-01-01T...)",
    )
    parser.add_argument(
        "--output-dir",
        type=Path,
        required=True,
        help="Output directory for converted results",
    )

    args = parser.parse_args()

    output_dir = args.output_dir
    output_dir.mkdir(parents=True, exist_ok=True)

    all_results = []

    if args.vanilla_dir and args.vanilla_dir.exists():
        vanilla_results = convert_experiment(args.vanilla_dir, "vanilla")
        all_results.extend(vanilla_results)
        print(f"Converted {len(vanilla_results)} vanilla results")

    if args.breezing_dir and args.breezing_dir.exists():
        breezing_results = convert_experiment(args.breezing_dir, "breezing")
        all_results.extend(breezing_results)
        print(f"Converted {len(breezing_results)} breezing results")

    # 各結果を個別の result.json として保存
    for result in all_results:
        task_id = result["task_id"]
        condition = result["condition"]
        iteration = result["iteration"]
        result_dir = output_dir / f"{condition}-{task_id}-iter{iteration}"
        result_dir.mkdir(parents=True, exist_ok=True)
        with open(result_dir / "result.json", "w", encoding="utf-8") as f:
            json.dump(result, f, indent=2, ensure_ascii=False)

    print(f"\nTotal: {len(all_results)} results written to {output_dir}")


if __name__ == "__main__":
    main()
