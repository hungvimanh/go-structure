# Design Standards — UI/UX

> **Standards & how-to for UI/UX.** This file holds *standards, skill routing, design profiles, media, and in-product help* — NOT process sequencing. For *when* each step happens, see `using-hero` and `hero-planning`. Per-project specifics are locked as `<TBD>` in the PRD / Design Brief.

## §1 UX ↔ UI (two layers)
- **UX (problem space):** user personas, primary flows, information architecture, screen inventory, per-screen states (empty / loading / error / success / partial).
- **UI (solution surface):** brand, one locked visual direction, design tokens, component system.

## §2 Design profiles — fit the product
Pick **one** profile per project (decided in the first PRD/TDD), record it in `docs/design/DESIGN_SYSTEM.md`. The profile sets the *creativity × discipline* balance and which steps are required vs skippable.

| Profile | Typical product | Creativity | Discipline | Brand & media | Required steps | Skippable |
|---|---|---|---|---|---|---|
| **system-strict** | internal tools, CRUD, B2B admin | low | very high | minimal; standard icon set, no bespoke media | one locked token set (lean direction, e.g. `minimalist-ui`), strict layout/spacing, component reuse, a11y, consistency QA | brand exploration, imagegen exploration, custom media |
| **branded-product** | end-user SaaS / productivity apps | medium | high | brand tokens + selective bespoke media (empty-states, key illustrations) | direction gate, design system, imagegen for key screens, selective media | heavy marketing media, per-section creative gates |
| **expressive** | consumer, marketing/landing, brand-forward | high | consistent | full brand + rich embedded media + motion | brand, direction exploration, media production, design system, per-section visual review, media-perf QA | — |

Default when the product type is unknown: **branded-product**, confirmed at the gate.

## §3 Design skill routing — job → skill
> Install via the `skills` CLI per `skills.manifest.json` groups. **Pick exactly ONE visual direction** and lock it.
>
> **Availability caveat:** in `skills.manifest.json`, `gpt-taste` (design-direction) and all of `design-tools` (`imagegen-frontend-web`, `imagegen-frontend-mobile`, `image-to-code`, `redesign-existing-projects`) currently have `"source": "<TBD>"` — no installable source is wired yet. Before routing to any of these, verify with the `skills` CLI that the skill actually installed; if it didn't, fall back to a skill with a real source (e.g. `design-taste-frontend`, `minimalist-ui`) or do the step manually. Do not assume the row below means the skill is installable today.

| Need | Skill | When |
|---|---|---|
| Brand / identity (logo, brand board) | `brandkit` | New product/brand or rebrand. Once. (branded-product / expressive) |
| **Lock a visual direction — PICK ONE** | one of `design-taste-frontend` · `minimalist-ui` · `industrial-brutalist-ui` · `high-end-visual-design` · `gpt-taste` | At the Phase-2 gate. Exactly one, then locked. |
| Generate reference images (web) | `imagegen-frontend-web` | Explore/lock the look per section before coding. |
| Generate reference images (mobile) | `imagegen-frontend-mobile` | Mobile screen exploration. |
| Turn an approved design image into code | `image-to-code` | Implementing from an approved visual reference. |
| Redesign / audit an existing UI | `redesign-existing-projects` | Router row #9 on existing surfaces. |
| Interactive wireframes / A-B visual choices | `brainstorming` + visual companion | UX wireframing and layout selection. |

**Pick a direction by vibe:** `minimalist-ui` = clean editorial · `industrial-brutalist-ui` = raw mechanical/terminal · `high-end-visual-design` = expensive agency feel · `gpt-taste` = editorial + GSAP motion · `design-taste-frontend` = anti-slop general default.

## §4 Design tokens & system
- Tokens are the **single source of truth**: color, type scale, spacing, radius, shadow, motion.
- **No hardcoded design values** in components — reference tokens.
- Name tokens consistently (semantic over literal, e.g. `color-surface` not `gray-100`).
- **Pick exactly one direction** for the project; do not mix design skills within one feature.

