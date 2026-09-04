workspace "ReqMD" "Specification authoring, validation, export, and source-code requirement traceability" {

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

            reqmdCli = container "reqmd CLI" "Go binary: check, ls, stats, export (CSV/HTML/graph), serve (live-reload), baseline diff, repin (version-pin updates), init — all with --json and --filter output" "Go 1.26" "CLI" {
                parser       = component "Markdown Parser"  "Discovers schema.yaml per dir and parses .md via goldmark AST with GFM and parallel worker pool; assembles the full content tree (every heading is a node except the level-1 document title: requirement / container / info; h1 is the title and only h2+ can be requirements) with nothing dropped and deterministic file-order merging" "Go / goldmark"
                validator    = component "Schema Validator" "Injects built-in attrs (including the status lifecycle and additional-status-values), validates attr maps against JSON Schema 2020-12, and accepts reqmd-suppress per requirement" "Go / google/jsonschema-go"
                graphBuilder = component "Graph Builder"    "Builds in-memory adjacency from parsed requirements, resolves doc-id-qualified traces, indexes x-reqmd.level for level-typed requires-trace-from coverage, and resolves each requirement's effective requires-trace-from (own attribute or inherited x-reqmd.requires-trace-from document default)" "Go"
                traceChecker = component "Trace Checker"   "Runs Pass 2 trace checks, Pass 3 sub-req parent validation, disjoint-attribute checks, and outcome-gated verdict checks against the in-memory cache; honors per-requirement reqmd-suppress" "Go"
                filter       = component "Attribute Filter" "Compiles --filter expressions (expr-lang) once, validates attribute names against all schemas, and scopes check/ls/stats/export/serve/baseline to matching requirements with filter-aware coverage" "Go / expr-lang"
                exporter     = component "Exporter"        "Renders standalone HTML (cards, collapsible container sections, info blocks, sidebar TOC with folders), CSV (Type column, document order), and LadybugDB graph outputs" "Go"
                reporter     = component "Reporter"        "Aggregates Pass 1/2/3 results, emits formatted or JSON output, sets exit code" "Go"
                differ       = component "Baseline Differ" "Compares requirements and schemas between two git tags via git archive, using r3labs/diff for attribute-level changes" "Go / r3labs-diff"
                verifyLoader = component "Verify Loader"   "Loads ephemeral verification results (CTRF JSON + manual markdown), binds entries via x-reqmd id/case/verifies, synthesizes RESULT: pseudo-requirements and TC:<case> test-case nodes (marked Synthetic), feeds them to the graph for rolled-up outcome-gated checks" "Go / stdlib"
                repinner     = component "Repinner"        "Computes version-pin deltas from the graph and rewrites ```attr blocks in place to update ~N pins to the upstream's current version" "Go / stdlib"
            }
        }

        // reqmd-import: source-code requirement trace extraction tool
        extractionTool = softwareSystem "Extraction Tool" "Parses source code with Tree-sitter (C, C++, Go, Python, Rust grammars), extracts symbols and requirement IDs, and writes ephemeral .md requirement files (proxy items) into a target spec directory" {

            scanner = container "Repository Scanner" "Discovers files, detects languages, computes hashes, and schedules parsing work" "CLI / File walker"

            parserRuntime = container "Parsing Runtime" "Loads Tree-sitter language grammars, parses files, and executes language-specific queries" "Tree-sitter runtime" {
                languageRegistry = component "Language Registry" "Selects the correct Tree-sitter grammar and query set for each file type" "Registry component"
                syntaxParser     = component "Syntax Parser"     "Builds syntax trees from source files using Tree-sitter parsers" "Tree-sitter parser"
                queryExecutor    = component "Query Executor"    "Runs queries that capture definitions, names, spans, and documentation nodes" "Tree-sitter query engine"
            }

            normalizer = container "Normalization & Trace Engine" "Maps raw captures into a common symbol model, binds adjacent docs, extracts requirement IDs, and resolves trace links" "Application service" {
                symbolNormalizer = component "Symbol Normalizer"        "Converts language-specific captures into a common symbol schema" "Transformation component"
                docBinder        = component "Documentation Binder"     "Associates adjacent comments or docstrings with symbols and cleans comment markers" "Binding component"
                reqExtractor     = component "Requirement ID Extractor" "Extracts IDs such as REQ-1 from documentation text using configured rules" "Rule engine"
                traceResolver    = component "Trace Resolver"           "Builds symbol-to-requirement links and de-duplicates trace edges" "Graph builder"
            }

            cache = container "Extraction Cache" "Stores file hashes, parse metadata, and extracted symbols to support incremental runs" "SQLite or embedded KV" {
                tags "Database"
            }

            proxyWriter = container "Proxy .md Writer" "Renders normalized symbols into ephemeral .md requirement files and writes them into a target spec directory" "CLI / File writer"
        }

        // External systems
        vscodeExt        = softwareSystem "VS Code / Editor"         "Author edits *.md and schema.yaml" "External"
        supplier         = softwareSystem "Tier-1 Supplier Tool"     "(Planned) Receives CSV export" "External"
        sourceRepository = softwareSystem "Source Repository"        "Git repository containing C, C++, Go, Python, Rust, Zig, and other source files with requirement references in comments or docstrings" "External"
        ci               = softwareSystem "CI Pipeline"              "Automated pipeline that runs extraction on changes and publishes trace artifacts" "External"
        testArtifacts    = softwareSystem "Test Result Artifacts"     "Ephemeral CTRF JSON reports and manual review/inspection markdown produced by CI runs and reviewers; loaded per check --results invocation, never persisted in the spec repo" "External"

        // People → systems
        author    -> vscodeExt          "Authors specs in"
        reviewer  -> specRepo           "Reviews diffs and comments on PRs"
        viewer    -> specRepo           "Reads rendered requirements from"
        developer -> extractionTool     "Configures and runs"
        developer -> sourceRepository   "Commits source code to"

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
        traceChecker -> graphBuilder "Runs Pass 2 checks against CachedNode cache"
        reporter     -> validator    "Receives Pass 1 errors from"
        reporter     -> traceChecker "Receives Pass 2 warnings and errors from"
        exporter     -> graphBuilder "Queries upstream/downstream neighbours from graph"
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
        extractionTool.scanner        -> extractionTool.cache          "Reads/writes file hashes and work state"
        extractionTool.scanner        -> extractionTool.parserRuntime  "Submits files for parsing"
        extractionTool.parserRuntime  -> extractionTool.cache          "Reads cached parse metadata from"
        extractionTool.parserRuntime  -> extractionTool.normalizer     "Sends raw symbol and doc captures to"
        extractionTool.normalizer     -> extractionTool.cache          "Stores normalized symbols and trace links in"
        extractionTool.normalizer     -> extractionTool.proxyWriter   "Provides normalized trace model to"
        extractionTool.proxyWriter    -> extractionTool.cache          "Reads incremental results from"

        // --- Extraction tool component internals ---
        extractionTool.parserRuntime.languageRegistry  -> extractionTool.parserRuntime.syntaxParser     "Provides parser configuration to"
        extractionTool.parserRuntime.languageRegistry  -> extractionTool.parserRuntime.queryExecutor    "Provides language queries to"
        extractionTool.parserRuntime.syntaxParser      -> extractionTool.parserRuntime.queryExecutor    "Provides syntax trees to"
        extractionTool.parserRuntime.queryExecutor     -> extractionTool.normalizer.symbolNormalizer   "Sends captured definitions to"
        extractionTool.parserRuntime.queryExecutor     -> extractionTool.normalizer.docBinder          "Sends captured comments/docstrings to"
        extractionTool.normalizer.symbolNormalizer     -> extractionTool.normalizer.docBinder          "Provides symbol identities to"
        extractionTool.normalizer.docBinder            -> extractionTool.normalizer.reqExtractor       "Provides cleaned documentation to"
        extractionTool.normalizer.reqExtractor         -> extractionTool.normalizer.traceResolver      "Provides extracted requirement IDs to"
        extractionTool.normalizer.symbolNormalizer     -> extractionTool.normalizer.traceResolver      "Provides normalized symbols to"
        extractionTool.normalizer.traceResolver        -> extractionTool.proxyWriter                  "Provides resolved symbol-to-requirement links to"

        // --- CLI → repo ---
        reqmdCli -> specRepo "Reads *.md and schema.yaml from"

        // --- Cross-system ---
        extractionTool                  -> sourceRepository   "Reads source files from"
        extractionTool                  -> reqmd.specRepo     "Reads requirement ID schemas from"
        extractionTool.proxyWriter      -> reqmd.specRepo               "Writes ephemeral .md requirement files to"
        extractionTool.proxyWriter      -> reqmd.specRepo.schemaFile    "Reads x-reqmd.id-prefix and x-reqmd.upstream from"

        // --- CI runs both tools ---
        ci -> reqmd.reqmdCli "Runs check/ls/stats/export/serve/baseline on"

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

        component extractionTool.parserRuntime "ParsingComponents" "Parsing runtime internals" {
            include *
            autoLayout
        }

        component extractionTool.normalizer "TraceComponents" "Trace engine internals" {
            include *
            autoLayout
        }

        // proxyWriter is a single-container box with no sub-components; it appears in ExtractionContainers

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
