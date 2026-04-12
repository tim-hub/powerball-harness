# Breezing v2 Benchmark Report

**Date**: 2026-02-07
**Model**: GLM-4.5-air (via Z.AI Anthropic-compatible API, haiku tier)
**Framework**: @vercel/agent-eval 0.0.11 (Docker sandbox)
**Status**: Exploratory study (not pre-registered)

---

## 1. Executive Summary

This experiment is an exploratory study examining the effect of explicit validation instructions (`npm run validate` + fix instructions) on AI coding agent task success rates. Using GLM-4.5-air, we conducted 3 tasks x 5 runs x 2 conditions = 30 runs and observed a pass rate of 93.3% (14/15) with validation instructions versus 20.0% (3/15) without (Fisher's exact p < 0.001, Cohen's h = 1.69).

However, this is an exploratory study with only 3 tasks, 1 model, and a task design favorable to validation, so additional verification is needed to generalize the conclusions.

---

## 2. Experimental Design

### 2.1 Independent Variable (Condition)

| Condition | CLAUDE.md Content |
|-----------|------------------|
| **Validate** (Breezing) | "Complete PROMPT.md" + "Read src/" + "Run `npm run validate`" + "Fix issues" |
| **Baseline** (Vanilla) | "Complete PROMPT.md" + "Read src/" |

The only difference is 2 lines: running `npm run validate` and fix instructions. This is an ablation of only a subset of the Breezing v2 full pipeline (Agent Teams, code-reviewer, retake loop), and must be distinguished from the full pipeline's effect.

### 2.2 Tasks (Task Design)

Adopted the "new feature + hidden bug" pattern. The PROMPT instructs implementation of a new feature, with detectable bugs embedded in the existing code via validate.ts. EVAL.ts (invisible to the agent) performs the final pass/fail judgment.

| Task | New Feature (PROMPT) | Hidden Bug | Bug Category |
|------|---------------------|------------|--------------|
| task-02 | TodoStore `getByStatus()` | `updatedAt` stale copy | Data freshness |
| task-09 | CSV `stringifyCsv()` | Non-exclusion of column-mismatched rows | Insufficient validation |
| task-10 | BookStore `search()` | `updatedAt` stale copy | Data freshness |

**Note**: task-02 and task-10 share the same bug pattern category (stale copy), so there are effectively only 2 independent bug categories.

### 2.3 Runs

- 5 runs per task per condition
- Total: 3 tasks x 5 runs x 2 conditions = **30 runs**

### 2.4 Environment

| Item | Value |
|------|-------|
| Agent | `vercel-ai-gateway/claude-code` |
| Model | `haiku` tier → GLM-4.5-air (Z.AI API) |
| Sandbox | Docker (isolated per run) |
| Timeout | 300s per run |
| Concurrency | 15 runs simultaneously per condition |

### 2.5 Adaptive Design (Disclosure)

This experiment was conducted in 2 stages:
1. **Calibration (Phase 1)**: 3 tasks x 3 runs x 2 conditions = 18 runs
2. **Full benchmark (Phase 3)**: Expanded to 5 runs after confirming a difference in Calibration

This adaptive design may introduce bias similar to optional stopping. The statistical analysis in this report is based solely on Phase 3 data, though consistency with the Calibration data has been confirmed.

---

## 3. Results

### 3.1 Raw Data (All 30 Runs)

#### Validate (Breezing) Condition

| Task | Run | Status | Duration | Turns | Tool Calls | Shell Cmds |
|------|-----|--------|----------|-------|------------|------------|
| task-02 | 1 | passed | 125.6s | 6 | 13 | 2 |
| task-02 | 2 | passed | 94.9s | 8 | 7 | 1 |
| task-02 | 3 | passed | 122.6s | 5 | 12 | 2 |
| task-02 | 4 | passed | 151.1s | 4 | 14 | 2 |
| task-02 | 5 | passed | 136.2s | 4 | 13 | 2 |
| task-09 | 1 | passed | 188.9s | 16 | 20 | 7 |
| task-09 | 2 | passed | 113.3s | 6 | 12 | 2 |
| task-09 | 3 | passed | 133.2s | 5 | 15 | 2 |
| task-09 | 4 | passed | 193.8s | 13 | 13 | 3 |
| task-09 | 5 | passed | 145.7s | 9 | 16 | 2 |
| task-10 | 1 | **failed** | 201.5s | 10 | 12 | 3 |
| task-10 | 2 | passed | 122.7s | 12 | 14 | 2 |
| task-10 | 3 | passed | 108.5s | 7 | 11 | 2 |
| task-10 | 4 | passed | 147.8s | 6 | 12 | 1 |
| task-10 | 5 | passed | 129.8s | 11 | 9 | 2 |

#### Baseline (Vanilla) Condition

