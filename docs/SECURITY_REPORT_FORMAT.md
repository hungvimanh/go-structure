# Security Report Format

> The standalone report template and severity guide `hero-security` writes findings into, at `docs/security-reports/YYYY-MM-DD-<slug>.md`.

`hero-security` owns this artifact; do not append its findings into coding/review/test reports as a substitute for the security report.

Use this format:

```markdown
# Security Review: <slug>

## Scope reviewed
- Files / modules:
- Endpoints / services / clients:
- Data class and exposure:
- Account types and identity model:
- Artifacts:
- Commands or evidence read:

## Applied domains
- Domain numbers and names:
- Domains not applicable:
- Domains not reviewed:

## Findings
| Severity | Area | Evidence | Exploit path | Risk / impact | Recommendation | Status |
|---|---|---|---|---|---|---|

## Required fixes before release
- ...

## Recommended improvements
- ...

## Verification
- Tests / scans / config checks / log checks performed:
- Missing verification:

## OWASP coverage
- Web Top 10 categories checked:
- AI/LLM categories checked, if applicable:

## Residual risks
- Accepted risk:
- Owner:
- Deadline:

## Not reviewed / uncertainty
- ...

## Verdict
- Pass / Pass with follow-ups / Blocked
- Required fixes before completion:
```

## Severity guide

- **Critical:** likely unauthorized access, data loss/exfiltration, RCE, secret exposure, supply-chain compromise, payment compromise, or tenant isolation break.
- **High:** exploitable sensitive action/data risk, missing core authz/authn control, unsafe privileged/tool action, or realistic injection path.
- **Medium:** realistic abuse path with limited impact or defense-in-depth gap on a sensitive surface.
- **Low:** hardening, clarity, logging, monitoring, or low-impact misconfiguration.
