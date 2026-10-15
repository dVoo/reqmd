"""Tests for the reqmd_io bundle builder."""

from __future__ import annotations

import json
import os
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from reqmd_io import FORMAT, Bundle, Document, Requirement  # noqa: E402


class RequirementTest(unittest.TestCase):
    def test_omits_empty_fields(self):
        self.assertEqual(Requirement(native_id="17").to_dict(), {"native-id": "17"})

    def test_keeps_numbers_and_false_attrs(self):
        req = Requirement(native_id="17", attrs={"line": 12, "imported": False})
        self.assertEqual(req.to_dict()["attrs"], {"line": 12, "imported": False})

    def test_maps_to_kebab_case(self):
        req = Requirement(native_id="42", parent="17", native_trace=["17"], trace=["SYS-001"])
        got = req.to_dict()
        self.assertEqual(got["native-trace"], ["17"])
        self.assertEqual(got["trace"], ["SYS-001"])


class DocumentTest(unittest.TestCase):
    def test_maps_snake_case_to_format_keys(self):
        doc = Document(id_prefix="DOORS-", dir="brakes", requires_trace_from=["system"])
        got = doc.to_dict()
        self.assertEqual(got["id-prefix"], "DOORS-")
        self.assertEqual(got["requires-trace-from"], ["system"])
        self.assertNotIn("external", got)

    def test_keeps_external_true(self):
        self.assertTrue(Document(id_prefix="X-", external=True).to_dict()["external"])

    def test_add_is_chainable(self):
        doc = Document(id_prefix="X-").add(Requirement(native_id="1")).add(Requirement(native_id="2"))
        self.assertEqual([r["native-id"] for r in doc.to_dict()["requirements"]], ["1", "2"])


class BundleTest(unittest.TestCase):
    def test_format_and_json(self):
        bundle = Bundle([Document(id_prefix="X-", requirements=[Requirement(native_id="1")])])
        payload = json.loads(bundle.to_json())
        self.assertEqual(payload["format"], FORMAT)
        self.assertEqual(payload["documents"][0]["requirements"][0]["native-id"], "1")
        self.assertTrue(bundle.to_json().endswith("\n"))

    def test_write_to_file(self):
        bundle = Bundle([Document(id_prefix="X-")])
        with tempfile.TemporaryDirectory() as tmp:
            path = os.path.join(tmp, "bundle.json")
            bundle.write(path)
            with open(path, encoding="utf-8") as fh:
                self.assertEqual(json.load(fh)["format"], FORMAT)


if __name__ == "__main__":
    unittest.main()
