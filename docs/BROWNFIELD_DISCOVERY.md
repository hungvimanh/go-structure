# Brownfield Discovery

Use this when the project already has code and may have partial, outdated, or non-standard documentation.

## When to run

For an existing project that has never used hero-mmt-kit:

```bash
npx hero-mmt-kit init
npx hero-mmt-kit discover
npx hero-mmt-kit doctor
```

`init` installs the workflow. `discover` scans the current repo and creates `docs/BROWNFIELD_DISCOVERY.md` with the first evidence map.

## What the AI must do after `discover`

1. Read `docs/BROWNFIELD_DISCOVERY.md` first.
2. Start with the summary and **Read next** list; do not load every discovered doc/config by default.
3. Read only the docs/config/code needed for the selected task path, then expand context if evidence is insufficient.
4. Record findings with certainty labels:
   - **Found** — directly observed in files.
   - **Likely** — inferred from names/structure, not yet proven.
   - **Needs confirmation** — requires the human or missing business context.
5. For the first code changes after discovery, use `using-hero` to pick the right skill. If it touches > 2 files or any **Needs confirmation** area, use Plan Mode (`hero-planning`) + impact analysis before implementing, and choose the review budget from actual risk and uncertainty.

## Context discipline

- Do not assume docs live in `docs/`.
- Do not assume missing docs mean missing behavior.
- Do not assume a folder name proves its role before reading files.
- Do not claim a command passes until it has been run.
- Do not paste long source excerpts into the report; use file paths, short evidence notes, and links to artifacts.

## Output expected from the first AI pass

Update the discovery report or create a report under `docs/reports/` with a summary-first structure:

1. **Summary** — project purpose, important flows, and likely architecture in 3–6 bullets.
2. **Read next** — prioritized files/docs/configs with one-line reasons.
3. **Evidence map** — key paths and what was found there.
4. **Verification commands** — commands observed or proven, clearly labeled.
5. **Risks / unknowns** — areas needing confirmation before edits.
6. **Open questions** — only questions that block safe next steps.
