---
speck_version: 0.1
mode: manual
created_at: 2026-07-01
---

# speck: idea → SPEC.md

This is speck's own spec, written by hand (dogfooding the section structure
that `speck` itself generates for other projects — see
[META-SPECS.md](META-SPECS.md) for why it's shaped this way).

## Problem Statement

Turning a rough idea into something an AI coding assistant can reliably
implement usually means writing a spec by hand, or dumping the idea straight
into an assistant and hoping the resulting back-and-forth gets captured
somewhere durable. There's no small, fast, provenance-tracked tool dedicated
to just the "idea → spec" step, decoupled from any particular coding
assistant.

## Goals

- Turn a short idea into a well-structured `SPEC.md` in one command
  (`speck oneshot`).
- Optionally drive a clarifying-questions interview (`speck chat`) when the
  idea needs to be fleshed out first, using the interview philosophy in
  [META-SPECS.md](META-SPECS.md).
- Accept the idea from wherever is convenient: CLI arg, `--file`, piped
  stdin, `./IDEA.md`, or an interactive multi-line prompt.
- Optionally enrich the prompt with existing notes via
  `--source-of-inspiration <dir>` (scans `.txt`/`.md`/`.html` files as
  secondary context).
- Auto-decide a sensible `<category>/<slug>/SPEC.md` destination under a
  configurable output directory (`--output-dir`/`-o`, `$SPECK_OUTPUT_DIR`,
  or `./out` by default) instead of always writing to `./SPEC.md`.
- Record full provenance in the output file's YAML frontmatter: a pointer to
  the exact idea (`input_prompt.md`, always written alongside the spec),
  inspiration sources used, timestamp, model, and token usage.
- Load `GEMINI_API_KEY` and other config from a `.env` file automatically,
  so common invocations need no manual `export`.
- Be genuinely pleasant to use from a terminal: colors, emoji, shell
  autocompletion.

## Non-Goals

- speck does not implement the resulting spec — that's for whichever AI
  coding assistant the user feeds `SPEC.md` to.
- speck does not manage multi-session "Goldfish" validation, critic review,
  or implementation-readiness passes in v1 (see META-SPECS.md).
- speck is not a multi-provider LLM abstraction — v1 is Gemini-only.

## Technical Plan / Approach

- **Language**: Go, single static binary, no runtime dependencies.
- **LLM backend**: `google.golang.org/genai`, the official Gemini Go SDK,
  authenticated via the `GEMINI_API_KEY` environment variable.
- **CLI framework**: `spf13/cobra` — gives shell completion for free.
- **Styling**: `charmbracelet/lipgloss` for colored output; plain Unicode
  emoji for status markers. No full TUI framework — the interaction model
  is sequential stdin/stdout prompts, not a screen-based UI.
- **Structured output**: both `oneshot` and `chat` request JSON from Gemini
  (`category`, `slug`, `title`, `spec_body`) via `responseSchema`, so speck
  can reliably place the file rather than parsing free-form markdown.
- **Frontmatter**: YAML, via `gopkg.in/yaml.v3`.
- **Config**: `.env` files are loaded via `internal/dotenv` (explicit
  environment variables still take precedence).

## Alternatives Considered

- **Python instead of Go** — faster to prototype, but the user explicitly
  chose Go for a single quick binary, with Python as the documented fallback
  if Go proved painful.
- **Shelling out to the `claude`/`gemini` CLI** instead of calling the API
  directly — rejected because the interactive mode needs fine-grained
  control over a multi-turn loop (phase transitions, structured JSON output)
  within a single process.
- **Bare `SPEC.md` in cwd** instead of an auto-chosen `<category>/<slug>/`
  folder — rejected per explicit request; auto-categorization makes it
  natural to accumulate many specs (`games/tetris-clone/`,
  `tools/log-dedup/`, …) without manual folder bookkeeping.

## Implementation Plan

- `cmd/speck/main.go` — entrypoint wiring the cobra root command.
- `internal/cli/` — `root.go` (shared flags), `oneshot.go`, `chat.go`.
- `internal/genai/` — thin wrapper over the SDK: client construction,
  structured-JSON generation, token usage accumulation.
- `internal/inspire/` — scans `--source-of-inspiration` directories.
- `internal/input/` — resolves the idea text (arg / file / stdin / `IDEA.md`
  / interactive prompt).
- `internal/spec/` — frontmatter (de)serialization, body template section
  headers, output path resolution + collision handling.
- `internal/dotenv/` — minimal `.env` loader.
- `internal/version/` — single source of truth for `--version` and
  `CHANGELOG.md` entries.
- `internal/ui/` — lipgloss styles and emoji constants.
- `docs/SPECS.md`, `docs/META-SPECS.md`, `README.md`, `CHANGELOG.md` — this
  document, the interview-philosophy rationale, user-facing usage docs, and
  version history.

## Open Questions

- Should `speck` eventually support other model providers, or stay
  Gemini-only by design?
- Should the deferred Goldfish/Critic/Readiness passes become a `speck
  review` subcommand, and if so, does that want its own spec first?
- Is a single flat `<category>/<slug>/` layout enough, or will deeper
  nesting be needed once there are many specs?
