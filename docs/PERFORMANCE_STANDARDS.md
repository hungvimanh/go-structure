# Performance Standards

> A **measurable** performance budget for this project. Set the budget during `hero-planning` and verify it during `hero-reviewing`/`hero-unit-test`.
>
> Fill the `<TBD>`s per product requirements when locked in the first PRD/TDD.

## 1. General budget (parameterized)
- **Latency**: API p95 ≤ `<TBD>` ms, p99 ≤ `<TBD>` ms.
- **Throughput**: ≥ `<TBD>` req/s at target load.
- **Size**: payload/response ≤ `<TBD>`; frontend bundle (if any) ≤ `<TBD>` KB.
- **Database**: no N+1; queries per request ≤ `<TBD>`; indexes for hot queries.
- **Memory/CPU**: no leaks; baseline ≤ `<TBD>`.

## 2. AI-specific (required for AI features)
> Cross-reference `templates/PRD_AI_FEATURE.md §9`.
- **Prompt caching — REQUIRED**: leverage Claude's caching for static prompt parts (system, grounding docs) to cut cost & latency. This is the biggest lever.
- **Model selection by task**: use the lightest sufficient model (Haiku → Sonnet → Opus); don't default to the strongest model for everything.
- **Token/cost ceiling**: a per-request token & cost ceiling `<TBD>`; alert on breach.
- **Streaming**: stream long responses to reduce perceived latency.
- **Context discipline**: load only the needed context; avoid stuffing whole documents into every turn.
- **Batching**: batch large non-realtime workloads.

## 3. Performance regression gate
- **`change/` + `refactor/` touching a hot path**: measure a **before & after bench**, don't regress more than `<TBD>`%.
- **Feature** (Full path): at least one load measurement/assessment for the critical flow before merge.
- **Observability**: track p95/p99, token/cost, error rate after going to prod (see PRD §10).

## 4. Checklist by path (feeds the DoD)
- **Fast path**: no obvious performance regression; no new N+1.
- **Standard path**: + before/after bench for code touching a hot path; + (AI) prompt caching preserved/applied.
- **Full path**: + meets the §1 budget; + (AI feature) meets the §2 token/latency ceiling; + a load measurement exists for the main flow.
