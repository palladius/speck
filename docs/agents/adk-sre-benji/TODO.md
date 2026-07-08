# TODO — needs Riccardo's input (Monday)

`SPEC.md` in this directory was drafted from a quick read-only look at
`../adk-sre-benjamin` (README, `conductor/product.md`,
`conductor/tech-stack.md`, `src/agents/` listing, `tests/` listing) plus
our conversation — not a deep code read, given time pressure before the
airport. Before implementing, please answer:

1. **"ADK agents don't work at all"** — what's the actual symptom? Paste the
   exact error/stack trace (or point me at a log) from trying to invoke, e.g.,
   `uv run python3 src/cli.py --agent commander --message "status"`. Without
   this, Step 0 in `SPEC.md` is a guess, and everything downstream (evals
   need working agents to eval) is blocked on it.
2. **CUJ priority list** — `SPEC.md` assumes "GKE cluster is down" (your
   example) and "frontend latency SLO violated" (already simulated by
   `run_simulation.py`) as the first two end-to-end CUJs. Are there others
   you consider must-have for v1 — e.g. a HIGH-risk mutation getting
   correctly blocked by the Logistics Lead?
3. **Real vs. sandboxed GCP for tier-1 fixtures** — for the GKE CUJ, is a
   mocked/fake GKE cluster state acceptable for now, or do you want this
   hitting a real (sandboxed) GCP project from the start? This changes how
   much Step 4 costs to build.
4. **Judge model & budget** — which model should grade the LLM-judge tier
   (`gemini-pro-latest`? something else?), and is there a rough
   budget/frequency preference (e.g. "full CUJ eval only pre-merge, not on
   every commit")?
5. **Does this spec also own fixing "ADK agents don't work at all"**, or is
   that a separate, already-tracked issue (e.g. under `conductor/tracks/`)?
   If the latter, point me at it so this spec doesn't duplicate it.

Once these are answered, this promotes from spec to implementation task —
happy to pick up Step 1 (scaffold `tests/eval/`) right after.