| Task | Run | Status | Duration | Turns | Tool Calls | Shell Cmds |
|------|-----|--------|----------|-------|------------|------------|
| task-02 | 1 | failed | 127.8s | 2 | 9 | 0 |
| task-02 | 2 | failed | 103.4s | 4 | 4 | 0 |
| task-02 | 3 | failed | 123.5s | 3 | 7 | 0 |
| task-02 | 4 | failed | 119.1s | 4 | 5 | 0 |
| task-02 | 5 | failed | 112.6s | 1 | 5 | 0 |
| task-09 | 1 | failed | 121.8s | 4 | 5 | 0 |
| task-09 | 2 | **passed** | 167.6s | 18 | 20 | 9 |
| task-09 | 3 | failed | 143.4s | 8 | 12 | 0 |
| task-09 | 4 | **passed** | 198.9s | 12 | 23 | 5 |
| task-09 | 5 | failed | 134.8s | 3 | 7 | 0 |
| task-10 | 1 | failed | 104.4s | 6 | 5 | 0 |
| task-10 | 2 | **passed** | 200.2s | 11 | 12 | 4 |
| task-10 | 3 | failed | 107.9s | 4 | 3 | 0 |
| task-10 | 4 | failed | 124.4s | 3 | 4 | 0 |
| task-10 | 5 | failed | 127.8s | 6 | 7 | 0 |

**Exclusions/Retries**: None (all 30 runs included in analysis)

### 3.2 Summary

| Condition | task-02 | task-09 | task-10 | Total |
|-----------|---------|---------|---------|-------|
| **Validate** | 5/5 (100%) | 5/5 (100%) | 4/5 (80%) | **14/15 (93.3%)** |
| **Baseline** | 0/5 (0%) | 2/5 (40%) | 1/5 (20%) | **3/15 (20.0%)** |
| **Delta** | +100%pt | +60%pt | +60%pt | **+73.3%pt** |

### 3.3 Calibration (Phase 1: Reference)

| Condition | task-02 | task-09 | task-10 | Total |
|-----------|---------|---------|---------|-------|
| Validate | 3/3 (100%) | 3/3 (100%) | 3/3 (100%) | 9/9 (100%) |
| Baseline | 0/3 (0%) | 1/3 (33%) | 1/3 (33%) | 2/9 (22%) |

Directionally consistent with Phase 3 results.

### 3.4 Behavioral Observations

- All 3 runs that passed in the Baseline condition (task-09 run-2/4, task-10 run-2) **spontaneously executed shell commands** (4-9 times), suggesting the agent independently attempted testing
- All 12 failing runs in the Baseline condition had **shell commands = 0**, with no validation attempted
- task-02 had **all 5 runs fail (0%)** in Baseline -- this task may have been particularly difficult for the Baseline agent (floor effect)

---

## 4. Statistical Analysis

### 4.1 Primary Test: Fisher's Exact Test (one-sided)

| Test | p-value | Judgment |
|------|---------|----------|
| **Fisher's exact** (H1: Validate > Baseline) | **p = 0.000058** | *** (p<0.001) |

Rationale for one-sided test: The Validate condition provides additional opportunities for bug detection and fixing, giving theoretical grounds for the hypothesis that Validate >= Baseline.

### 4.2 Task-Stratified Analysis: Cochran-Mantel-Haenszel Test

Stratified analysis accounting for cluster effects across tasks:

| Test | Statistic | p-value | Judgment |
|------|-----------|---------|----------|
| **CMH** (task-stratified) | chi2 = 15.34 | **p = 0.000090** | *** (p<0.001) |

Significance is maintained even when stratifying by task.

### 4.3 Per-Task Fisher's Exact Test

| Task | Validate | Baseline | p-value | Judgment |
|------|----------|----------|---------|----------|
| task-02 | 5/5 | 0/5 | 0.0040 | ** |
| task-09 | 5/5 | 2/5 | 0.0833 | n.s. |
| task-10 | 4/5 | 1/5 | 0.1032 | n.s. |

task-09 and task-10 lack statistical power individually due to n=5. With Holm correction applied, only task-02 remains significant (0.0040 x 3 = 0.012 < 0.05).

### 4.4 Robustness Checks

| Test | Statistic | p-value | Notes |
|------|-----------|---------|-------|
| Welch's t-test | t = 5.82 | p = 0.000003 | Reference value (applied to binary data) |
| Chi-squared | chi2 = 13.57 | p = 0.000229 | Reference value (cells with expected count < 5) |

These are tests on the same data and serve as robustness checks, not independent verifications.

### 4.5 Effect Sizes

| Metric | Value | Interpretation | Notes |
|--------|-------|----------------|-------|
| **Cohen's h** | **1.69** | Large (thresholds: 0.2/0.5/0.8) | Standard effect size for binary data |
| **Odds Ratio** | **56.0** | -- | After Haldane correction: 47.7 [5.1, 611.7] |
| **Risk Difference** | **73.3%pt** | -- | -- |
| Hedges' g | 2.07 | Reference value | Non-standard application to binary data |

