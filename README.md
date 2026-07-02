# speck

Turn a rough idea into a structured `SPEC.md` you can hand to any AI coding
assistant.

```
speck oneshot "a tetris clone with a twist: gravity reverses every 30s"
✨ thinking...
✅ wrote games/tetris-clone-gravity-flip/SPEC.md
```

## Why

Specs are the durable artifact; code is increasingly just a rendering of
them. `speck` is a small, fast tool dedicated to the "idea → spec" step —
decoupled from whichever assistant eventually implements it. See
[`docs/SPECS.md`](docs/SPECS.md) for speck's own (self-hosted) spec, and
[`docs/META-SPECS.md`](docs/META-SPECS.md) for the interview philosophy
behind `speck chat`.

## Install

Requires Go 1.22+.

```
go install github.com/palladius/speck/cmd/speck@latest
```

## Setup

```
export GEMINI_API_KEY=...   # https://ai.google.dev/gemini-api/docs/api-key
```

## Usage

### One-shot

Generate a spec directly from an idea:

```
speck oneshot "a CLI that dedups photos by perceptual hash"
```

The idea can come from several places, in priority order: a positional
argument, `--file <path>`, piped stdin, a `./IDEA.md` in the current
directory, or — if none of those are given — an interactive multi-line
prompt (press Enter twice, or Ctrl-D, to finish).

### Interactive

Let speck interview you before writing the spec:

```
speck chat "a tool to dedup photos"
```

speck asks one focused clarifying question at a time, first about the
*problem* and then about *acceptance criteria* (see
[`docs/META-SPECS.md`](docs/META-SPECS.md)). Type `/done` at any prompt to
skip ahead. The full transcript is saved next to the generated spec.

### Inspiration material

Feed existing notes in as extra context (not the primary idea):

```
speck oneshot "a photo dedup tool" --source-of-inspiration ./notes
```

Scans `.txt`, `.md`, and `.html` files under the given directory.

## Output

speck decides where to write the spec: `<category>/<slug>/SPEC.md`, e.g.
`games/tetris-clone/SPEC.md`. Override the base directory with `--base-dir`,
or force-overwrite an existing spec with `--force`.

Every generated `SPEC.md` carries YAML frontmatter for provenance:

```yaml
---
speck_version: 0.1
mode: oneshot
idea: |
  a CLI that dedups photos by perceptual hash
created_at: 2026-07-01T15:20:00Z
model: gemini-flash-latest
tokens:
  prompt: 1234
  output: 567
  total: 1801
---
```

Long ideas are moved to a sidecar `original_idea.md` instead of being
embedded inline; interactive sessions also get a `speck_transcript.md`
sidecar with the full Q&A.

## Flags

| Flag | Description |
|---|---|
| `--model` | Gemini model to use (default: `gemini-flash-latest`; use `gemini-pro-latest` for a smarter but slower model) |
| `--api-key` | Overrides `GEMINI_API_KEY` |
| `--base-dir` | Base directory for the `<category>/<slug>/` output (default: cwd) |
| `--source-of-inspiration`, `-s` | Directory of `.txt`/`.md`/`.html` files to use as supporting context |
| `--file` | Read the idea from a file |
| `--force` | Overwrite an existing `SPEC.md` at the resolved path |

## Shell completion

```
speck completion bash|zsh|fish|powershell
```
