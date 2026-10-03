---
title: "Step 0 — Installation"
description: "Install reqmd in 30 seconds from a prebuilt binary or a source build, then verify the install against the bundled 274-requirement spec."
weight: -1
---

Two ways to get reqmd on your machine. Pick one.

## Option 1 — Prebuilt binary (easiest)

Download a release from [GitHub Releases](https://github.com/dVoo/reqmd/releases). Extract the archive and put `reqmd` on your `$PATH`.

## Option 2 — Build from source

Requires Go 1.26 or newer.

```sh
git clone https://github.com/dVoo/reqmd
cd reqmd
go build -o reqmd ./cmd/reqmd
./reqmd --help
```

Works on every platform Go supports — Linux, macOS, Windows, \*BSD.

To install into `$(go env GOPATH)/bin` instead of building in place:

```sh
go install ./cmd/reqmd
```

> `go install <module>@latest` is not available: the Go module is named
> `reqmd` (not the GitHub path), so `@latest` cannot resolve it. Install from
> a clone, or use a prebuilt binary.

## Prerequisites

- **Go 1.26 or newer** — required for the source build.
- **Cgo toolchain** (gcc or clang) — only if you build the optional `export graph` subcommand with the `ladybug` build tag.

## With graph export (optional)

`reqmd export graph` writes a LadybugDB database for Cypher queries. It pulls in Cgo and an extra dependency, so it's gated behind a build tag:

```sh
go build -tags ladybug -o reqmd ./cmd/reqmd
```

Without the tag, `reqmd export graph` prints an error and does nothing. The rest of the CLI works exactly the same.

## Verify the install

```sh
reqmd check spec/       # validate the bundled dogfooded spec (274 requirements)
```

The last line should read:

```text
Summary: 274 total, 274 valid, 0 invalid, 0 parse errors, 23 warnings
```

The warnings are expected — they come from the external ASPICE base practices in `spec/00-aspice/`, which have no upstream traces.

## What's next

Now write your first requirement — go to [Step 1](../01-get-started/).
