---
title: "Step 0 — Installation"
description: "Install reqmd in 30 seconds with go install, build from source, or download a prebuilt binary. Then verify the install against the bundled 249-requirement spec."
weight: -1
---

Three ways to get reqmd on your machine. Pick one.

## Option 1 — `go install` (fastest)

If you already have Go on your `$PATH`:

```sh
go install github.com/dVoo/reqmd/cmd/reqmd@latest
```

The binary lands in `$(go env GOBIN)` (or `$(go env GOPATH)/bin`).

## Option 2 — Build from source

```sh
git clone https://github.com/dVoo/reqmd
cd mdreq
go build -o reqmd ./cmd/reqmd
./reqmd --help
```

Works on every platform Go supports — Linux, macOS, Windows, \*BSD.

## Option 3 — Prebuilt binary

Download a release from [GitHub Releases](https://github.com/dVoo/reqmd/releases). Extract the archive and put `reqmd` on your `$PATH`.

## Prerequisites

- **Go 1.25 or newer** — required for the source build.
- **Cgo toolchain** (gcc or clang) — only if you build the optional `export graph` subcommand with the `ladybug` build tag.

## With graph export (optional)

`reqmd export graph` writes a LadybugDB database for Cypher queries. It pulls in Cgo and an extra dependency, so it's gated behind a build tag:

```sh
go build -tags ladybug -o reqmd ./cmd/reqmd
```

Without the tag, `reqmd export graph` is not built. The rest of the CLI works exactly the same.

## Verify the install

```sh
reqmd --version         # shows the version
reqmd check spec/       # validate the bundled dogfooded spec (249 requirements)
```

If the second command prints `Summary: 249 total, 249 valid, 0 invalid, 0 parse errors`, you're good to go.

## What's next

Now write your first requirement — go to [Step 1](../01-get-started/).