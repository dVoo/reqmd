# Package: reqmd\_io

<!-- reqmd: generated -->

## python/reqmd\_io/\_\_init\_\_.py
### IMP-reqmd\_io-SCHEMA\_URL-dfaeec4: SCHEMA\_URL
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 41
x-reqmd.symbol-kind: const
trace: []
```

### IMP-reqmd\_io-KIND\_REQUIREMENT-34bc767: KIND\_REQUIREMENT
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 43
x-reqmd.symbol-kind: const
trace: []
```

### IMP-reqmd\_io-KIND\_SECTION-0c8e72b: KIND\_SECTION
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 44
x-reqmd.symbol-kind: const
trace: []
```

### IMP-reqmd\_io-\_\_all\_\_-dceea74: \_\_all\_\_
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 46
x-reqmd.symbol-kind: const
trace: []
```

### IMP-reqmd\_io-\_clean-0a9e533: \_clean
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 58
x-reqmd.symbol-kind: function
trace: []
```

Drop empty values so the emitted bundle stays small (mirrors the Go
side's `omitempty`).

### IMP-reqmd\_io-Item-a926aae: Item
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 65
x-reqmd.symbol-kind: type
trace: []
```

Base class for one content item.

``kind`` is the only difference between content and structure: a
requirement renders a reqmd block, a section renders structure only. Items
nest through ``children``; the flat ``parent`` argument of
:meth:`Document.add` is resolved into that tree, so a producer that holds a
flat table never has to build it by hand.

#### IMP-reqmd\_io-to\_dict-ddc0fb6: to\_dict
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 85
x-reqmd.symbol-kind: method
trace: []
```

### IMP-reqmd\_io-Requirement-a66ec10: Requirement
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 101
x-reqmd.symbol-kind: type
trace: []
```

One imported requirement, identified by the source tool's own stable id.

``key`` is required: the rendered reqmd ID is the owning document's
``id-prefix`` + ``key``. ``links`` name other items by key (``"17"`` in the
same document, ``"brakes/17"`` across documents). ``refs`` hold external
requirement IDs and are passed through verbatim.

### IMP-reqmd\_io-Section-ec9f374: Section
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 114
x-reqmd.symbol-kind: type
trace: []
```

Structure: a heading with prose and children, never a reqmd block.

A section carries no ``fields`` and no ``status``; both are rejected by the
renderer. An untitled section renders as a plain prose block.

### IMP-reqmd\_io-Document-7a8454c: Document
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 125
x-reqmd.symbol-kind: type
trace: []
```

One reqmd document: a schema plus, when it has items, one Markdown file.

``key`` and ``title`` are required. Everything reqmd-specific lives in
``x_reqmd`` (``id-prefix``, ``level``, ``document-id``, ``upstream``,
``requires-trace-from``, ``external``, ``dir``, ``file``); a producer that
knows nothing about reqmd leaves it empty.

#### IMP-reqmd\_io-add-f7e3675: add
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 145
x-reqmd.symbol-kind: method
trace: []
```

add appends an item, nesting it under the item whose key is `parent`.
A flat source table (DOORS objects, CSV rows) therefore stays a flat
loop; the tree is assembled here rather than in every importer.

Append an item, under ``parent`` when given. Returns self.

#### IMP-reqmd\_io-\_find-1767f78: \_find
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 156
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-reqmd\_io-to\_dict-3171cbc: to\_dict
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 165
x-reqmd.symbol-kind: method
trace: []
```

### IMP-reqmd\_io-Bundle-309242d: Bundle
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 179
x-reqmd.symbol-kind: type
trace: []
```

A list of documents: the root of the exchange format.

#### IMP-reqmd\_io-add-64bce80: add
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 184
x-reqmd.symbol-kind: method
trace: []
```

Append a document. Returns self so calls can be chained.

#### IMP-reqmd\_io-to\_dict-bfbb389: to\_dict
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 189
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-reqmd\_io-to\_json-960ba93: to\_json
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 195
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-reqmd\_io-write-c799b20: write
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: python/reqmd_io/__init__.py
x-reqmd.source-line: 198
x-reqmd.symbol-kind: method
trace: []
```

Write the bundle to ``path``; ``"-"`` (the default) writes to stdout.

