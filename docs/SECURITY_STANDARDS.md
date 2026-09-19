# Security Standards

> A **measurable** security baseline for this project. It is the bar `hero-security` grades against. Shift-left: most of it is considered during threat modeling in `hero-planning`.
>
> Fill the `<TBD>`s per your stack/compliance when locked in the first PRD/TDD.
>
> `hero-security`'s fixed review methodology lives alongside this file: [SECURITY_CONTROL_BASELINE.md](./SECURITY_CONTROL_BASELINE.md) (14-domain control checklist), [SECURITY_REVIEW_CHECKLIST.md](./SECURITY_REVIEW_CHECKLIST.md) (OWASP/AI-LLM checklist + release gate), [SECURITY_REPORT_FORMAT.md](./SECURITY_REPORT_FORMAT.md) (report template + severity guide).

## 1. General baseline (every project)
- **Secrets**: NEVER hardcode secrets/keys/API tokens in code or logs. Use env vars / a secret manager. `.env` must be in `.gitignore`. Enable secret-scanning (§4).
- **Dependency & supply-chain**: pin versions (lockfile); run `<AUDIT_CMD>` (e.g. `npm audit` / `pip-audit`); enable auto-updates (Dependabot/Renovate); only add dependencies from trusted sources.
- **Input validation & output encoding**: validate all external input (user, API, file); encode output by context (prevent XSS/SQLi/command injection). Use parameterized queries, no string concatenation.
- **AuthN/AuthZ**: explicit authentication; **least privilege** for every role/key/resource; check permissions server-side, never trust the client.
- **Transport & data at rest**: TLS for all traffic; encrypt sensitive data at rest where needed.
- **Logging**: log enough to investigate but **never leak secrets/PII**; don't print stack traces/internals to end-users.
- **Error handling**: user-facing errors don't reveal system details (paths, versions, queries…).

## 2. AI-specific — OWASP LLM Top 10 (required for AI features)
> Cross-reference `templates/PRD_AI_FEATURE.md §5`.
- **Prompt injection**: separate the **system prompt** from **untrusted content** (user input, fetched data, tool output). Do NOT blindly execute/follow instructions embedded in untrusted data.
- **Sensitive information disclosure**: don't let the model leak secrets/PII/other customers' data; filter output; minimize data put into context.
- **Excessive agency**: restrict **tool/agent permissions** (grant only the tools truly needed); hard-to-reverse actions require confirmation/human-in-the-loop (see [INTERACTION_PATTERNS.md](./INTERACTION_PATTERNS.md) P2/P6).
- **Insecure output handling**: treat model output as **untrusted data** — validate/sanitize before rendering HTML, executing code, or passing it downstream.
- **Guardrails & refusal**: a policy for refusing forbidden content (PRD §5); resist abuse/jailbreaks.
- **Model/plugin supply chain**: only use models/MCP/plugins from trusted sources; record versions.

## 3. Process security (enforced by hooks)
- `main` is protected; changes land via MR. The `git-guard` hook blocks force-push, `commit --no-verify`, `reset --hard`.
- Don't bypass hooks/CI "to go faster".

## 4. Automated tools (optional — degrade if absent)
> None of the hooks bundled by hero-mmt-kit (§3: `git-guard`, `active-state-bridge`, `stop-reminder`) run secret scanning, dependency audit, or SAST. These are placeholders for the consuming project to wire up (e.g. a `gitleaks` pre-commit hook, an `<AUDIT_CMD>` step in CI). Do not assume any of §4 is enforced until the project has actually wired it — check for the tool/hook before claiming this baseline is met.
- **Secret scanning**: e.g. `gitleaks` (can run as a pre-commit hook to **block**). `<TBD: on/off>`
- **`<AUDIT_CMD>`** in CI / the `doctor` command.
- **SAST** (stack-dependent): `<TBD>`.

## 5. Checklist by path (feeds the DoD)
- **Fast path**: no new secret committed; `<AUDIT_CMD>` reports no new critical issues.
- **Standard path**: + input validation for touched code; + permission checks if touching authz.
- **Full path**: + the threat model (Phase 2) is addressed; + `hero-security` passes; + (AI feature) all of §2 verified, including abuse cases.
