# Verdict Framework

## Severity Classification

Establish this framework before conducting any review:

| Severity | Definition | Verdict Impact |
|----------|------------|----------------|
| **critical** | Security vulnerabilities, data loss risk, potential production incidents | 1 item → REQUEST_CHANGES |
| **major** | Breaking existing functionality, clear contradiction with specifications, test failures | 1 item → REQUEST_CHANGES |
| **minor** | Naming improvements, insufficient comments, style inconsistencies | No impact on verdict |
| **recommendation** | Best practice suggestions, future improvement ideas | No impact on verdict |

## AI Residuals Severity Classification

| Severity | Representative Examples | Classification Rationale |
|----------|------------------------|--------------------------|
| **major** | `localhost` / `127.0.0.1` / `0.0.0.0` connection targets, `it.skip` / `describe.skip` / `test.skip`, hardcoded secret-like values, dev/staging fixed URLs | Directly linked to production incidents, misconfiguration, or missed validation. 1 item → `REQUEST_CHANGES` |
| **minor** | `mockData`, `dummy`, `fakeData`, `TODO`, `FIXME` | Likely residuals, but not necessarily an immediate incident |
| **recommendation** | `temporary implementation`, `replace later`, `placeholder implementation` comments | Cannot be immediately classified as a bug |

## Verdict Decision Rule

- **If critical or major findings exist**: Verdict = REQUEST_CHANGES (cite which finding(s) triggered it)
- **If only minor and recommendation findings exist**: Verdict = APPROVE
