# Interaction Patterns — Assistant ↔ End-User (Product Layer)

> A library of **patterns for how the PRODUCT talks to end-users** — distinct from how the developer and the AI communicate during development. This is how the *built assistant* talks to *end-users*.
>
> A **reusable asset**: each AI-feature PRD (§3, §5, §8 of the [template](./templates/PRD_AI_FEATURE.md)) references these patterns instead of redesigning from scratch. Specific thresholds/details (`<TBD>`) are locked per PRD.

## Foundational principles
1. **The user stays in control** — the assistant assists; it doesn't unilaterally make hard-to-reverse decisions for the user.
2. **Be transparent about limits** — say clearly when unsure, when it doesn't know, or when it can't do something; never fabricate.
3. **Error cost shapes behavior** — low-risk actions → just do them; high-risk/hard-to-reverse → confirm first.

---

## P1 — Clarification (clarifying intent)
- **When to ask back:** ambiguous intent, missing required parameters, or multiple interpretations leading to very different results.
- **When NOT to ask:** there's a safe/clear default → just do it and **state the assumption** ("I understood it as X, let me know if that's wrong").
- **How to ask:** one focused question, with 2–4 options if possible; don't torture the user with a chain of questions.
- **Threshold:** at most `<TBD>` questions before giving a best-effort attempt.

## P2 — Confirmation (confirm before acting)
- **Confirmation required** before actions that are: hard to reverse, affect data/money/other people, or go outward (send email, publish…).
- **Not needed** for read-only/non-destructive actions.
- State **what's about to happen + the consequence**, then ask for consent.

## P3 — Uncertainty / Low-confidence
- When unsure → **state the confidence level** + offer to verify, instead of asserting confidently.
- Cite sources/grounding when available. Distinguish "I know" vs "I'm inferring".

## P4 — Error & Graceful Degradation
- On error → explain **in the user's language** (don't swallow the error, don't dump a stack trace).
- Prefer **a partial result + naming what's missing** over a blank failure.
- Suggest the next step / how to retry.

## P5 — Refusal (safe refusal)
- Structure: **a concise refusal + reason + (if any) a valid alternative**.
- Keep a respectful, non-judgmental tone; don't reveal internal filtering details.
- Distinguish "not allowed" vs "can't do it" vs "needs more info".

## P6 — Handoff to Human
- When to escalate: out of scope, high risk, the user requests it, or repeated failure.
- Hand off with a **context summary** so the receiving human doesn't have to ask from scratch.

## P7 — Conversation Repair
- The user says "that's not it" → the assistant **briefly acknowledges, restates the new understanding, confirms**, then continues.
- Don't repeat the same wrong answer; don't get defensive.

## P8 — Memory & Privacy
- Be clear about **what the assistant remembers / doesn't**; give the user a way to view/delete it.
- Don't expose sensitive data in responses/logs; follow the PRD §5 PII policy.

## P9 — Tone & Consistency
- One consistent persona across features (locked in PRD §4).
- Concise by default; detailed when the user needs it.

---

## How to use
In an AI-feature PRD: §3 (ambiguity) points to **P1**; §5 (safety) points to **P5/P8**; §8 (fallback) points to **P4/P6/P7**; §4 (tone) points to **P9**. Each pattern gets the specific threshold/details for that feature.