## §5 Platform
- Declare scope per project: web-only / mobile-only / web+mobile.
- **One shared token source is canonical.** Platform layers add platform-specific patterns:
  - **Web:** responsive breakpoints, hover/focus, pointer + keyboard.
  - **Mobile:** native navigation, platform conventions (HIG / Material), touch targets, safe areas.
- Routing: `imagegen-frontend-web` vs `imagegen-frontend-mobile`.

## §6 Media / asset production
Beyond CSS, products often need embedded media.
- **Asset types:** icons, illustrations, photography/hero imagery, OG/social share images, diagrams, help/onboarding visuals (annotated screenshots, GIF/short video — see §8), optional short motion (lottie/video).
- **Render paths:**
  1. A **configured AI media provider** (provider-agnostic; API key from the consumer's environment) for higher-fidelity raster/illustration. Configure via `hero-mmt-kit init` (stores provider + env-var name only).
  2. **Claude-subagent fallback** (`imagegen-frontend-web/mobile`, `brandkit`) when no provider is configured — produces references, mockups, prompts, and within-capability assets.
- **Storage:** source/reference & exploration → `docs/design/assets/` (durable, reviewable; keep small — large raw files under a gitignored path); production-ready optimized assets → the app's own asset directory (`<TBD>`).
- **Asset DoD:** optimized/compressed, right format (SVG/WebP/AVIF where apt), rights-clear/licensed, alt text, responsive `srcset` where apt.

## §7 Visual QA checklist
- Responsive across declared breakpoints.
- a11y **WCAG AA** baseline: contrast, visible focus, keyboard navigation, alt text, semantic landmarks.
- All states covered (empty/loading/error/success/partial).
- Cross-browser; dark mode (if applicable); `prefers-reduced-motion` respected.
- Media weight within the [PERFORMANCE_STANDARDS.md](./PERFORMANCE_STANDARDS.md) budget.

## §8 In-product help & onboarding
Every product with a UI ships an in-product help surface on a **secondary** surface (under a `?` icon / settings / footer — never the primary functional nav).
- **Pattern library (standardize 4):**
  1. **Dedicated Help/Guide page** — its own section in secondary nav, table of contents by task.
  2. **First-run onboarding** — tour/coachmarks, skippable and replayable.
  3. **Contextual help** — tooltips, inline hints, empty-state guidance where the user needs it.
  4. **FAQ / troubleshooting** — common questions + error resolution, searchable.
- **Depth by profile:** `system-strict` = structured text + screenshots; `branded-product` = + annotated images + empty-state hints; `expressive` = + onboarding tour + GIF/short video.
- **Content source:** authored as Markdown/MDX in `docs/help/` (versioned, reviewable), rendered in-app; mirrors the project language(s); follows the §7 a11y baseline.
- **Sync rule:** a feature with user-facing impact is **not done** until its help entry exists and is accurate.

## §9 Artifacts
Design artifacts and their storage: UX spec (in the PRD), `docs/design/DESIGN_SYSTEM.md`, `docs/design/assets/`, `docs/help/`, and the Visual QA report (part of `hero-reviewing`'s output).

## §10 References-first (avoid AI-slop)
- AI interfaces look generic mainly because no real references were given. Gather references **before** generating.
- Collect 3–5 real references (Dribbble, Mobbin, Behance, Muzli); paste/link them and **name the products** in the Design Brief. Note what specifically you want from each (layout, type, color, motion).
- Generate multiple variations and compare; then lock exactly one direction (see §4 token/lock rules).
- Avoid LLM defaults: AI-purple gradients, centered hero over dark mesh, three equal feature cards, generic glassmorphism, Inter + slate-900 everywhere.
- **Optional design MCP** (license-respecting, not bundled): a browser MCP (Chrome/Playwright) to capture references and validate the built UI via screenshots; a Figma MCP to turn an existing Figma design into code. Provider-agnostic, like media generation (§6).
- **Optional storytelling layer** (expressive profile): give the page a narrative — scroll-story, parallax, sequenced reveals. Record the intent briefly in `docs/design/design-story.md`; keep effects tasteful and within the performance budget (§7).