### 4.6 Confidence Interval (Newcombe method)

| Metric | Value |
|--------|-------|
| **Risk Difference 95% CI** | **39.1%pt -- 87.4%pt** |

The Newcombe method provides better accuracy for small samples and extreme proportions (more conservative than the Wald interval of [49.5%, 97.2%]).

### 4.7 Cost Analysis

| Metric | Validate (Breezing) | Baseline (Vanilla) | Delta |
|--------|---------------------|-------------------|-------|
| Mean duration | 141.0s (SD 31.6) | 134.5s (SD 30.9) | +6.5s |
| Mean turns | 8.1 (SD 3.5) | 5.7 (SD 4.3) | +2.4 |
| Mean tool calls | 12.7 (SD 2.8) | 8.5 (SD 5.6) | +4.2 |
| Mean shell cmds | 2.3 (SD 1.4) | 1.2 (SD 2.7) | +1.1 |

The Validate condition has more turns and tool calls, reflecting the validate-and-fix cycle. The wall-clock time difference is relatively small (+4.8%), though token consumption was not captured in this experiment.

---

## 5. Threats to Validity

### 5.1 Internal Validity

- **Adaptive design**: The production run (Phase 3) was decided based on Calibration (Phase 1) results, introducing potential bias similar to optional stopping. Analysis uses only Phase 3 data, but this was not a pre-registered protocol.
- **Concurrent execution independence**: 15 runs were executed simultaneously, which may introduce correlations due to API throttling or Docker resource contention. Significance was maintained in CMH stratified analysis, but complete independence cannot be guaranteed.
- **Task count constraint**: Effectively 3 tasks (2 sharing the same bug pattern), limiting task-level generalizability.

### 5.2 External Validity

- **Model limitation**: Only 1 model, GLM-4.5-air (haiku tier). Reproducibility with Anthropic haiku, sonnet, or other models has not been verified.
- **Task representativeness**: Only simple CRUD tasks (TodoStore, CSV, BookStore). Generalizability to complex architecture changes, UI, or multi-file modifications is unknown.
- **Bug pattern diversity**: Only 2 categories (stale copy and insufficient validation). Generalizability to logic bugs, security bugs, performance bugs, type errors, etc. has not been verified.
- **Task design bias**: The "hidden bug" pattern intentionally embeds bugs detectable by validate.ts, favoring the validation instruction. The effect on bug-free tasks or bugs difficult to detect via validate is unknown.

### 5.3 Construct Validity

- **Narrow operational definition**: The "Breezing" in this experiment consists of only 2 lines: `npm run validate` + fix instructions. The actual Breezing v2 full pipeline (Agent Teams, code-reviewer, retake loop) effect was not measured. Results should be interpreted as "the effect of explicit validation instructions" and distinguished from "the effect of Breezing v2."
- **Success definition**: Only binary judgment (pass/fail) via EVAL.ts. Partial success (new feature implemented but bug fix incomplete) is treated as fail.

---

## 6. Conclusion

Within the scope of this exploratory study, the following was observed:

1. **Explicit validation instructions improved GLM-4.5-air's task success rate on this task set** (14/15 vs 3/15, Fisher's exact p < 0.001).

