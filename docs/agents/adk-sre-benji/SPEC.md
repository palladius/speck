---
speck_version: "0.1"
mode: manual
created_at: 2026-07-08
---

# ADK SRE Benjamin: Guardrails & Testing

Spec for testing `../adk-sre-benjamin` — a 5-agent Google ADK incident-response
system — at both the individual-agent and end-to-end level. Drafted from a
quick read-only look at that repo (README, `conductor/product.md`,
`conductor/tech-stack.md`, `src/agents/` listing, `tests/` listing) plus
conversation with Riccardo. **See `TODO.md` in this directory for open
questions that need his input before implementation starts.**

## Problem Statement

Project Benjamin is a production-grade SRE incident-response framework built
on Google ADK, implementing Google's IMAG Incident Command System (ICS): a
strict top-down, hierarchical, non-conversational delegation model across
five agent personas — Incident Commander ("Benjamin"), Operations Lead,
Planning Lead, Logistics Lead, and Communications Lead ("Madhavi").

Two acute problems:

1. **ADK agents reportedly do not run at all right now.** The root cause is
   undiagnosed as of this writing — see `TODO.md` item 1.
2. **Even once working, there's no systematic way to evaluate agent quality.**
   The existing test suite (30+ pytest files, including `test_orchestrator.py`,
   `test_integration.py`, `test_agents.py`, and the `run_simulation.py`
   harness) covers deterministic plumbing well — trigger parsing, incident
   scaffolding, safety-command parsing, registry/artifact bookkeeping — but
   likely mocks or bypasses live LLM reasoning, so it can't catch failures in
   actual agent judgment (bad delegation decisions, safety-boundary
   violations) or in the real ADK runtime integration.

This is the same "compiles but doesn't work" gap described in
[`docs/meta/executable-spec-framework/SPEC.md`](../../meta/executable-spec-framework/SPEC.md),
applied specifically to a multi-agent system where "does it work" means both
"does each agent behave well" and "does the handoff between agents produce
the right outcome."

## Goals

- **Baseline liveness**: verify each of the 5 ADK agents (`commander.py`,
  `comms.py`, `ops_lead.py`, `planning_lead.py`, `logistics_lead.py`) can
  actually be invoked and produce a valid LLM response. This is the
  prerequisite for everything else and is currently broken (TODO item 1).
- **Individual-agent evals** via `agents-cli eval` (see this repo's own
  `google-agents-cli-eval` skill): for each lead, a dataset + metrics judging
  tool selection, instruction-following, and — critically — role adherence.
  E.g. Operations Lead must stay strictly read-only and delegate mutations
  rather than execute them; that's a safety invariant, not just a quality
  score.
