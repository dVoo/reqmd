"""reqmd_io — build reqmd exchange-format bundles from Python.

A thin, dependency-free helper for writing reqmd importers. It only produces
the JSON exchange format; reqmd-import is the single renderer that turns that
into reqmd documents (schema.yaml + .md). This library deliberately contains
no Markdown or YAML logic, so it cannot drift from the Go renderer.

Typical importer::

    from reqmd_io import Bundle, Document, Requirement, Section

    doc = Document(key="brakes", title="Brakes Module",
                   x_reqmd={"id-prefix": "DOORS-", "level": "stakeholder-requirements"},
                   properties={"priority": {"type": "string"}},
                   required=["priority"])

    doc.add(Section(title="Hydraulics", body="The hydraulic circuit."))
    for item in doors_module.objects:
        doc.add(
            Requirement(key=item.number, title=item.heading, body=item.text,
                        status=item.status, fields={"priority": item.priority},
                        links=[link.target_key for link in item.links],
                        refs=[link.external_id for link in item.external_links]),
            parent="Hydraulics",
        )

    Bundle([doc]).write("doors-brakes.json")

Then::

    reqmd-import doors-brakes.json -o spec/00-doors/
"""

from __future__ import annotations

import json
import sys
from dataclasses import dataclass, field
from typing import Any

SCHEMA_URL = "https://reqmd.dev/exchange/v1/schema.json"

KIND_REQUIREMENT = "requirement"
KIND_SECTION = "section"

__all__ = [
    "SCHEMA_URL",
    "KIND_REQUIREMENT",
    "KIND_SECTION",
    "Bundle",
    "Document",
    "Item",
    "Requirement",
    "Section",
]


def _clean(data: dict[str, Any]) -> dict[str, Any]:
    """Drop empty values so the emitted bundle stays small (mirrors the Go
    side's `omitempty`)."""
    return {k: v for k, v in data.items() if v not in ("", None, [], {})}


@dataclass
class Item:
    """Base class for one content item.

    ``kind`` is the only difference between content and structure: a
    requirement renders a reqmd block, a section renders structure only. Items
    nest through ``children``; the flat ``parent`` argument of
    :meth:`Document.add` is resolved into that tree, so a producer that holds a
    flat table never has to build it by hand.
    """

    kind: str = KIND_REQUIREMENT
    key: str = ""
    title: str = ""
    body: str = ""
    fields: dict[str, Any] = field(default_factory=dict)
    status: str = ""
    links: list[str] = field(default_factory=list)
    refs: list[str] = field(default_factory=list)
    children: list[Item] = field(default_factory=list)

    def to_dict(self) -> dict[str, Any]:
        data = {
            "kind": self.kind,
            "key": self.key,
            "title": self.title,
            "body": self.body,
            "fields": self.fields,
            "status": self.status,
            "links": self.links,
            "refs": self.refs,
            "children": [c.to_dict() for c in self.children],
        }
        return _clean(data)


@dataclass
class Requirement(Item):
    """One imported requirement, identified by the source tool's own stable id.

    ``key`` is required: the rendered reqmd ID is the owning document's
    ``id-prefix`` + ``key``. ``links`` name other items by key (``"17"`` in the
    same document, ``"brakes/17"`` across documents). ``refs`` hold external
    requirement IDs and are passed through verbatim.
    """

    kind: str = KIND_REQUIREMENT


@dataclass
class Section(Item):
    """Structure: a heading with prose and children, never a reqmd block.

    A section carries no ``fields`` and no ``status``; both are rejected by the
    renderer. An untitled section renders as a plain prose block.
    """

    kind: str = KIND_SECTION


@dataclass
class Document:
    """One reqmd document: a schema plus, when it has items, one Markdown file.

    ``key`` and ``title`` are required. Everything reqmd-specific lives in
    ``x_reqmd`` (``id-prefix``, ``level``, ``document-id``, ``upstream``,
    ``requires-trace-from``, ``external``, ``dir``, ``file``); a producer that
    knows nothing about reqmd leaves it empty.
    """

    key: str = ""
    title: str = ""
    body: str = ""
    items: list[Item] = field(default_factory=list)
    properties: dict[str, Any] = field(default_factory=dict)
    required: list[str] = field(default_factory=list)
    x_reqmd: dict[str, Any] = field(default_factory=dict)

    # add appends an item, nesting it under the item whose key is `parent`.
    # A flat source table (DOORS objects, CSV rows) therefore stays a flat
    # loop; the tree is assembled here rather than in every importer.
    def add(self, item: Item, parent: str = "") -> Document:
        """Append an item, under ``parent`` when given. Returns self."""
        if parent:
            target = self._find(parent)
            if target is None:
                raise KeyError(f"document {self.key!r}: no item with key {parent!r}")
            target.children.append(item)
        else:
            self.items.append(item)
        return self

    def _find(self, key: str) -> Item | None:
        stack = list(self.items)
        while stack:
            it = stack.pop()
            if it.key == key:
                return it
            stack.extend(it.children)
        return None

    def to_dict(self) -> dict[str, Any]:
        data = {
            "key": self.key,
            "title": self.title,
            "body": self.body,
            "items": [i.to_dict() for i in self.items],
            "properties": self.properties,
            "required": self.required,
            "x-reqmd": self.x_reqmd,
        }
        return _clean(data)


@dataclass
class Bundle:
    """A list of documents: the root of the exchange format."""

    documents: list[Document] = field(default_factory=list)

    def add(self, document: Document) -> Bundle:
        """Append a document. Returns self so calls can be chained."""
        self.documents.append(document)
        return self

    def to_dict(self) -> dict[str, Any]:
        return {
            "$schema": SCHEMA_URL,
            "documents": [d.to_dict() for d in self.documents],
        }

    def to_json(self, indent: int = 2) -> str:
        return json.dumps(self.to_dict(), indent=indent, ensure_ascii=False) + "\n"

    def write(self, path: str = "-") -> None:
        """Write the bundle to ``path``; ``"-"`` (the default) writes to stdout."""
        if path in ("-", ""):
            sys.stdout.write(self.to_json())
            return
        try:
            with open(path, "w", encoding="utf-8") as fh:
                fh.write(self.to_json())
        except OSError as exc:
            raise OSError(f"cannot write bundle to {path}: {exc}") from exc
