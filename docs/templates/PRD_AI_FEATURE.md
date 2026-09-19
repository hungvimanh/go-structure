# PRD — <AI feature name>

> A PRD template **specifically for AI/assistant features**. Purpose: force the dimensions that deterministic software usually misses. Fill everything in; for unknowns write `ASSUMPTION:` or `<TBD>` (don't leave silent blanks). No silent assumptions — clarify with the human instead.
>
> Copy this file to `docs/specs/YYYY-MM-DD-<feature>.md` when starting `hero-planning` for an AI feature.

**Status:** Draft | In Review | Approved
**Date:** YYYY-MM-DD · **Product Owner:** · **Path (router):** Full / Standard

---

## 1. Context & goals
- **User problem:** <what pain does this feature solve?>
- **Measurable goal:** <e.g. reduce X% time, reach Y% correct answers>
- **Persona / audience:**

## 2. Assistant Intent & Scope  ⭐
- **DOES (in-scope):** <what the assistant must be able to do>
- **DOES NOT (explicitly out-of-scope):** <boundaries — crucial against scope creep & unintended behavior>
- **Primary use cases (prioritized):**

## 3. Behavior under AMBIGUITY / missing info  ⭐
> A core decision for an assistant. Reference [INTERACTION_PATTERNS.md](../INTERACTION_PATTERNS.md).
- When user intent is unclear → does the assistant **ask back / guess-then-state / refuse**? <pick & justify>
- **Confidence threshold** to act vs ask back: `<TBD>`
- Max questions before it must act: `<TBD>`

## 4. Tone & Persona
> Reference [INTERACTION_PATTERNS.md](../INTERACTION_PATTERNS.md) P9 (Tone & Consistency).
- Voice (formal/friendly/concise…):
- Things the assistant must **never** say/do stylistically:
- Supported languages:

## 5. Guardrails & Safety  ⭐
> Follow [SECURITY_STANDARDS.md](../SECURITY_STANDARDS.md) — especially §2 (OWASP LLM Top 10). Reference [INTERACTION_PATTERNS.md](../INTERACTION_PATTERNS.md) P5 (Refusal) and P8 (Memory & Privacy).
- **Refusal policy:** which requests must be refused + how (see the Refusal pattern in INTERACTION_PATTERNS).
- **Sensitive data / PII:** what's collected, what's stored, how it's masked, which regulations.
- **Prompt injection / abuse:** defenses (separate system/user content, check output…).
- **Forbidden / restricted content:**

## 6. Definition of "CORRECT" (output is non-deterministic)  ⭐
> You can't use a single `assertEquals` for AI output. You must define *what good looks like*.
- **Evaluation rubric** (criteria + scale): e.g. factual / on-spec / safe / tone.
- **Golden examples** (input → expected behavior/output): ≥ 5–10 representative ones, including hard/edge cases.
  | Input | Expected behavior/output | Why |
  |-------|--------------------------|-----|
  | | | |

## 7. Eval strategy  ⭐
- **Automated eval set** run on the golden set (scored by the §6 rubric) — this is the AI feature's "tests".
- **Prompt regression:** when changing prompt/model → re-run the eval, compare before/after scores.
- Pass threshold: `<TBD>` (e.g. average rubric score ≥ X, no safety case failing).
- Relation to Definition of Done: the eval set is part of `hero-unit-test`'s coverage for AI features.

## 8. Fallback & Human-in-the-loop  ⭐
> Reference [INTERACTION_PATTERNS.md](../INTERACTION_PATTERNS.md) P4 (Error & Graceful Degradation), P6 (Handoff to Human), P7 (Conversation Repair).
- When the model **fails / times out / is low-confidence** → what does the assistant do? (apologize + suggest / return partial / hand off to a human).
- **Human checkpoints (HITL):** which actions need human approval before execution?
- Behavior when a dependent tool/API fails.

## 9. Model, cost & performance
> Follow [PERFORMANCE_STANDARDS.md](../PERFORMANCE_STANDARDS.md) §2 (AI-specific).
- Model used (e.g. Claude Opus/Sonnet/Haiku) + rationale; **prompt caching REQUIRED** for static prompt parts.
- Token budget / target cost per request (ceiling).
- Target latency (p95).

## 10. Observability
- What to log (with PII redacted), how to assess quality *after* going to prod (collect feedback, sample review).
- Metrics to track (refusal rate, confidence, satisfaction…).

## 11. Acceptance Criteria
- [ ] <criterion 1 — measurable>
- [ ] Eval meets the §7 threshold
- [ ] §5 guardrails tested (including abuse cases)

## 12. API / Interface contract (for parallelization)
> Must be locked before spawning FE & BE in parallel — see `hero-planning`.

## 13. Decision Log
| Date | Question/Issue | Decision | Decided by | Rationale |
|------|----------------|----------|------------|-----------|
| | | | | |

## 14. Assumptions Register
| Assumption | Status (pending/confirmed/rejected) |
|------------|-------------------------------------|
| | |
