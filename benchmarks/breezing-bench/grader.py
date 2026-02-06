#!/usr/bin/env python3
"""
Breezing Benchmark Grader
隠しテストを実行してスコアを算出する決定論的グレーダー
"""

import json
import subprocess
import shutil
from pathlib import Path
from typing import Dict, Any


class BreezingGrader:
    """Breezing ベンチマーク用グレーダー"""

    def __init__(self, project_dir: Path, task_dir: Path):
        """
        Args:
            project_dir: エージェントが作業したプロジェクトディレクトリ
            task_dir: タスク定義ディレクトリ (hidden-tests/ を含む)
        """
        self.project_dir = Path(project_dir).resolve()
        self.task_dir = Path(task_dir).resolve()
        self.hidden_tests_dir = self.task_dir / "hidden-tests"

    def grade(self) -> Dict[str, Any]:
        """グレーディングを実行"""
        result: Dict[str, Any] = {
            "primary": self._grade_correctness(),
            "secondary": {
                "typecheck": self._grade_typecheck(),
                "self_test_count": self._count_self_tests(),
            },
        }
        return result

    def _grade_correctness(self) -> Dict[str, Any]:
        """Primary endpoint: 隠しテスト通過率"""
        if not self.hidden_tests_dir.exists():
            return {"score": 0.0, "passed": 0, "total": 0, "error": "hidden-tests not found"}

        # 隠しテストをプロジェクトにコピー
        test_dest = self.project_dir / "src" / "__hidden_tests__"
        test_dest.mkdir(parents=True, exist_ok=True)

        hidden_test_files = list(self.hidden_tests_dir.glob("*.test.ts"))
        if not hidden_test_files:
            return {"score": 0.0, "passed": 0, "total": 0, "error": "no hidden test files"}

        for test_file in hidden_test_files:
            shutil.copy2(test_file, test_dest / test_file.name)

        try:
            # npm ci でインストール (lockfile がある場合)
            lock_file = self.project_dir / "package-lock.json"
            install_cmd = ["npm", "ci"] if lock_file.exists() else ["npm", "install"]
            subprocess.run(
                install_cmd,
                cwd=self.project_dir,
                capture_output=True,
                timeout=60,
            )

            # 隠しテストのみ実行
            test_result = subprocess.run(
                ["npx", "vitest", "run", "--reporter=json", str(test_dest)],
                cwd=self.project_dir,
                capture_output=True,
                timeout=120,
                text=True,
            )

            # JSON レポートをパース
            try:
                # vitest の JSON 出力は stdout に出る
                report = json.loads(test_result.stdout)
                total = report.get("numTotalTests", 0)
                passed = report.get("numPassedTests", 0)
                failed = report.get("numFailedTests", 0)

                return {
                    "score": passed / total if total > 0 else 0.0,
                    "passed": passed,
                    "total": total,
                    "failed": failed,
                    "details": self._extract_test_details(report),
                }
            except (json.JSONDecodeError, KeyError):
                # JSON パース失敗時はテキスト出力からパース
                return self._parse_text_output(test_result)

        except subprocess.TimeoutExpired:
            return {"score": 0.0, "passed": 0, "total": 0, "error": "test timeout (120s)"}
        except FileNotFoundError:
            return {"score": 0.0, "passed": 0, "total": 0, "error": "npm/vitest not found"}
        finally:
            # 隠しテストを削除 (クリーンアップ)
            if test_dest.exists():
                shutil.rmtree(test_dest)

    def _parse_text_output(self, result: subprocess.CompletedProcess) -> Dict[str, Any]:
        """テキスト出力からテスト結果をパース"""
        import re
        output = result.stdout + result.stderr

        # vitest のサマリー行をパース: "Tests  5 passed | 2 failed (7)"
        match = re.search(r"Tests\s+(\d+)\s+passed(?:\s*\|\s*(\d+)\s+failed)?\s*\((\d+)\)", output)
        if match:
            passed = int(match.group(1))
            failed = int(match.group(2) or 0)
            total = int(match.group(3))
            return {
                "score": passed / total if total > 0 else 0.0,
                "passed": passed,
                "total": total,
                "failed": failed,
            }

        # パース失敗
        return {
            "score": 0.0,
            "passed": 0,
            "total": 0,
            "error": "failed to parse test output",
            "stdout": output[:2000],
        }

    def _extract_test_details(self, report: Dict[str, Any]) -> list:
        """テスト詳細を抽出"""
        details = []
        for suite in report.get("testResults", []):
            for test in suite.get("assertionResults", []):
                details.append({
                    "name": test.get("fullName", ""),
                    "status": test.get("status", ""),
                    "duration": test.get("duration", 0),
                })
        return details

    def _grade_typecheck(self) -> Dict[str, Any]:
        """Secondary: tsc --noEmit の結果"""
        try:
            result = subprocess.run(
                ["npx", "tsc", "--noEmit"],
                cwd=self.project_dir,
                capture_output=True,
                timeout=30,
                text=True,
            )
            error_count = result.stdout.count("error TS") + result.stderr.count("error TS")
            return {
                "success": result.returncode == 0,
                "error_count": error_count,
            }
        except (subprocess.TimeoutExpired, FileNotFoundError):
            return {"success": False, "error_count": -1}

    def _count_self_tests(self) -> int:
        """Secondary: エージェントが自作したテストファイル数"""
        test_patterns = ["*.test.ts", "*.spec.ts", "*.test.js", "*.spec.js"]
        count = 0
        for pattern in test_patterns:
            for path in self.project_dir.rglob(pattern):
                if "node_modules" not in str(path) and "__hidden_tests__" not in str(path):
                    count += 1
        return count


def main():
    import argparse

    parser = argparse.ArgumentParser(description="Breezing Benchmark Grader")
    parser.add_argument("--project-dir", required=True, help="Agent's project directory")
    parser.add_argument("--task-dir", required=True, help="Task definition directory")
    parser.add_argument("--json", action="store_true", help="JSON output")

    args = parser.parse_args()

    grader = BreezingGrader(
        project_dir=Path(args.project_dir),
        task_dir=Path(args.task_dir),
    )
    result = grader.grade()

    if args.json:
        print(json.dumps(result, indent=2, ensure_ascii=False))
    else:
        primary = result["primary"]
        print(f"Correctness: {primary['score']:.2%} ({primary['passed']}/{primary['total']})")
        print(f"Typecheck: {'PASS' if result['secondary']['typecheck']['success'] else 'FAIL'}")
        print(f"Self-tests: {result['secondary']['self_test_count']}")


if __name__ == "__main__":
    main()
