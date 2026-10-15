#!/usr/bin/env python3
"""Import a DOORS or codeBeamer CSV export into a reqmd-import/v1 bundle.

One generic reader covers both tools, because both export rows with an item
id, a heading, a description, optional parent, and optional link columns.
Column names are flags, so no code changes are needed per tool.

Example:

    python examples/csv_import.py doors-brakes.csv -o brakes.json \\
        --dir brakes --id-prefix DOORS- --title "Brakes Module" \\
        --level stakeholder-requirements --tool doors \\
        --id-column "Absolute Number" --title-column "Object Heading" \\
        --body-column "Object Text" --parent-column Parent \\
        --trace-column Links --attr priority=Priority

    reqmd-import brakes.json -o spec/00-doors/

Then validate as usual:  reqmd check spec/
"""

from __future__ import annotations

import argparse
import csv
import re
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from reqmd_io import Bundle, Document, Requirement  # noqa: E402

# A fully-qualified reqmd ID (optionally doc-qualified, optionally version-pinned)
# as written by hand in the spec tree. Anything else in a link column is treated
# as a native id of another item in this import.
REQMD_ID = re.compile(r"^(?:[a-z0-9_-]+/)?[A-Z][A-Z0-9-]+(?:~\d+)?$")


def parse_args(argv: list[str]) -> argparse.Namespace:
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("csv", help="CSV export from DOORS/codeBeamer")
    p.add_argument("-o", "--output", default="-", help="output bundle path ('-' = stdout)")
    p.add_argument("--dir", default="", help="output document directory (default: bundle root)")
    p.add_argument("--id-prefix", required=True, help='reqmd ID prefix, e.g. "DOORS-"')
    p.add_argument("--title", default="", help="document title")
    p.add_argument("--level", default="", help="x-reqmd.level for the document")
    p.add_argument("--external", action="store_true", help="mark the document x-reqmd.external (a proxy for the ALM)")
    p.add_argument("--id-column", default="ID", help="column holding the item's stable id")
    p.add_argument("--title-column", default="Name", help="column holding the heading")
    p.add_argument("--body-column", default="Description", help="column holding the body text")
    p.add_argument("--parent-column", default="", help="column holding the parent item's id")
    p.add_argument("--trace-column", default="", help="column holding link targets (';' or '|' separated)")
    p.add_argument("--attr", action="append", default=[], metavar="NAME=COLUMN",
                   help="map a CSV column to an attr-block attribute; repeatable")
    return p.parse_args(argv)


def parse_attrs(specs: list[str]) -> dict[str, str]:
    attrs: dict[str, str] = {}
    for spec in specs:
        name, _, column = spec.partition("=")
        if not name or not column:
            raise SystemExit(f"invalid --attr {spec!r}: expected NAME=COLUMN")
        attrs[name] = column
    return attrs


def split_links(cell: str) -> list[str]:
    return [part.strip() for part in re.split(r"[;|]", cell or "") if part.strip()]


def main(argv: list[str] | None = None) -> int:
    args = parse_args(sys.argv[1:] if argv is None else argv)
    attr_columns = parse_attrs(args.attr)

    doc = Document(
        id_prefix=args.id_prefix,
        dir=args.dir,
        file="package.md",
        title=args.title or "Imported requirements",
        notice="This file is auto-generated. Do not edit by hand.",
        level=args.level,
        external=args.external,
        # Every mapped attr is a free-form string; declare it so the generated
        # schema accepts the attr block.
        properties={name: {"type": "string"} for name in attr_columns},
    )
    bundle = Bundle([doc])

    try:
        with open(args.csv, newline="", encoding="utf-8-sig") as fh:
            rows = list(csv.DictReader(fh))
    except OSError as exc:
        print(f"cannot read {args.csv}: {exc}", file=sys.stderr)
        return 1

    for row in rows:
        native_id = (row.get(args.id_column) or "").strip()
        if not native_id:
            continue  # skip section/heading-only rows

        trace, native_trace = [], []
        if args.trace_column:
            for link in split_links(row.get(args.trace_column, "")):
                (trace if REQMD_ID.match(link) else native_trace).append(link)

        req = Requirement(
            native_id=native_id,
            title=(row.get(args.title_column) or "").strip(),
            body=(row.get(args.body_column) or "").strip(),
            parent=(row.get(args.parent_column) or "").strip(),
            trace=trace,
            native_trace=native_trace,
            attrs={name: (row.get(col) or "").strip() for name, col in attr_columns.items()},
        )
        doc.add(req)

    bundle.write(args.output)
    print(f"csv_import: {len(doc.requirements)} requirement(s) from {args.csv}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
