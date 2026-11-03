"""Tests for the reqmd_io exchange-bundle builder."""

from __future__ import annotations

import io
import json
import sys
import tempfile
import unittest
from contextlib import redirect_stdout
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from reqmd_io import (  # noqa: E402
    KIND_REQUIREMENT,
    KIND_SECTION,
    SCHEMA_URL,
    Bundle,
    Document,
    Item,
    Requirement,
    Section,
)

SCHEMA_PATH = (
    Path(__file__).resolve().parents[2] / "site" / "static" / "exchange" / "v1" / "schema.json"
)


class ItemTest(unittest.TestCase):
    def test_omits_empty_fields(self):
        self.assertEqual(Requirement(key="17").to_dict(), {"kind": "requirement", "key": "17"})

    def test_kind_distinguishes_content_from_structure(self):
        self.assertEqual(Requirement(key="1").to_dict()["kind"], KIND_REQUIREMENT)
        self.assertEqual(Section(title="S").to_dict()["kind"], KIND_SECTION)
        # The base class defaults to a requirement, matching the Go zero value.
        self.assertEqual(Item(key="1").to_dict()["kind"], KIND_REQUIREMENT)

    def test_keeps_numbers_and_false_fields(self):
        req = Requirement(key="17", fields={"line": 12, "imported": False})
        self.assertEqual(req.to_dict()["fields"], {"line": 12, "imported": False})

    def test_maps_to_kebab_case(self):
        req = Requirement(key="42", links=["17"], refs=["SYS-001"])
        got = req.to_dict()
        self.assertEqual(got["links"], ["17"])
        self.assertEqual(got["refs"], ["SYS-001"])

    def test_nested_children_render(self):
        parent = Requirement(key="1", title="Parent")
        parent.children.append(Requirement(key="2", title="Child"))
        got = parent.to_dict()
        self.assertEqual(got["children"][0]["key"], "2")


class DocumentTest(unittest.TestCase):
    def test_requires_key_and_title(self):
        doc = Document(key="brakes", title="Brakes Module")
        got = doc.to_dict()
        self.assertEqual(got["key"], "brakes")
        self.assertEqual(got["title"], "Brakes Module")

    def test_x_reqmd_is_namespaced(self):
        doc = Document(
            key="brakes",
            title="Brakes",
            x_reqmd={"id-prefix": "DOORS-", "level": "stakeholder-requirements"},
        )
        got = doc.to_dict()
        self.assertEqual(got["x-reqmd"]["id-prefix"], "DOORS-")
        self.assertNotIn("id-prefix", got)

    def test_add_is_chainable(self):
        doc = Document(key="d", title="D")
        self.assertIs(doc.add(Requirement(key="1")), doc)
        self.assertEqual(doc.to_dict()["items"][0]["key"], "1")

    def test_add_nests_under_parent_key(self):
        doc = Document(key="d", title="D")
        doc.add(Section(title="Hydraulics"))
        doc.add(Requirement(key="1", title="Brake request"))
        doc.add(Requirement(key="2", title="Pressure"), parent="1")
        items = doc.to_dict()["items"]
        self.assertEqual(len(items), 2)
        self.assertEqual(items[0]["kind"], KIND_SECTION)
        self.assertEqual(items[1]["children"][0]["key"], "2")

    def test_add_unknown_parent_raises(self):
        doc = Document(key="d", title="D")
        with self.assertRaises(KeyError):
            doc.add(Requirement(key="2"), parent="missing")

    def test_add_nests_below_the_top_level(self):
        """Parent lookup walks the whole tree, not just the top level."""
        doc = Document(key="d", title="D")
        doc.add(Section(key="chapter", title="Chapter"))
        doc.add(Section(key="sub", title="Sub"), parent="chapter")
        doc.add(Requirement(key="9"), parent="sub")
        chapter = doc.items[0]
        self.assertEqual(chapter.children[0].children[0].key, "9")


class BundleTest(unittest.TestCase):
    def test_declares_the_schema_url(self):
        got = Bundle([Document(key="d", title="D")]).to_dict()
        self.assertEqual(got["$schema"], SCHEMA_URL)

    def test_to_json_is_valid_json(self):
        bundle = Bundle([Document(key="d", title="D").add(Requirement(key="1", title="R"))])
        self.assertEqual(json.loads(bundle.to_json())["documents"][0]["key"], "d")

    def test_write_to_file(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "bundle.json"
            Bundle([Document(key="d", title="D")]).write(str(path))
            self.assertEqual(json.loads(path.read_text())["$schema"], SCHEMA_URL)

    def test_write_to_stdout(self):
        buf = io.StringIO()
        with redirect_stdout(buf):
            Bundle([Document(key="d", title="D")]).write()
        self.assertEqual(json.loads(buf.getvalue())["documents"][0]["key"], "d")


class SchemaConformanceTest(unittest.TestCase):
    """The neutrality proof: a Python producer validates against the published
    schema without any Go tooling. Skipped when `jsonschema` is not installed,
    the same way the client JS tests skip without node."""

    def setUp(self):
        try:
            import jsonschema  # noqa: F401
        except ImportError:  # pragma: no cover - environment dependent
            self.skipTest("jsonschema not installed")

    def test_generated_bundle_matches_schema(self):
        import jsonschema

        schema = json.loads(SCHEMA_PATH.read_text())
        doc = Document(
            key="brakes",
            title="Brakes Module",
            body="Module intro.",
            properties={"priority": {"type": "string"}},
            required=["priority"],
            x_reqmd={"id-prefix": "DOORS-", "level": "stakeholder-requirements", "dir": "brakes"},
        )
        doc.add(Section(title="Hydraulics", body="The hydraulic circuit."))
        doc.add(Requirement(key="17", title="Brake request", fields={"priority": "High"}))
        doc.add(Requirement(key="42", title="Pressure", status="Approved", links=["17"]), parent="17")
        doc.add(Requirement(key="43", title="External", refs=["SYS-BRAKE-001"]))
        bundle = Bundle([doc]).to_dict()

        jsonschema.validate(bundle, schema)

        # The schema rejects a section that carries fields: sections are
        # structure and never own attributes.
        broken = Bundle(
            [Document(key="d", title="D", items=[Section(title="S", fields={"a": 1})])]
        ).to_dict()
        with self.assertRaises(jsonschema.ValidationError):
            jsonschema.validate(broken, schema)


if __name__ == "__main__":
    unittest.main()
