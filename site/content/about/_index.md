---
title: "About"
description: "reqmd's design principles, license, and the self-hosting story."
weight: 8
---

reqmd is a Go CLI for writing, validating, and exporting requirement specs in plain Markdown. It's MIT-licensed, developed in the open, and dogfoods its own format for its own spec.

## Design principles

- **No lock-in.** Your data is plain Markdown and YAML — openable in any editor, renderable on GitHub/GitLab, diffable with standard Git tools.
- **No database.** The trace graph is ephemeral, built in a temp directory on each invocation and discarded when done. The `*.md` files are the single source of truth — nothing is committed except text files.
- **Per-directory schemas.** Different subsystems (IVI, safety, system) can have different required fields and validation rules, all validated in one pass.
- **Scale.** Parses thousands of files in parallel using a `runtime.NumCPU()` worker pool.
- **Dogfooding.** The reqmd tool's own requirements are defined in `spec/` using the reqmd format and validate with `reqmd check spec/`. This ensures the format is always production-ready for the team's own use.

## Self-hosted requirements

ReqMD dogfoods its own format. The `spec/` directory contains a complete 6-level V-model tree defining the tool itself:

| Level | Directory | Reqs | Description |
|-------|-----------|:----:|-------------|
| External | `spec/00-aspice/` | 191 | Reference base practices (`external: true`) |
| Stakeholder | `spec/01-stakeholder/` | 5 | Stakeholder goals (top boundary, no upstream) |
| Stakeholder mapping | `spec/01a-aspice-stakeholder/` | 18 | Stakeholder requirements mapped to reference practices |
| System | `spec/02-system/` | 9 | Feature specifications |
| Software | `spec/03-software/` | 6 | Component-level design |
| Tests | `spec/04-tests/` | 4 | Test specifications (mandatory-disposition) |

All 833 (or 233, depending on whether you count imported proxy reqs) requirements validate cleanly:

```sh
$ reqmd check spec/
# ... 833 total, 833 valid, 0 invalid, 0 parse errors, 541 warnings

$ reqmd serve spec/          # live-reloading HTML preview
$ reqmd export html spec/ -o /tmp/out/
```

The 541 warnings are expected — external reference practices (`00-aspice`) and stakeholder goals (`01-stakeholder`) have no upstream traces (untraced warnings suppressed by boundary inference and `external: true`). The stakeholder mapping layer (`01a-aspice-stakeholder`) is the first fully-traced tier.

## License

MIT — see [`LICENSE`](https://github.com/dVoo/reqmd/blob/main/LICENSE) on GitHub. The bundled Inter font is OFL-licensed; the chroma styles are MIT/BSD.

## Contributing

Issues, PRs, and design discussions are all on [GitHub](https://github.com/dVoo/reqmd). The codebase is small (~5 000 lines of Go across 8 packages) and the trace model is well-documented in `AGENTS.md`.

For substantial changes, open an issue first to discuss. The `spec/` tree is the design surface — every change to the model or the CLI is reflected there as a requirement update.
