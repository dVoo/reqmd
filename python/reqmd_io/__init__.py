"""reqmd_io — build `reqmd-import/v1` exchange bundles from Python.

A thin, dependency-free helper for writing reqmd importers. It only produces
the JSON exchange format; reqmd-import is the single renderer that turns that
into reqmd documents (schema.yaml + .md). This library deliberately contains
no Markdown or YAML logic, so it cannot drift from the Go renderer.

Typical importer::

    from reqmd_io import Bundle, Document, Requirement

    doc = Document(dir="brakes", id_prefix="DOORS-", level="stakeholder-requirements",
                   title="Brakes Module")
    for item in doors_module.objects:
        doc.add(Requirement(native_id=item.number, title=item.heading, body=item.text,
                            parent=item.parent_number,
                            native_trace=[l.target for l in item.out_links],
                            attrs={"priority": item.priority}))

    Bundle([doc]).write("doors-brakes.json")

Then::

    reqmd-import doors-brakes.json -o spec/00-doors/
"""

from __future__ import annotations

import json
import sys
from dataclasses import dataclass, field
from typing import Any

FORMAT = "github.com/dVoo/reqmd/reqmd-import/v1"

__all__ = ["FORMAT", "Bundle", "Document", "Requirement"]


def _clean(data: dict[str, Any]) -> dict[str, Any]:
    """Drop empty values so the emitted bundle stays small (mirrors the Go
    side's `omitempty`)."""
    return {k: v for k, v in data.items() if v not in ("", None, [], {})}


@dataclass
class Requirement:
    """One imported requirement, keyed by the source tool's own stable ID.

    ``native_id`` is the only required field: the rendered reqmd ID is the
    owning document's ``id_prefix`` + ``native_id``. ``trace`` holds
    fully-qualified reqmd IDs passed through verbatim; ``native_trace`` holds
    native ids resolved within the bundle (use it for links to sibling items).
    """

    native_id: str
    title: str = ""
    body: str = ""
    status: str = ""
    parent: str = ""
    trace: list[str] = field(default_factory=list)
    native_trace: list[str] = field(default_factory=list)
    attrs: dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> dict[str, Any]:
        return _clean(
            {
                "native-id": self.native_id,
                "title": self.title,
                "body": self.body,
                "status": self.status,
                "parent": self.parent,
                "trace": self.trace,
                "native-trace": self.native_trace,
                "attrs": self.attrs,
            }
        )


@dataclass
class Document:
    """One reqmd document directory: schema.yaml plus, when it has
    requirements, one .md file."""

    id_prefix: str = ""
    dir: str = ""
    file: str = ""
    title: str = ""
    notice: str = ""
    level: str = ""
    document_id: str = ""
    upstream: dict[str, Any] | None = None
    requires_trace_from: list[str] = field(default_factory=list)
    external: bool = False
    schema_id: str = ""
    schema_title: str = ""
    schema_description: str = ""
    properties: dict[str, Any] = field(default_factory=dict)
    required: list[str] = field(default_factory=list)
    requirements: list[Requirement] = field(default_factory=list)

    def add(self, requirement: Requirement) -> Document:
        """Append a requirement. Returns self so calls can be chained."""
        self.requirements.append(requirement)
        return self

    def to_dict(self) -> dict[str, Any]:
        data: dict[str, Any] = {
            "dir": self.dir,
            "file": self.file,
            "title": self.title,
            "notice": self.notice,
            "id-prefix": self.id_prefix,
            "level": self.level,
            "document-id": self.document_id,
            "upstream": self.upstream,
            "requires-trace-from": self.requires_trace_from,
            "external": self.external,
            "schema-id": self.schema_id,
            "schema-title": self.schema_title,
            "schema-description": self.schema_description,
            "properties": self.properties,
            "required": self.required,
            "requirements": [r.to_dict() for r in self.requirements],
        }
        # external=False is meaningful-intent-free when absent, so drop it;
        # every other empty value is dropped by _clean.
        if not self.external:
            data.pop("external", None)
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
        return {"format": FORMAT, "documents": [d.to_dict() for d in self.documents]}

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