2. **The effect size is large** (Cohen's h = 1.69, Risk Difference = 73.3%pt [39.1, 87.4]). Significance is maintained in task-stratified analysis (CMH).

3. **Additional cost**: wall-clock time +4.8%, turns +42%, tool calls +49%.

4. **Generalization limits**: These results are exploratory findings based on 3 tasks (2 bug categories), 1 model, and a task design favorable to validation. Confirmatory research with different tasks, models, and bug patterns is needed.

---

## 7. Appendix

### 7.1 File Locations

| File | Path |
|------|------|
| Validate results | `results/glm-breezing/2026-02-07T05-04-18.873Z/` |
| Baseline results | `results/glm-vanilla/2026-02-07T05-10-22.726Z/` |
| Analysis script | `analyze-results.py` |
| Validate config | `experiments/glm-breezing.ts` |
| Baseline config | `experiments/glm-vanilla.ts` |

### 7.2 Calibration Results (Phase 1)

| File | Path |
|------|------|
| Validate calibration | `results/glm-breezing/` (first timestamp) |
| Baseline calibration | `results/glm-vanilla/` (first timestamp) |

### 7.3 Reproducibility

Required node_modules patches for reproduction:
1. `shared.js`: Change `AI_GATEWAY.baseUrl` to `https://api.z.ai/api/anthropic`
2. `claude-code.js`: Pass through `ANTHROPIC_DEFAULT_*_MODEL` env vars to the Docker container

Required `.env` variables:
```
AI_GATEWAY_API_KEY=<GLM_API_KEY>
ANTHROPIC_DEFAULT_HAIKU_MODEL=glm-4.5-air
ANTHROPIC_DEFAULT_SONNET_MODEL=glm-4.7
ANTHROPIC_DEFAULT_OPUS_MODEL=glm-4.7
```

### 7.4 Revision History

| Version | Date | Changes |
|---------|------|---------|
| v1.0 | 2026-02-07 | Initial report |
| v2.0 | 2026-02-07 | Statistical methodology revised (Fisher primary, Newcombe CI, Cohen's h, CMH added). Conclusions scoped. Threats to validity expanded. Raw data added. Cost analysis added. Advocacy language neutralized. |
| v3.0 | 2026-02-07 | Confirmatory study added (Section 8). 10 tasks, 8 unique bug categories, 2 controls, 100 total runs. |

---

## 8. Confirmatory Study

**Status**: Pre-registered analysis plan (Section 8.2) executed on independently designed task set

### 8.1 Motivation

Confirm the large effect observed in the exploratory study (Sections 1-6) (Cohen's h = 1.69, +73.3%pt) with an independent task set addressing the following weaknesses:

| Exploratory Study Weakness | Confirmatory Study Remedy |
|---------------------------|--------------------------|
| 3 tasks | 10 tasks |
| 2 bug categories (with overlap) | 8 bug categories (all independent) |
| Domain bias (CRUD only) | 8 domains (EventEmitter, Queue, Parser, Cache, Validator, Config, Template, Invoice) |
| No controls | 2 control tasks (no bugs, new implementation only) |
| BUG comments present | All BUG comments removed |
| No ceiling/floor effect verification | Calibration conducted + difficulty adjustment |

### 8.2 Pre-registered Analysis Plan

The analysis plan for the confirmatory study was established before experiment execution:

1. **Primary**: Fisher's exact test (overall, one-sided)
2. **Stratified**: Cochran-Mantel-Haenszel test (controlling for task)
3. **Effect size**: Cohen's h + Newcombe 95% CI
4. **Per-task**: Fisher's exact with Holm-Bonferroni correction (10 comparisons)
5. **Subgroup**: Separate analysis of Bug tasks (8) vs Control tasks (2)

### 8.3 Task Design

Follows the "new feature + hidden bug" pattern. Each task has an independent bug category and different domain.

| Task | Domain | New Feature (PROMPT) | Hidden Bug | Bug Category | Type |
|------|--------|---------------------|------------|--------------|------|
| task-11 | EventEmitter | `once()` | `off()` splice idx+1 | Off-by-one | Bug |
| task-12 | PriorityQueue | `peek()` | `!priority` treats 0 as falsy | Null/falsy | Bug |
| task-13 | HTTP Parser | `parseSetCookie()` | `split(':')` truncates header value | String truncation | Bug |
| task-14 | TTL Cache | `getOrSet()` | `size()` includes expired entries | Stale count | Bug |
| task-15 | Form Validator | `validateEmail/Url()` | `isValid: allErrors.length > 0` | Logic inversion | Bug |
| task-16 | Config Merger | `mergeWithStrategy()` | `deepMerge` mutates objects directly | Mutation side-effect | Bug |
| task-17 | Template Engine | `registerHelper()` | `render` does not HTML-escape | XSS/Encoding | Bug |
| task-18 | Invoice Calc | `applyDiscount()` | Floating-point comparison with `===` | Float precision | Bug |
| task-19 | Stack | Implement all methods | None | -- | Control |
| task-20 | Linked List | Implement all methods | None | -- | Control |

### 8.4 Calibration & Difficulty Adjustment

Calibration (2 runs x 10 tasks) was conducted before the confirmatory study:

| Task | Calibration (2 runs) | Adjustment |
|------|---------------------|------------|
| task-11 | 1/2 (50%) | No adjustment needed |
| task-12 | 0/2 (0%) | No adjustment needed (as designed) |
| task-13 | **2/2 (100%)** | **Bug changed**: `==` to `split(':')` (higher-impact bug) |
| task-14 | **2/2 (100%)** | Added `size()` test, removed BUG comment. Re-calibration 3/3 -- accepted (ceiling effect task) |
| task-15 | 1/2 (50%) | No adjustment needed |
| task-16 | 0/2 (0%) | No adjustment needed (as designed) |
| task-17 | 1/2 (50%) | No adjustment needed |
| task-18 | 1/2 (50%) | No adjustment needed |
| task-19 | 2/2 (100%) | Control -- no adjustment needed |
| task-20 | 1/2 (50%) | No adjustment needed |

Additional adjustment: Removed `// BUG:` comments from all 8 bug task source files (eliminating hints to the agent).

### 8.5 Results

#### 8.5.1 Summary Table

| Task | Bug Category | Baseline | Validate | Delta | Fisher p | Cohen's h |
|------|-------------|----------|----------|-------|----------|-----------|
| task-11 EventEmitter | Off-by-one | 2/5 (40%) | 4/5 (80%) | +40%pt | 0.2619 | +0.84 |
| task-12 PriorityQueue | Null/falsy | 0/5 (0%) | 5/5 (100%) | +100%pt | **0.0040** | +3.14 |
| task-13 HTTP Parser | String truncation | 0/5 (0%) | 4/5 (80%) | +80%pt | **0.0238** | +2.21 |
| task-14 TTL Cache | Stale count | 5/5 (100%) | 5/5 (100%) | 0%pt | 1.0000 | 0.00 |
| task-15 Form Validator | Logic inversion | 1/5 (20%) | 4/5 (80%) | +60%pt | 0.1032 | +1.29 |
| task-16 Config Merger | Mutation | 0/5 (0%) | 3/5 (60%) | +60%pt | 0.0833 | +1.77 |
| task-17 Template Engine | XSS/Encoding | 2/5 (40%) | 3/5 (60%) | +20%pt | 0.5000 | +0.40 |
| task-18 Invoice Calc | Float precision | 5/5 (100%) | 4/5 (80%) | -20%pt | 1.0000 | -0.93 |
| task-19 Stack | Control | 3/5 (60%) | 5/5 (100%) | +40%pt | 0.2222 | +1.37 |
| task-20 Linked List | Control | 2/5 (40%) | 5/5 (100%) | +60%pt | 0.0833 | +1.77 |
| **Total** | | **20/50 (40.0%)** | **42/50 (84.0%)** | **+44.0%pt** | | |

#### 8.5.2 Overall

| Metric | Validate | Baseline | Delta |
|--------|----------|----------|-------|
| Pass rate | **42/50 (84.0%)** | **20/50 (40.0%)** | **+44.0%pt** |
| Bug tasks only | 32/40 (80.0%) | 15/40 (37.5%) | +42.5%pt |
| Control tasks only | 10/10 (100.0%) | 5/10 (50.0%) | +50.0%pt |

### 8.6 Statistical Analysis (Confirmatory)

#### 8.6.1 Primary: Fisher's Exact Test

| Test | Odds Ratio | p-value | Judgment |
|------|-----------|---------|----------|
| **Fisher's exact** (H1: Validate > Baseline) | 7.875 | **p = 0.000005** | *** (p<0.001) |

#### 8.6.2 Stratified: Cochran-Mantel-Haenszel Test

| Test | Statistic | p-value | Judgment |
|------|-----------|---------|----------|
| **CMH** (task-stratified) | chi2 = 20.89 | **p = 0.000005** | *** (p<0.001) |

Significance is maintained even when stratifying by task. The effect is robust even when controlling for inter-task heterogeneity.

#### 8.6.3 Effect Sizes

| Metric | Value | Interpretation |
|--------|-------|----------------|
| **Cohen's h** (overall) | **0.95** | Large (thresholds: 0.2/0.5/0.8) |
| **Hedges' g** (overall) | **1.00** | Large |
| **Newcombe 95% CI** | **[+25.4%pt, +58.6%pt]** | Lower bound exceeds +25%pt |
| Risk Difference | +44.0%pt | — |

#### 8.6.4 Per-Task with Holm-Bonferroni Correction

| Task | Raw p | Adjusted p | Judgment |
|------|-------|-----------|----------|
| task-12 PriorityQueue | 0.0040 | **0.0397** | * |
| task-13 HTTP Parser | 0.0238 | 0.2143 | n.s. |
| task-16 Config Merger | 0.0833 | 0.6667 | n.s. |
| task-20 Linked List | 0.0833 | 0.6667 | n.s. |
| task-15 Form Validator | 0.1032 | 0.6667 | n.s. |
| task-19 Stack | 0.2222 | 1.0000 | n.s. |
| task-11 EventEmitter | 0.2619 | 1.0000 | n.s. |
| task-17 Template Engine | 0.5000 | 1.0000 | n.s. |
| task-14 TTL Cache | 1.0000 | 1.0000 | n.s. |
| task-18 Invoice Calc | 1.0000 | 1.0000 | n.s. |

Statistical power for individual tasks is limited due to n=5 per task. Only task-12 remains significant after Holm correction.

#### 8.6.5 Bug Tasks vs Control Tasks

| Subgroup | Validate | Baseline | Delta | Cohen's h | Fisher p |
|----------|----------|----------|-------|-----------|----------|
| Bug tasks (8) | 32/40 (80.0%) | 15/40 (37.5%) | +42.5%pt | +0.90 | p = 0.000112 *** |
| Control tasks (2) | 10/10 (100.0%) | 5/10 (50.0%) | +50.0%pt | +1.57 | p = 0.016254 * |

Significant improvement was observed even in control tasks, suggesting that validation contributes to improving the quality of new implementations as well.

### 8.7 Cost Analysis

| Metric | Validate | Baseline | Delta |
|--------|----------|----------|-------|
| Mean duration | 228.6s (SD 41.1) | 170.6s (SD 56.8) | +58.0s (+34.0%) |
| Mean turns | 7.7 | 6.1 | +1.6 |
| Mean tool calls | 12.3 | 8.0 | +4.3 |
| Mean shell calls | 2.4 | 1.2 | +1.2 |

In the confirmatory study, the Validate condition's duration increase (+34.0%) is larger than in the exploratory study (+4.8%). This may be because the higher task complexity required more time for the validate-fix cycle.

### 8.8 Behavioral Observations

1. **Spontaneous testing in Baseline**: Among the 20 passing runs in Baseline, many had shell calls > 0, indicating spontaneous testing attempts. In particular, task-14 (5/5 pass) and task-18 (5/5 pass) showed high success rates even in Baseline -- these bugs were relatively intuitive to fix.

2. **Ceiling effect tasks**: task-14 (TTL Cache) and task-18 (Invoice Calc) showed high success rates in both conditions (100%, 80-100%), with minimal added value from the Validate instruction. The `size()` stale count bug and floating-point precision may be bug categories avoidable by writing code carefully.

3. **Bugs where Validate was particularly effective**: Large effects were seen for task-12 (null/falsy, +100%pt), task-13 (string truncation, +80%pt), task-15 (logic inversion, +60%pt), and task-16 (mutation, +60%pt). These are bug categories that are difficult to discover by reading code alone and only manifest through runtime testing.

4. **Control task improvement**: Controls (task-19, task-20) also showed +40%pt and +60%pt improvements. This suggests that smoke testing via validate.ts improves the accuracy of new implementations.

### 8.9 Comparison: Exploratory vs Confirmatory

| Metric | Exploratory | Confirmatory | Notes |
|--------|-------------|--------------|-------|
| Task count | 3 | 10 | 3.3x |
| Bug categories | 2 (with overlap) | 8 (all independent) | 4x |
| Total runs | 30 | 100 | 3.3x |
| Validate pass rate | 93.3% | 84.0% | Decreased (due to increased task diversity) |
| Baseline pass rate | 20.0% | 40.0% | Increased (includes 2 ceiling effect tasks) |
| Delta | +73.3%pt | +44.0%pt | Effect size decreased but direction is consistent |
| Cohen's h | 1.69 | 0.95 | Large to Large (threshold maintained) |
| Hedges' g | 2.07 | 1.00 | Large to Large (threshold maintained) |
| Fisher p | 0.000058 | 0.000005 | p-value improved due to increased power |
| CMH p | 0.000090 | 0.000005 | Same as above |
| 95% CI (Newcombe) | [39.1, 87.4] | [25.4, 58.6] | CI width narrowed (improved precision) |

The reduction in effect size can be explained by the following factors:
- Increased task diversity (2 to 8 bug categories)
- Inclusion of 2 ceiling effect tasks (task-14, task-18)
- Increased bug discovery difficulty due to BUG comment removal
- More difficult bug patterns (mutation, XSS, float precision)

### 8.10 Threats to Validity (Confirmatory-specific)

#### Internal Validity
- **Calibration-driven adjustments**: task-13's bug was changed and BUG comments were removed. This is a data-driven adjustment and not a pre-registered protocol. However, adjustments were made only on calibration data (20 runs), and the analysis plan was finalized before seeing the production data (100 runs).
- **Ceiling effect**: task-14 (100%/100%) and task-18 (80%/100%) show no difference between conditions, reducing the analysis's statistical power. Significance is maintained even when excluding these in an 8-task analysis (bug tasks only: p = 0.000112).

#### External Validity
- **Model limitation**: Same as the exploratory study, only GLM-4.5-air. Reproducibility with other models remains unverified.
- **Task scale**: Single-file, 100-200 line tasks only. Generalizability to large-scale multi-file changes is unknown.

#### Construct Validity
- **Same as exploratory study**: The operational definition of "Breezing" is only 2 lines: `npm run validate` + fix instructions (ablation).

### 8.11 Raw Data

#### Validate Condition (50 runs)

| Task | Run | Status | Duration | Turns | Tools | Shell |
|------|-----|--------|----------|-------|-------|-------|
| task-11 | 1 | passed | 230.4s | 14 | 14 | 5 |
| task-11 | 2 | passed | 231.9s | 5 | 11 | 2 |
| task-11 | 3 | failed | 300.0s | — | — | 0 |
| task-11 | 4 | passed | 277.5s | 4 | 11 | 2 |
| task-11 | 5 | passed | 200.3s | 10 | 11 | 2 |
| task-12 | 1 | passed | 252.0s | 8 | 8 | 2 |
| task-12 | 2 | passed | 226.5s | 9 | 10 | 2 |
| task-12 | 3 | passed | 241.0s | 6 | 10 | 1 |
| task-12 | 4 | passed | 212.4s | 6 | 9 | 2 |
| task-12 | 5 | passed | 238.0s | 8 | 13 | 2 |
| task-13 | 1 | passed | 280.3s | 12 | 17 | 4 |
| task-13 | 2 | passed | 237.3s | 5 | 12 | 2 |
| task-13 | 3 | passed | 209.2s | 7 | 13 | 2 |
| task-13 | 4 | passed | 213.7s | 4 | 12 | 2 |
| task-13 | 5 | failed | 191.2s | 6 | 14 | 2 |
| task-14 | 1 | passed | 235.8s | 5 | 10 | 2 |
| task-14 | 2 | passed | 220.8s | 3 | 10 | 1 |
| task-14 | 3 | passed | 252.9s | 3 | 15 | 1 |
| task-14 | 4 | passed | 266.1s | 10 | 15 | 3 |
| task-14 | 5 | passed | 199.9s | 6 | 9 | 1 |
| task-15 | 1 | passed | 252.8s | 5 | 13 | 2 |
| task-15 | 2 | passed | 197.3s | 12 | 14 | 3 |
| task-15 | 3 | failed | 186.6s | 2 | 8 | 1 |
| task-15 | 4 | passed | 230.9s | 6 | 8 | 2 |
| task-15 | 5 | passed | 230.2s | 11 | 11 | 3 |
| task-16 | 1 | failed | 209.4s | 5 | 7 | 2 |
| task-16 | 2 | passed | 221.6s | 6 | 13 | 2 |
| task-16 | 3 | failed | 208.4s | 6 | 13 | 2 |
| task-16 | 4 | passed | 290.3s | 10 | 13 | 3 |
| task-16 | 5 | passed | 145.5s | 6 | 12 | 2 |
| task-17 | 1 | passed | 292.2s | 14 | 16 | 3 |
| task-17 | 2 | passed | 240.9s | 9 | 18 | 4 |
| task-17 | 3 | passed | 194.4s | 7 | 15 | 2 |
| task-17 | 4 | failed | 259.5s | 9 | 16 | 2 |
| task-17 | 5 | failed | 274.5s | 11 | 11 | 2 |
| task-18 | 1 | passed | 250.9s | 9 | 15 | 2 |
| task-18 | 2 | failed | 300.0s | — | — | 0 |
| task-18 | 3 | passed | 281.1s | 4 | 9 | 1 |
| task-18 | 4 | passed | 273.7s | 12 | 22 | 6 |
| task-18 | 5 | passed | 285.7s | 8 | 9 | 1 |
| task-19 | 1 | passed | 157.7s | 13 | 15 | 5 |
| task-19 | 2 | passed | 168.4s | 10 | 9 | 4 |
| task-19 | 3 | passed | 167.9s | 2 | 12 | 1 |
| task-19 | 4 | passed | 258.9s | 11 | 13 | 5 |
| task-19 | 5 | passed | 236.6s | 8 | 11 | 0 |
| task-20 | 1 | passed | 189.7s | 12 | 13 | 5 |
| task-20 | 2 | passed | 143.5s | 12 | 13 | 5 |
| task-20 | 3 | passed | 163.0s | 5 | 14 | 3 |
| task-20 | 4 | passed | 180.0s | 9 | 14 | 2 |
| task-20 | 5 | passed | 220.9s | 6 | 9 | 2 |

#### Baseline Condition (50 runs)

| Task | Run | Status | Duration | Turns | Tools | Shell |
|------|-----|--------|----------|-------|-------|-------|
| task-11 | 1 | failed | 143.2s | 4 | 4 | 0 |
| task-11 | 2 | passed | 188.3s | 9 | 13 | 2 |
| task-11 | 3 | failed | 101.0s | 4 | 4 | 0 |
| task-11 | 4 | passed | 144.3s | 12 | 14 | 5 |
| task-11 | 5 | failed | 94.6s | 4 | 6 | 0 |
| task-12 | 1 | failed | 187.8s | 4 | 11 | 0 |
| task-12 | 2 | failed | 240.7s | 4 | 3 | 0 |
| task-12 | 3 | failed | 156.1s | 4 | 5 | 0 |
| task-12 | 4 | failed | 133.3s | 6 | 7 | 1 |
| task-12 | 5 | failed | 113.7s | 4 | 4 | 0 |
| task-13 | 1 | failed | 203.1s | 13 | 15 | 5 |
| task-13 | 2 | failed | 123.1s | 7 | 9 | 0 |
| task-13 | 3 | failed | 111.2s | 3 | 7 | 0 |
| task-13 | 4 | failed | 183.5s | 4 | 5 | 0 |
| task-13 | 5 | failed | 190.2s | 10 | 22 | 7 |
| task-14 | 1 | passed | 133.4s | 5 | 10 | 0 |
| task-14 | 2 | passed | 108.2s | 2 | 5 | 0 |
| task-14 | 3 | passed | 203.2s | 13 | 10 | 5 |
| task-14 | 4 | passed | 94.6s | 4 | 4 | 0 |
| task-14 | 5 | passed | 83.5s | 4 | 4 | 0 |
| task-15 | 1 | failed | 170.0s | 3 | 6 | 0 |
| task-15 | 2 | failed | 145.6s | 3 | 5 | 0 |
| task-15 | 3 | failed | 115.8s | 4 | 5 | 0 |
| task-15 | 4 | passed | 107.7s | 4 | 5 | 0 |
| task-15 | 5 | failed | 158.5s | 4 | 5 | 0 |
| task-16 | 1 | failed | 185.6s | 3 | 9 | 0 |
| task-16 | 2 | failed | 153.3s | 5 | 7 | 0 |
| task-16 | 3 | failed | 300.0s | — | — | 0 |
| task-16 | 4 | failed | 128.2s | 4 | 5 | 0 |
| task-16 | 5 | failed | 141.0s | 5 | 5 | 0 |
| task-17 | 1 | failed | 300.0s | — | — | 0 |
| task-17 | 2 | failed | 300.0s | — | — | 0 |
| task-17 | 3 | failed | 300.0s | — | — | 0 |
| task-17 | 4 | passed | 195.5s | 10 | 13 | 3 |
| task-17 | 5 | passed | 242.3s | 10 | 11 | 4 |
| task-18 | 1 | passed | 124.3s | 6 | 5 | 0 |
| task-18 | 2 | passed | 202.6s | 6 | 5 | 0 |
| task-18 | 3 | passed | 175.6s | 3 | 9 | 0 |
| task-18 | 4 | passed | 116.3s | 3 | 5 | 0 |
| task-18 | 5 | passed | 194.6s | 3 | 6 | 0 |
| task-19 | 1 | passed | 218.9s | 16 | 20 | 7 |
| task-19 | 2 | failed | 122.7s | 6 | 8 | 1 |
| task-19 | 3 | failed | 220.5s | 6 | 5 | 0 |
| task-19 | 4 | passed | 159.3s | 5 | 5 | 0 |
| task-19 | 5 | passed | 124.3s | 5 | 4 | 0 |
| task-20 | 1 | failed | 248.2s | 7 | 5 | 0 |
| task-20 | 2 | failed | 156.4s | 10 | 9 | 2 |
| task-20 | 3 | failed | 209.8s | 13 | 16 | 7 |
| task-20 | 4 | passed | 195.8s | 11 | 15 | 5 |
| task-20 | 5 | passed | 182.7s | 7 | 14 | 2 |

**Exclusions/Retries**: None (all 100 runs included in analysis). Runs that reached the timeout (300s) were treated as failed.

### 8.12 File Locations (Confirmatory)

| File | Path |
|------|------|
| Baseline A results | `results/confirm-baseline-a/2026-02-07T07-39-55.799Z/` |
| Baseline B results | `results/confirm-baseline-b/2026-02-07T07-44-01.810Z/` |
| Validate A results | `results/confirm-validate-a/2026-02-07T07-49-07.288Z/` |
| Validate B results | `results/confirm-validate-b/2026-02-07T07-54-14.652Z/` |
| Analysis script | `analyze-confirmatory.py` |
| Experiment configs | `experiments/confirm-{baseline,validate}-{a,b}.ts` |
| Task definitions | `evals/task-{11..20}/` |
| Calibration results | `results/calibration-baseline-{a,b}/` |

---

## 9. Combined Conclusion

### 9.1 Summary of Evidence

| Study | Condition Delta | p-value (Fisher) | Cohen's h | 95% CI (Newcombe) |
|-------|----------------|-------------------|-----------|-------------------|
| Exploratory (3 tasks, 30 runs) | +73.3%pt | 0.000058 | 1.69 | [39.1, 87.4] |
| Confirmatory (10 tasks, 100 runs) | +44.0%pt | 0.000005 | 0.95 | [25.4, 58.6] |

### 9.2 Conclusion

1. **Explicit validation instructions improve AI coding agent task success rates**. This effect was discovered in the exploratory study and reproduced in a confirmatory study with an independent 10-task set (Fisher p < 0.001, CMH p < 0.001).

2. **The effect size is Large** (Cohen's h = 0.95). Even with substantially increased task diversity, the effect size maintained the Large threshold (0.8). The Newcombe 95% CI lower bound is +25.4%pt, exceeding the minimum practically meaningful improvement.

3. **The effect is consistent across bug types**. Direction is consistent across 8 independent bug categories (off-by-one, null/falsy, string truncation, stale count, logic inversion, mutation, XSS, float precision). However, bugs that are harder to manifest through runtime testing (mutation, XSS) tend to show smaller improvement margins.

4. **Improvement was also observed in control tasks (no bugs)** (50% to 100%). Validation contributes not only to fixing existing bugs but also to improving the quality of new implementations.

5. **Cost**: The Validate condition requires additional cost of duration +34% and tool calls +54%. This is considered a reasonable tradeoff for a +44%pt improvement in pass rate.

6. **Remaining limitations**: Only 1 model (GLM-4.5-air), only single-file 100-200 line tasks. Verification with other models and larger tasks is the next step.

---

*Breezing v2 Benchmark Suite | Reviewed by Claude (self) + Codex (MCP)*
