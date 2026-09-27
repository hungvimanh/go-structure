<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **go-structure** (865 symbols, 1962 relationships, 38 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> If any GitNexus tool warns the index is stale, run `npx gitnexus analyze` in terminal first.

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `gitnexus_impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `gitnexus_detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `gitnexus_query({query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `gitnexus_context({name: "symbolName"})`.

## Never Do

- NEVER edit a function, class, or method without first running `gitnexus_impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `gitnexus_rename` which understands the call graph.
- NEVER commit changes without running `gitnexus_detect_changes()` to check affected scope.

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/go-structure/context` | Codebase overview, check index freshness |
| `gitnexus://repo/go-structure/clusters` | All functional areas |
| `gitnexus://repo/go-structure/processes` | All execution flows |
| `gitnexus://repo/go-structure/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
|------|---------------------|
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->

<!-- hero-mmt-kit:start -->
<!-- Managed by hero-mmt-kit. Update via the framework; edits inside this block are overwritten on `update`. -->
## 🧭 hero-mmt-kit workflow

> This is a human-led Claude Code workflow: you (the developer) decide what to work on and invoke the relevant skill directly. There is no router doc and no hard gate — use `using-hero` for the overview when it's unclear which skill applies next.

Core Hero skills:
- `hero-planning` — turn a request into an agreed, implementation-only plan before code changes.
- `hero-coding` — implement and track every approved requirement/task; no testing or self-review.
- `hero-reviewing` — single read-only entry point for a fresh plan-backed review and/or technical assessment of supplied review feedback.
- `hero-unit-test` — write and run post-implementation unit tests without changing production code.
- `hero-security` — standalone OWASP + AI/LLM security review when you want a dedicated security pass.
- `hero-strict` — opt-in, separate full verification pass before a "done" claim.

### Context loading
- **Resume (fast path):** read `.hero-mmt-kit/session.json` first → then the `resumePath` it names → then the latest artifact only if needed. Read [ACTIVE_STATE](docs/ACTIVE_STATE.md) only to switch work items or when session is blank/stale.
- **Read on demand:** [SECURITY_STANDARDS](docs/SECURITY_STANDARDS.md), [PERFORMANCE_STANDARDS](docs/PERFORMANCE_STANDARDS.md), [DESIGN_STANDARDS](docs/DESIGN_STANDARDS.md), [BROWNFIELD_DISCOVERY](docs/BROWNFIELD_DISCOVERY.md), and templates under `docs/templates/` only when the current skill requires them.

### Non-negotiables
- **Gates:** for risky or hard-to-reverse changes, use Plan Mode (`EnterPlanMode` → `ExitPlanMode`) for real sign-off — not a prose promise.
- **Plan Mode:** do not enter Plan Mode unless the user explicitly requests planning. When they do, invoke `hero-planning`.
- **Git:** do not check out a branch, commit, or merge unless the user explicitly requests it.
- **Enforcement:** hooks installed under `.claude/` — `git-guard` (PreToolUse on Bash), `session-bridge` (PostToolUse), `stop-reminder` (Stop). Don't bypass them.
- **GitNexus (if installed):** skip on empty repos; use impact/context analysis for non-trivial code changes and stale-index checks.
- **Skills (if installed):** invoke relevant process skills via the `Skill` tool; do not load unrelated skill/docs context.
- **Language:** framework instructions, skills, rules, and templates remain English-only. User-facing replies and artifacts must use the user's language unless they request another language.
- **Sub-agent delegation is optional, not automatic.** Sub-agents are best for research and review; use one when it helps, and skip it when direct work is clearer. Don't hardcode model IDs — let the harness resolve the active model.

## Response Style

### Language and terminology

- Use the user's language consistently for explanations and prose.
- Keep code identifiers, commands, file paths, API names, configuration keys, and exact error messages verbatim.
- Keep an established technical term in its standard form when translating it would be inaccurate or awkward; briefly explain it in the user's language on first use.
- Do not translate technical tokens or invent awkward equivalents. Avoid unnecessary English code-switching in non-English prose.

### Clarity and audience

- Prefer plain, familiar wording and complete sentences.
- Do not assume technical expertise beyond what the conversation establishes.
- Include enough context for the user to understand the result, make a decision, and act safely.
- Expand an unfamiliar acronym on first use unless it is already established in the conversation.

### Structure

- For substantive output, lead with the conclusion or current status, follow with key facts or evidence, and end with the next action, decision, or a clear statement that none is needed.
- Use headings, lists, tables, and compact notation only when they improve scanning.
- Make plans and reports actionable: use concrete actions, targets, conditions, owners, and evidence where relevant; avoid vague labels such as "handle appropriately."

### Brevity and completeness

- Match the level of detail to the task's complexity.
- Keep simple answers concise, but do not impose a fixed line count.
- Remove filler and repetition, not context required to understand, decide, or act.
- Clarity, completeness, safety, and honesty outrank terseness.
- Fully explain risks, blockers, uncertainty, destructive or irreversible actions, production changes, and consequential decisions.
<!-- hero-mmt-kit:end -->
