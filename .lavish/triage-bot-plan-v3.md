# k8s-snap Triage Bot — v3 Design Plan (LangGraph)

## 0. What changed from v2 -> v3 (your decisions)
- **Q1:** ONE LangGraph workflow; ship the **triage node(s)** first, extend to reproducer/fixer nodes later in the SAME graph (not a separate bot).
- **Q2:** **LangGraph**, not Claude Code Action. Provider-agnostic via a `make_llm("provider/model")` factory; **start with Gemini**.
- **Q3:** self-hosted runners (unchanged).
- **Q4:** provider-independent from day one (Gemini default, swap by config/env).
- **Q5:** **no Jira**.
- **Q6:** must be **locally invocable + unit-testable**, integrated into the existing **`ci/k8s-ci.py`** CLI as a subcommand.
- **Inspiration:** `canonical/slupgrader` (itself a provider-agnostic multi-agent LangGraph app) — copy its logging/observability + LLM-factory + node-wrapper + offline-test seams.

## 1. Repo grounding (verified)
- `ci/k8s-ci.py` = argparse dispatcher: `add_<group>_cmds(subparsers)` -> each `p.set_defaults(func=cmd_*)`, `cmd_*(args)->int`. Invoked via `tox -e k8s-ci -- <args>`.
- House style in `ci/`: stdlib only, `print(..., file=sys.stderr)` for progress, `--dry-run` everywhere, `_set_output()` for GitHub Actions. **No `logging` module used yet** -> our structured logger is additive, not a conflict.
- Deps compiled via `pip-compile` from `ci/requirements-ci.in` (hash-pinned). We add a NEW isolated extra so the base CLI stays lean.

## 2. slupgrader patterns we adopt (source-verified)
- **LLM factory** `llm.py::make_llm("provider/model")` -> `BaseChatModel`; default `gemini/gemini-flash-latest`; lazy per-provider imports; `extract_token_usage()` unifies Google/OpenAI/Anthropic token metadata.
- **StateGraph** over a `TypedDict`; nodes are `state -> dict` partial updates; cumulative metrics via `Annotated[..., operator.add]` reducers.
- **`make_validated_node(name, fn)`** wrapper (validate.py:118) — the seam where we emit ONE structured record per node transition.
- **Observability = LangGraph callbacks** threaded via `app.invoke(state, config={"callbacks":[...], "run_name":..., "metadata":{...}})`; opt-in Langfuse/OTEL/LangSmith, isolated as packaging extras.
- **`CredentialRedactor`** (util/redact.py) — value-substitution of known secrets in tool output / logs.
- **Offline tests**: `patch(make_llm)` + a fake agent returning canned `AIMessage`; `--no-llm` mechanical path for deterministic runs.
- **Exit-code enum** (FIXED/EXHAUSTED/API_ERROR/BAILED_OUT/...).

## 3. Where slupgrader is a WEAK precedent (we improve)
- Its logging has **no run_id / node_id on records** and **no JSON sink** — you can't isolate one run from plaintext logs. We add a `run_id`+`node` via a `logging` filter/adapter and an optional JSON formatter, so the node-by-node execution log is queryable offline (not only in a hosted trace UI).

(Full sections 4-11 assembled in the artifact.)