- **End-to-end CUJ evals**: full-trace scenarios capturing every lead's turn
  (`agent_data.agents` / `turns[].events[]` per the ADK eval schema), graded
  holistically for whether the incident was actually handled correctly.
  Starting CUJs:
  1. "GKE cluster is down" — start an investigation, observe every actor do
     the right thing (Riccardo's own example).
  2. "Frontend latency SLO violated" — already simulated by
     `run_simulation.py`; reuse as the first e2e eval case since the fixture
     already exists.
- **Guardrails as testable invariants, not prose.** Turn the IMAG ICS role
  boundaries into checks that run against a trace, not just instructions in a
  prompt:
  - Operations Lead never directly executes a mutation (must delegate to the
    Mutation Agent).
  - Logistics Lead blocks HIGH-risk commands without explicit approval.
  - Communications Lead is the *only* agent that sends outbound
    Telegram/GitHub messages.
- **Tiered cost/speed model**, per the Executable Spec Framework: cheap
  deterministic guardrail checks run on every commit; expensive LLM-judged
  CUJ evals run sparingly (pre-merge / nightly).

## Non-Goals

- Diagnosing *why* ADK agents don't currently run — that needs real error
  output from Riccardo (see `TODO.md` item 1). This spec defines the target
  testing/guardrail architecture assuming that gets fixed first or in
  parallel, not how to fix it.
- Replacing the existing 30+ pytest deterministic tests. Those stay as the
  fast tier-2 layer; this spec adds the missing individual-agent and e2e
  LLM-judged tiers on top.
- Redesigning agent orchestration. The existing "Pure ADK Hierarchical Tree,"
  strictly non-conversational, top-down delegation pattern is taken as a
  given constraint, not something to change.

## Technical Plan / Approach

- Use `agents-cli eval` as the eval engine rather than hand-rolling one:
  - `tests/eval/datasets/` — per-agent single-turn/multi-turn cases for each
    of the 5 leads.
  - `agents-cli eval generate` runs the *real* ADK agent (no LLM mocking)
    against each dataset, writing traces to `artifacts/traces/`.
  - `agents-cli eval grade` scores with a mix of built-ins
    (`multi_turn_tool_use_quality`, `multi_turn_trajectory_quality`) and
    custom metrics for the guardrails above.
- For e2e CUJs, adapt `run_simulation.py` (or a variant) to also emit
  ADK-eval-shaped traces — `agent_data.agents` populated with all 5 leads,
  `turns[].events[]` recording each lead's actions — so the same run feeds
  both the existing file-based audit trail (`state.md` / `timeline.md`) and
  `agents-cli eval grade`.
- **Tier-1 fixture** for the GKE CUJ: a mocked/sandboxed GKE cluster state so
  the scenario runs deterministically and repeatably without touching a real
  cluster (see `TODO.md` item 3 on whether this should instead be a real
  sandboxed GCP project from the start).
- Prefer `CodeExecutionMetric` (deterministic Python inspecting the trace —
  e.g. "no `function_call` from `ops_lead` names a mutation tool") for
  guardrails, since these are inspectable facts. Reserve `LLMMetric` judges
  for genuinely fuzzy judgment (e.g. "was the Incident Commander's delegation
  decision reasonable given this alert").

## Alternatives Considered

- **Hand-rolled eval harness instead of `agents-cli eval`** — rejected;
  reinvents dataset schema, trace capture, and grading that already exist and
  are actively maintained.
- **Skip individual-agent evals, only test the orchestration** — rejected: a
  failing e2e CUJ alone doesn't say *which* of the 5 agents is at fault.
  Per-agent and cross-agent-handoff failures are different signals and need
  separate coverage.
- **Mock all LLM calls in eval, like the existing pytest suite likely does**
  — rejected as the primary signal. That's exactly the "compiles but doesn't
  work" trap this whole exercise exists to close. Mocked tests stay as fast
  pre-commit smoke tests; the eval tier must exercise real agent reasoning.

## Implementation Plan

- **Step 0 (blocking).** Diagnose why ADK agents don't currently run at all
  — needs Riccardo's error output (`TODO.md` item 1).
- **Step 1.** Scaffold `tests/eval/` per the `agents-cli` convention; one
  single-turn dataset per lead (5 datasets), 2-3 cases each.
- **Step 2.** Run `agents-cli eval generate && agents-cli eval grade`
  against each with built-in metrics first, to get a real baseline —
  don't assume anything passes yet.
- **Step 3.** Write the guardrail custom metrics (role-boundary /
  mutation-without-clearance / comms-exclusivity) as `CodeExecutionMetric`s
  in `tests/eval/eval_config.yaml`.
- **Step 4.** Adapt `run_simulation.py` to also produce an ADK-eval-shaped
  trace; add "GKE cluster is down" and "frontend latency SLO violated" as
  `eval_dataset` cases (`eval dataset synthesize` may help for the GKE one,
  since no fixture exists yet).
- **Step 5.** Wire the fast guardrail subset into CI/pre-commit; wire the
  full CUJ + LLM-judge tier into a nightly or pre-merge job.

## Open Questions

See [`TODO.md`](TODO.md) in this directory — these need Riccardo's input
before implementation starts.
