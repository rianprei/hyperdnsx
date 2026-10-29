# HyperDNSx — engineering principles

## Ponytail / YAGNI

**Climb the ladder** — before writing code, take the first rung that works:

1. Does it already exist? → use it
2. Reuse something local in this repo
3. Standard library
4. Native platform capability
5. An already-installed dependency
6. A one-liner
7. Minimum new code

**Deletion over addition.** Removing code is a valid, preferred fix.

**No speculative scaffolding. No unrequested abstractions.** Build what was
asked, not what might be needed later.

**Before editing, list the blast radius:** callers, tests, fixtures, configs,
exports affected.

**Never cut:** trust-boundary validation, data-loss error handling, or the
requested scope. These are not candidates for trimming.

**Non-trivial logic gets a small test or assert.**

**Known-limit shortcuts are explicit:** annotate with
`shortcut: <limit>, upgrade when <condition>`.

**Output code/facts first; close with what was skipped or left unverified.**

## Applying this here

- The httpx engine is a **library dependency**, not a source fork. Recovered
  fork features are added as thin, additive modules that reuse the engine.
- Prefer exposing capabilities httpx already has (CdnCheck, TLS grab, vhost
  options) over reimplementing them.
- New binary `hyperx` is separate; the pure-DNS `dnsx` surface is not touched.
- See `docs/deep-research.md` for the architecture and fork-feature roadmap.
