workspace "ReqMD" "Specification authoring, validation, export, source-code and ALM requirement import" {

    model {

        // People
        author    = person "Requirement Author"  "Engineers writing specs in Markdown"
        viewer    = person "Viewer / Consumer"   "Engineers viewing and evaluating specs"
        reviewer  = person "Reviewer / Supplier" "Stakeholders and tier-1 suppliers reviewing specs"
        developer = person "Developer"           "Runs the extraction tool locally or in CI to generate traceability links"

        // ReqMD software system
        reqmd = softwareSystem "ReqMD Toolchain" "Text-first authoring, validation, export" {

            specRepo = container "Specification Repository" "Git repo holding *.md and schema.yaml" "Git / plain text" "Repository" {
                mdFiles    = component "*.md Spec Files"   "Requirement prose + attr blocks + YAML frontmatter per subsystem" "Markdown"
                schemaFile = component "schema.yaml"   "JSON Schema 2020-12 in YAML + x-reqmd upstream" "YAML"
            }

            reqmdCli = container "reqmd CLI" "Go binary: check, ls, stats, export (CSV/HTML/NDJSON/graph), serve (live-reload), baseline diff, repin (version-pin updates), init — all with --json and --filter output" "Go 1.26" "CLI" {
                parser       = component "Markdown Parser"  "Discovers schema.yaml per dir and parses .md via goldmark AST with GFM and parallel worker pool; assembles the full content tree (every heading is a node except the level-1 document title: requirement / container / info; h1 is the title and only h2+ can be requirements) with nothing dropped and deterministic file-order merging" "Go / goldmark"
                validator    = component "Schema Validator" "Injects built-in attrs (including the status lifecycle and additional-status-values), rejects reserved-name redefinition, validates attr maps against JSON Schema 2020-12, and accepts reqmd-suppress per requirement" "Go / google/jsonschema-go"
                graphBuilder = component "Graph Builder"    "Builds in-memory adjacency from parsed requirements, resolves doc-id-qualified traces, indexes x-reqmd.level for level-typed requires-trace-from coverage, and resolves each requirement's effective requires-trace-from (own attribute or inherited x-reqmd.requires-trace-from document default)" "Go"
                traceChecker = component "Trace Checker"   "Runs the graph check catalog against the in-memory node cache: structural integrity, boundary-aware coverage, disjoint attributes, version pins, and outcome-gated verdicts with the strict evidence roll-up; honors per-requirement reqmd-suppress" "Go"
                filter       = component "Attribute Filter" "Compiles --filter expressions (expr-lang) once, validates attribute names against all schemas, and scopes check/ls/stats/export/serve/baseline to matching requirements with filter-aware coverage" "Go / expr-lang"
                exporter     = component "Exporter"        "Renders standalone, fully offline HTML (server-rendered Document view + on-demand Table/Trace/Coverage views from one shared page model, project home index, graph-derived coverage/verification/link state) and CSV (Type column, document order)" "Go"
                jsonExporter = component "NDJSON Table Writer" "Writes requirements/traces/results/findings NDJSON tables for warehouse ingestion: built-ins promoted to columns, null-padded fixed columns, byte-deterministic output" "Go"
                graphExporter = component "Graph Exporter"  "Writes a LadybugDB graph database (requirement nodes, TracesTo edges, optional RESULT: verdict nodes) behind the ladybug build tag" "Go / ladybugdb"
                scaffolder   = component "Scaffolder"      "Renders embedded (generic, aspice, results) or custom-directory templates into a new schema.yaml plus example document" "Go / text-template"
                reporter     = component "Reporter"        "Aggregates Pass 1 and graph check results, groups result-attributed findings into a verification section, emits formatted or JSON output, sets exit code" "Go"
                differ       = component "Baseline Differ" "Compares requirements and schemas between two git tags via git archive, using r3labs/diff for attribute-level changes" "Go / r3labs-diff"
                verifyLoader = component "Verify Loader"   "Loads ephemeral verification results (CTRF JSON + manual markdown), binds entries via x-reqmd id/case/verifies, synthesizes RESULT: pseudo-requirements and TC:<case> test-case nodes (marked Synthetic), feeds them to the graph for rolled-up outcome-gated checks" "Go / stdlib"
                repinner     = component "Repinner"        "Computes version-pin deltas from the graph and rewrites ```attr blocks in place to update ~N pins to the upstream's current version" "Go / stdlib"
            }
        }

        // reqmd-code: source-code requirement trace extraction tool
        extractionTool = softwareSystem "Extraction Tool" "Parses source code with Tree-sitter (C, C++, Go, Python, Rust grammars), extracts symbols and requirement IDs, and emits a reqmd-import/v1 exchange bundle (JSON) for the Import Tool to render as proxy requirements" {

            extractCli = container "Extraction CLI" "Walks the source tree, dispatches each file to the plugin claiming its extension, parses in parallel, binds doc comments, groups symbols per package, and writes one reqmd-import/v1 bundle to stdout or -o" "Go / cobra" "CLI"

            languages = container "Language Plugins" "Self-registering Tree-sitter plugins: the Language interface plus C, C++, Go, Python, and Rust grammars, queries, and capture-to-symbol mappings" "Go / tree-sitter" {
                registry     = component "Plugin Registry"          "Resolves the plugin for a file extension; plugins self-register via blank imports, so adding a language needs no change to the extractor" "Go"
                symbolMapper = component "Capture to Symbol Mapper" "Maps language-specific captures to the shared symbol model, populating method receivers from the enclosing type and excluding function-local declarations" "Go / tree-sitter query"
            }

            traceEngine = container "Trace Engine" "Extracts explicit reqmd:trace markers from bound doc comments and, opt-in via --heuristic-traces, bare requirement IDs found in prose" "Go"

            symbolModel = container "Symbol Model" "Normalized symbol type plus the stable short hash over package and symbol that makes generated IDs line-independent" "Go"

            attrSchema = container "Standard Attr Schema" "Canonical provenance attribute set (x-reqmd.imported, source-file, source-line, symbol-kind, heuristic-trace) projected into every emitted document's schema" "YAML"
        }

        // reqmd-import: the single renderer for the exchange format
        importTool = softwareSystem "Import Tool" "Renders a reqmd-import/v1 exchange bundle into reqmd documents (schema.yaml + .md). The only component that writes Markdown or YAML, so every importer shares one set of quoting, heading, and schema rules" {

            exchangeFormat = container "Exchange Format" "reqmd-import/v1 JSON contract: documents carrying id-prefix, level, document-id, upstream and attr-schema properties, plus native-id-keyed requirements (parent, trace, native-trace, attrs). Building a bundle rejects unsupported formats, empty document lists, duplicate native IDs, unknown parent/native-trace refs, reserved attrs, escaping paths, parent cycles, and duplicate output paths" "JSON"

            renderer = container "Document Renderer" "Resolves native IDs to reqmd IDs, orders children under their parents, renders attr blocks with correct YAML quoting, emits each document's schema.yaml, and writes files idempotently behind a generated marker" "Go" "CLI" {
                idResolver   = component "ID & Reference Resolver"   "Stamps <id-prefix><native-id>, rewrites native-trace refs against a bundle-wide index, and rejects duplicate IDs, unknown refs, reserved attrs, escaping paths, and parent cycles" "Go"
                schemaWriter = component "Schema Writer"             "Emits a document's schema.yaml from its properties/required plus the x-reqmd block (level, document-id, id-prefix, upstream, external)" "Go / yaml.v3"
            }
        }

        // External systems
        vscodeExt        = softwareSystem "VS Code / Editor"         "Author edits *.md and schema.yaml" "External"
        supplier         = softwareSystem "Tier-1 Supplier Tool"     "(Planned) Receives CSV export" "External"
        sourceRepository = softwareSystem "Source Repository"        "Git repository containing C, C++, Go, Python, Rust, Zig, and other source files with requirement references in comments or docstrings" "External"
        alm              = softwareSystem "ALM System"              "DOORS / codeBeamer / Polarion: database-backed requirement stores that export CSV or expose REST APIs" "External"
        customImporter   = softwareSystem "Custom Importer"         "A Python (or any language) script using the reqmd-io helper to turn an ALM export, CSV, or API into a reqmd-import/v1 bundle" "External"
        ci               = softwareSystem "CI Pipeline"              "Automated pipeline that runs extraction on changes and publishes trace artifacts" "External"
        testArtifacts    = softwareSystem "Test Result Artifacts"     "Ephemeral CTRF JSON reports and manual review/inspection markdown produced by CI runs and reviewers; loaded per check --results invocation, never persisted in the spec repo" "External"

        // People → systems
        author    -> vscodeExt          "Authors specs in"
        reviewer  -> specRepo           "Reviews diffs and comments on PRs"
        viewer    -> specRepo           "Reads rendered requirements from"
        developer -> extractionTool     "Configures and runs"
        developer -> sourceRepository   "Commits source code to"
        developer -> customImporter     "Writes and runs"

        // Editor → repo
        vscodeExt -> specRepo "Reads and writes"

        ci -> extractionTool "Runs on push / pull request"
        ci -> testArtifacts   "Publishes CTRF reports to"
        verifyLoader -> testArtifacts "Loads CTRF JSON and manual-results markdown from"

        // --- ReqMD CLI internal flow ---
        parser       -> mdFiles      "Reads and parses *.md via goldmark AST with frontmatter"
        parser       -> schemaFile   "Reads x-reqmd upstream from"
        validator    -> schemaFile   "Loads JSON Schema from"
        validator    -> parser       "Receives parsed attr maps from"
        graphBuilder -> parser       "Receives all parsed requirements from"
        traceChecker -> graphBuilder "Runs the graph check catalog against the CachedNode cache"
        reporter     -> validator    "Receives Pass 1 errors from"
        reporter     -> traceChecker "Receives Pass 2 warnings and errors from"
        exporter     -> graphBuilder "Reads coverage state, verification state, and resolved trace links from graph"
        jsonExporter -> graphBuilder "Projects trace edges, findings, and rolled-up verdicts from"
        jsonExporter -> parser      "Reads content nodes in document order from"
        graphExporter -> graphBuilder "Writes requirement nodes and TracesTo edges from"
        scaffolder   -> specRepo    "Writes the scaffolded schema.yaml and example .md to"
        differ       -> parser       "Parses both tag snapshots via"
        differ       -> specRepo     "Extracts tags via git archive from"
        verifyLoader -> graphBuilder  "Synthesizes RESULT: pseudo-requirements and appends to doc slice before graph build"
        verifyLoader -> parser        "Reuses parser.Discover for manual-results markdown dirs"
        repinner     -> graphBuilder "Reads OutboundPins and OutboundRefs from graph to compute version-pin deltas"
        repinner     -> specRepo     "Rewrites ```attr blocks in .md files to update ~N pins"
        traceChecker -> verifyLoader  "Runs missing-verdict / failing-verdict checks against result nodes"
        filter       -> parser        "Validates filter attribute names against all parsed schemas"
        filter       -> graphBuilder  "Scopes per-requirement checks and coverage to matching requirements"
        filter       -> reporter      "Filters results to matching requirements"
        filter       -> exporter      "Scopes CSV/HTML export to matching requirements"

        // --- Extraction tool internal flow ---
        extractionTool.extractCli -> extractionTool.languages   "Asks for the plugin owning each file extension"
        extractionTool.extractCli -> extractionTool.traceEngine "Sends bound doc comments for requirement-ID extraction"
        extractionTool.extractCli -> extractionTool.symbolModel "Builds native IDs from"
        extractionTool.extractCli -> extractionTool.attrSchema  "Projects each document's property set from"

        // --- Extraction tool component internals ---
        extractionTool.languages.registry -> extractionTool.languages.symbolMapper "Selects the mapper matching the file's grammar"

        // --- CLI → repo ---
        reqmdCli -> specRepo "Reads *.md and schema.yaml from"

        // --- Cross-system ---
        extractionTool                  -> sourceRepository   "Reads source files from"

        // --- Import tool: one renderer for every producer ---
        extractionTool.extractCli       -> importTool          "Emits a reqmd-import/v1 bundle to"
        customImporter                  -> alm                 "Pulls requirements from"
        customImporter                  -> importTool          "Emits a reqmd-import/v1 bundle to"
        importTool.renderer             -> reqmd.specRepo      "Writes generated .md and schema.yaml to"
        importTool.renderer.idResolver  -> importTool.exchangeFormat "Reads documents and requirements from"
        importTool.renderer.schemaWriter -> importTool.renderer  "Defers document writes to"

        // --- CI runs both tools ---
        ci -> reqmd.reqmdCli "Runs check/ls/stats/export/serve/baseline on"
        ci -> importTool     "Runs extraction and import on push / pull request"

        // --- Exports ---
        exporter     -> supplier         "Exports CSV to" "Planned"
    }

    views {

        // ReqMD views
        systemContext reqmd "SystemContext" "ReqMD system context" {
            include *
            autoLayout
        }

        container reqmd "Containers" "ReqMD containers" {
            include *
            autoLayout
        }

        component reqmdCli "CLI_Components" "reqmd CLI internals" {
            include *
            autoLayout
        }

        component specRepo "Repo_Components" "Specification repository contents" {
            include *
            autoLayout
        }

        // Extraction tool views
        systemContext extractionTool "ExtractionSystemContext" "Extraction tool system context" {
            include *
            autoLayout
        }

        container extractionTool "ExtractionContainers" "Extraction tool containers" {
            include *
            autoLayout
        }

        component extractionTool.languages "LanguagePluginComponents" "Language plugin registry and mappers" {
            include *
            autoLayout
        }

        // Import tool views
        systemContext importTool "ImportSystemContext" "Import tool system context" {
            include *
            autoLayout
        }

        container importTool "ImportContainers" "Import tool containers" {
            include *
            autoLayout
        }

        component importTool.renderer "ImportRendererComponents" "Document renderer internals" {
            include *
            autoLayout
        }

        // Single-container boxes (extractCli, traceEngine, symbolModel, attrSchema) appear in ExtractionContainers

        styles {
            element "Person" {
                shape Person
                background #08427b
                color #ffffff
            }
            element "Software System" {
                background #1168bd
                color #ffffff
            }
            element "External" {
                background #999999
                color #ffffff
            }
            element "Container" {
                background #438dd5
                color #ffffff
            }
            element "Component" {
                background #85bbf0
                color #000000
            }
            element "Repository" {
                shape Cylinder
                background #438dd5
                color #ffffff
            }
            element "Database" {
                shape Cylinder
                background #2e7d32
                color #ffffff
            }
            element "CLI" {
                shape RoundedBox
            }
            element "Relationship" {
                color #707070
            }
            element "Planned" {
                opacity 40
                stroke dashed
            }
        }
    }
}
