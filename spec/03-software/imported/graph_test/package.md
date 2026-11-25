# Package: graph\_test

<!-- reqmd: generated -->

## internal/graph/analysis\_test.go
### IMP-graph\_test-analysisDoc-e526b2a: analysisDoc
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/analysis_test.go
x-reqmd.source-line: 12
x-reqmd.symbol-kind: function
trace: []
```

analysisDoc builds a document with a document-id so qualified trace refs
and coverage sources resolve.

### IMP-graph\_test-attrs-c0facb2: attrs
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/analysis_test.go
x-reqmd.source-line: 16
x-reqmd.symbol-kind: function
trace: []
```

### IMP-graph\_test-TestCoverageStates\_Semantic-82af143: TestCoverageStates\_Semantic
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/analysis_test.go
x-reqmd.source-line: 27
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-CoverageStates-6accec7
```

TestCoverageStates_Semantic pins the coverage states to the same rules the
requires-trace-from check applies: approved providers only, draft inbound
requirements ignored, explicit opt-out, unknown sources unresolved.

### IMP-graph\_test-TestCoverageStates\_GenericFallback-303ba4a: TestCoverageStates\_GenericFallback
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/analysis_test.go
x-reqmd.source-line: 111
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-CoverageStates-6accec7
```

TestCoverageStates_GenericFallback checks the boundary-aware fallback used
when a requirement declares no requires-trace-from expectation.

### IMP-graph\_test-TestTraceRefs-cce5c9d: TestTraceRefs
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/analysis_test.go
x-reqmd.source-line: 156
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-TraceRefs-8ad701a
```

TestTraceRefs covers pin/version annotation and unresolved links.

### IMP-graph\_test-TestVerificationStates-bda6743: TestVerificationStates
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/analysis_test.go
x-reqmd.source-line: 187
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-VerificationStates-edc734e
```

TestVerificationStates checks that eligibility follows the missing-verdict
gate and that unverified eligible measures are visible as empty outcomes.

## internal/graph/filter\_test.go
### IMP-graph\_test-TestFilterAwareCoverage\_FilteredOutInboundNoFalseFailure-f72457f: TestFilterAwareCoverage\_FilteredOutInboundNoFalseFailure
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/filter_test.go
x-reqmd.source-line: 20
x-reqmd.symbol-kind: function
trace: []
```

Filter-aware coverage checking (RFC §3.2)
TestFilterAwareCoverage_FilteredOutInboundNoFalseFailure verifies that
a requirement excluded by the filter does not cause a false coverage
failure for a filtered-in requirement. This is the core behavioral
guarantee of RFC §3.2: "A requirement excluded by the filter cannot be
required as, or count as, a coverage provider for another excluded
requirement."

### IMP-graph\_test-TestFilterAwareCoverage\_FilteredInProviderSatisfies-5cc3985: TestFilterAwareCoverage\_FilteredInProviderSatisfies
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/filter_test.go
x-reqmd.source-line: 97
x-reqmd.symbol-kind: function
trace: []
```

TestFilterAwareCoverage_FilteredInProviderSatisfies verifies that a
filtered-in inbound requirement properly satisfies coverage.

### IMP-graph\_test-TestFilterAwareCoverage\_DocumentDefault-8fe20cd: TestFilterAwareCoverage\_DocumentDefault
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/filter_test.go
x-reqmd.source-line: 140
x-reqmd.symbol-kind: function
trace: []
```

TestFilterAwareCoverage_DocumentDefault verifies filter-aware coverage
interacts with the x-reqmd.requires-trace-from document default: an
inherited expectation behaves like a per-requirement one under --filter.

### IMP-graph\_test-TestFilter\_NoFilter\_ParityWithUnfiltered-a6c1b74: TestFilter\_NoFilter\_ParityWithUnfiltered
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/filter_test.go
x-reqmd.source-line: 189
x-reqmd.symbol-kind: function
trace: []
```

TestFilter_NoFilter_ParityWithUnfiltered verifies that an empty filter
set produces identical results to no filter at all.

### IMP-graph\_test-TestDisjointAttribute\_ZeroIntersection\_Errors-0665bca: TestDisjointAttribute\_ZeroIntersection\_Errors
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/filter_test.go
x-reqmd.source-line: 238
x-reqmd.symbol-kind: function
trace: []
```

Disjoint-attribute check (RFC §3.5)
TestDisjointAttribute_ZeroIntersection_Errors verifies that a trace link
between two requirements with non-overlapping variant arrays produces an
ERROR.

### IMP-graph\_test-TestDisjointAttribute\_PartialOverlap\_NoError-c19a427: TestDisjointAttribute\_PartialOverlap\_NoError
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/filter_test.go
x-reqmd.source-line: 276
x-reqmd.symbol-kind: function
trace: []
```

TestDisjointAttribute_PartialOverlap_NoError verifies that overlapping
values do not trigger an error.

### IMP-graph\_test-TestDisjointAttribute\_EmptyAttribute\_Exempt-7d2dc7e: TestDisjointAttribute\_EmptyAttribute\_Exempt
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/filter_test.go
x-reqmd.source-line: 308
x-reqmd.symbol-kind: function
trace: []
```

TestDisjointAttribute_EmptyAttribute_Exempt verifies that a requirement
with an empty/absent variant attribute is exempt (treated as "applies to all").

### IMP-graph\_test-TestDisjointAttribute\_NoAttrsSet\_NoCheck-59286f6: TestDisjointAttribute\_NoAttrsSet\_NoCheck
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/filter_test.go
x-reqmd.source-line: 340
x-reqmd.symbol-kind: function
trace: []
```

TestDisjointAttribute_NoAttrsSet_NoCheck verifies that without
SetDisjointAttrs, the check is a no-op.

### IMP-graph\_test-filterByMessage-f8326fc: filterByMessage
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/filter_test.go
x-reqmd.source-line: 373
x-reqmd.symbol-kind: function
trace: []
```

Test helpers

### IMP-graph\_test-filterByCode-940e9bc: filterByCode
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/filter_test.go
x-reqmd.source-line: 383
x-reqmd.symbol-kind: function
trace: []
```

### IMP-graph\_test-contains-cf5eec3: contains
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/filter_test.go
x-reqmd.source-line: 393
x-reqmd.symbol-kind: function
trace: []
```

### IMP-graph\_test-containsImpl-3e592f5: containsImpl
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/filter_test.go
x-reqmd.source-line: 397
x-reqmd.symbol-kind: function
trace: []
```

## internal/graph/graph\_test.go
### IMP-graph\_test-req-267e4ee: req
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 18
x-reqmd.symbol-kind: function
trace: []
```

Helper factories

### IMP-graph\_test-doc-89b07e8: doc
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 27
x-reqmd.symbol-kind: function
trace: []
```

### IMP-graph\_test-docWithXReqmd-56289a9: docWithXReqmd
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 40
x-reqmd.symbol-kind: function
trace: []
```

docWithXReqmd creates a document with a custom XReqmd configuration.

### IMP-graph\_test-TestNew\_EmptyDocs-3951219: TestNew\_EmptyDocs
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 56
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

1. Empty docs

### IMP-graph\_test-TestNew\_SingleDocNoTraces-25a3ee8: TestNew\_SingleDocNoTraces
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 74
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

Helpers

### IMP-graph\_test-TestNew\_ValidTrace-4ec3872: TestNew\_ValidTrace
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 109
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

3. Valid trace — two requirements, one traces to the other

### IMP-graph\_test-TestNew\_DanglingTrace-cc5846e: TestNew\_DanglingTrace
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 183
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

4. Dangling trace — one req traces to a non-existent ID

### IMP-graph\_test-TestNew\_CircularTrace-b4c4fb9: TestNew\_CircularTrace
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 230
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

5. Circular trace — A→B→C→A

### IMP-graph\_test-TestNew\_OrphanDetection-ca28cfa: TestNew\_OrphanDetection
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 268
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

6. Orphan detection — A→B means A has no incoming → orphan

### IMP-graph\_test-TestNew\_LeafBoundarySuppressesUntraced-eeaf8a0: TestNew\_LeafBoundarySuppressesUntraced
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 303
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

7. Orphan suppression via directory boundary (no upstream sources)
TestNew_LeafBoundarySuppressesUntraced closes the gap left when the leaf rule
was added to the check: a requirement at the bottom of the V has no possible
downstream tracer, so generic untraced must stay silent there.

### IMP-graph\_test-TestNew\_RootBoundarySuppressesOrphan-fa26002: TestNew\_RootBoundarySuppressesOrphan
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 331
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

### IMP-graph\_test-TestNew\_NonBoundaryOrphanAndMissing-92dfcdf: TestNew\_NonBoundaryOrphanAndMissing
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 380
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

8. Orphan and missing-downstream when NOT in a boundary directory

### IMP-graph\_test-TestNew\_DispositionSuppressesOrphan-b5783e9: TestNew\_DispositionSuppressesOrphan
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 441
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

9. Orphan suppression via disposition (no root boundary, but disposition set)

### IMP-graph\_test-TestNew\_DispositionWithoutRationale-887598d: TestNew\_DispositionWithoutRationale
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 483
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

10. Disposition without rationale — deferred/rejected with no rationale

### IMP-graph\_test-TestNew\_NeedsCoverageSatisfied-7fe97b4: TestNew\_NeedsCoverageSatisfied
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 536
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

11. `needs` attribute — explicit coverage expectations

### IMP-graph\_test-TestNew\_NeedsCoverageMissing-0c33538: TestNew\_NeedsCoverageMissing
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 588
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

### IMP-graph\_test-TestNew\_NeedsCoverageUnknownTarget-ae444a3: TestNew\_NeedsCoverageUnknownTarget
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 628
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

### IMP-graph\_test-TestNew\_NeedsEmptySaysNoCoverageExpected-6cd29a6: TestNew\_NeedsEmptySaysNoCoverageExpected
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 656
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

### IMP-graph\_test-TestNew\_NeedsSuppressesGenericOrphan-6ee3508: TestNew\_NeedsSuppressesGenericOrphan
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 681
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

### IMP-graph\_test-TestXReqmdDefault\_InheritedAndSatisfied-a0808da: TestXReqmdDefault\_InheritedAndSatisfied
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 720
x-reqmd.symbol-kind: function
trace: []
```

x-reqmd.requires-trace-from document default (SYS-FMT-003)

### IMP-graph\_test-TestXReqmdDefault\_InheritedMissingCoverageWarns-7c196f2: TestXReqmdDefault\_InheritedMissingCoverageWarns
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 747
x-reqmd.symbol-kind: function
trace: []
```

### IMP-graph\_test-TestXReqmdDefault\_EmptyDefaultOptsOutDocument-06f714a: TestXReqmdDefault\_EmptyDefaultOptsOutDocument
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 774
x-reqmd.symbol-kind: function
trace: []
```

### IMP-graph\_test-TestXReqmdDefault\_RequirementEmptyOverridesDefault-98ff597: TestXReqmdDefault\_RequirementEmptyOverridesDefault
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 802
x-reqmd.symbol-kind: function
trace: []
```

### IMP-graph\_test-TestXReqmdDefault\_RequirementTokensOverrideDefault-8fa6afb: TestXReqmdDefault\_RequirementTokensOverrideDefault
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 834
x-reqmd.symbol-kind: function
trace: []
```

### IMP-graph\_test-TestXReqmdDefault\_LevelTokenResolves-7721197: TestXReqmdDefault\_LevelTokenResolves
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 873
x-reqmd.symbol-kind: function
trace: []
```

### IMP-graph\_test-TestXReqmdDefault\_UnknownTokenWarns-59f401b: TestXReqmdDefault\_UnknownTokenWarns
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 906
x-reqmd.symbol-kind: function
trace: []
```

### IMP-graph\_test-TestXReqmdDefault\_StatusGateAppliesToInheritedDefault-88bf732: TestXReqmdDefault\_StatusGateAppliesToInheritedDefault
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 924
x-reqmd.symbol-kind: function
trace: []
```

### IMP-graph\_test-TestNew\_QualifiedTraceByDocumentID-beaad4b: TestNew\_QualifiedTraceByDocumentID
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 956
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

### IMP-graph\_test-TestNeighbors-b80a675: TestNeighbors
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 998
x-reqmd.symbol-kind: function
trace: []
```

12. Neighbors — A traces to B and C

### IMP-graph\_test-TestNeighbors\_NonExistent-be45fdd: TestNeighbors\_NonExistent
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1051
x-reqmd.symbol-kind: function
trace: []
```

13. Neighbors — non-existent reqID

### IMP-graph\_test-TestTitle-d3f4598: TestTitle
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1076
x-reqmd.symbol-kind: function
trace: []
```

14. Title

### IMP-graph\_test-TestNew\_CrossDocTrace-75db6ec: TestNew\_CrossDocTrace
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1139
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

### IMP-graph\_test-TestNew\_MultipleBrokenRefs-4d09bde: TestNew\_MultipleBrokenRefs
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1189
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

15. Multiple broken refs from same requirement

### IMP-graph\_test-TestNew\_ExternalSuppressesOrphan-edb2f77: TestNew\_ExternalSuppressesOrphan
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1225
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

16. External suppresses orphan

### IMP-graph\_test-TestNew\_ChildSuppressesOrphan-6dc8030: TestNew\_ChildSuppressesOrphan
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1253
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

17. Child requirement suppresses orphan via IsChild

### IMP-graph\_test-TestNew\_DuplicateIDAcrossDocuments-3378030: TestNew\_DuplicateIDAcrossDocuments
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1284
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

18. Duplicate ID across documents

### IMP-graph\_test-TestNew\_PrefixCollisionIsError-2c9d507: TestNew\_PrefixCollisionIsError
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1317
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

19. Prefix collision is WARNING, not ERROR

### IMP-graph\_test-TestNew\_VersionPinCurrent-689a8e7: TestNew\_VersionPinCurrent
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1347
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

20. Version-pin checks — current, outdated, predated, no-pin, missing version, external, suppressed

### IMP-graph\_test-TestNew\_VersionPinOutdated-b73f1f6: TestNew\_VersionPinOutdated
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1371
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

### IMP-graph\_test-TestNew\_VersionPinPredated-a7d0045: TestNew\_VersionPinPredated
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1412
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

### IMP-graph\_test-TestNew\_VersionPinDeterministicOrder-d48b40c: TestNew\_VersionPinDeterministicOrder
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1450
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

### IMP-graph\_test-slicesEqual-8eaea0f: slicesEqual
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1504
x-reqmd.symbol-kind: function
trace: []
```

slicesEqual is a tiny string-slice equality helper for the test above.

### IMP-graph\_test-TestNew\_VersionPinNoPinIsClean-b8c1fdd: TestNew\_VersionPinNoPinIsClean
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1516
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

### IMP-graph\_test-TestNew\_VersionPinUpstreamNoVersionIsClean-3422b1d: TestNew\_VersionPinUpstreamNoVersionIsClean
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1540
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

### IMP-graph\_test-TestNew\_VersionPinExternalUpstreamIsClean-31b24f9: TestNew\_VersionPinExternalUpstreamIsClean
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1562
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

### IMP-graph\_test-TestNew\_VersionPinSuppression-329c1f1: TestNew\_VersionPinSuppression
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1588
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

### IMP-graph\_test-TestNew\_VersionPinChildWithPin-6a4fc53: TestNew\_VersionPinChildWithPin
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1614
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

### IMP-graph\_test-TestNew\_VersionPinQualifiedRef-e26aa59: TestNew\_VersionPinQualifiedRef
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1654
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-New-16672c9
```

### IMP-graph\_test-TestStatus\_DefaultIsApproved-2cdf0f5: TestStatus\_DefaultIsApproved
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1703
x-reqmd.symbol-kind: function
trace: []
```

Status lifecycle tests
T3: missing status attr → default (approved). A node with no `status`
attr is treated as approved, so it counts as a coverage provider.

### IMP-graph\_test-TestStatus\_DraftDoesNotProvideCoverage-59c4fc5: TestStatus\_DraftDoesNotProvideCoverage
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1737
x-reqmd.symbol-kind: function
trace: []
```

T4: a draft inbound does NOT satisfy coverage. STK-001 needs [system]
but the only system doc is draft → missing-coverage warning.

### IMP-graph\_test-TestStatus\_ApprovedProvidesCoverage-591dffe: TestStatus\_ApprovedProvidesCoverage
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1772
x-reqmd.symbol-kind: function
trace: []
```

T5: an approved inbound DOES satisfy coverage.

### IMP-graph\_test-TestStatus\_ApprovedIsCaseInsensitive-8c90b6f: TestStatus\_ApprovedIsCaseInsensitive
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1803
x-reqmd.symbol-kind: function
trace: []
```

T6: case-insensitive — "APPROVED" still counts as approved.

### IMP-graph\_test-TestStatus\_IgnoreStatusOptsOut-af8b7d5: TestStatus\_IgnoreStatusOptsOut
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1835
x-reqmd.symbol-kind: function
trace: []
```

T11: ignore-status: true document → all requirements are coverage
providers regardless of their `status` value (even draft).

### IMP-graph\_test-TestStatus\_MessageMentionsDraftCount-7492021: TestStatus\_MessageMentionsDraftCount
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1868
x-reqmd.symbol-kind: function
trace: []
```

T13: updated checkNeedsCoverage message includes "(%d draft downstreams
ignored)" when at least one inbound is filtered by the status gate.

### IMP-graph\_test-TestStatus\_MessageHasNoParentheticalWhenZeroDrafts-f363bab: TestStatus\_MessageHasNoParentheticalWhenZeroDrafts
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1912
x-reqmd.symbol-kind: function
trace: []
```

T16: when draftCount == 0, message has no parenthetical. This is the
classic case where the need simply has no inbound at all (orphan-ish).

### IMP-graph\_test-TestStatus\_IgnoreAndAdditionsCombination-9d797f0: TestStatus\_IgnoreAndAdditionsCombination
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1955
x-reqmd.symbol-kind: function
trace: []
```

T19: ignore-status + additional-status-values combination. The ignore
flag bypasses the status gate entirely, AND the schema accepts the
extension value.

### IMP-graph\_test-TestStatus\_UnknownValueIsNotApproved-9baa120: TestStatus\_UnknownValueIsNotApproved
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 1995
x-reqmd.symbol-kind: function
trace: []
```

T7: schema-level reject of capitalized status. Compile fails when
user puts `status: Implemented` in a reqmd block that doesn't match
the new built-in enum. (The validator catches it; we mirror the test
here on the graph side: an uppercase status attr in the data is
treated as a non-coverage provider because it's not "approved".)

### IMP-graph\_test-TestStatus\_AdditionalPropertiesFalseGraphBuilds-f4017b9: TestStatus\_AdditionalPropertiesFalseGraphBuilds
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 2038
x-reqmd.symbol-kind: function
trace: []
```

T15 (graph-side counterpart): schemas with `additionalProperties: false`
(e.g. 00-aspice, 01-stakeholder) must still parse and inject the
built-in `status` attribute. The full Validate() path is exercised
against the real spec/ tree by the CLI check command. Here we just
confirm the graph builds without panic on a flat additionalProperties
schema that contains `status` in the resulting attribute set.

### IMP-graph\_test-TestStatus\_FullPipelineAgainstMigratedSpec-edc0f99: TestStatus\_FullPipelineAgainstMigratedSpec
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/graph_test.go
x-reqmd.source-line: 2070
x-reqmd.symbol-kind: function
trace: []
```

T14: full pipeline against the migrated spec/ tree. The 4 migrated
dirs (01a, 02, 03, 04) plus 00 and 01 must all build a graph
successfully and produce no coverage errors stemming from the status
gate. The graph must reflect the new status values without errors.

## internal/graph/repin\_test.go
### IMP-graph\_test-reqWithSuppressions-de5b2e2: reqWithSuppressions
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/repin_test.go
x-reqmd.source-line: 13
x-reqmd.symbol-kind: function
trace: []
```

reqWithSuppressions builds a Requirement with a non-empty Suppressions
slice — needed for tests that exercise reqmd-suppress handling, because
the synthetic req() factory leaves Suppressions as nil.

### IMP-graph\_test-TestRepinDeltas\_Outdated-ddde2ae: TestRepinDeltas\_Outdated
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/repin_test.go
x-reqmd.source-line: 23
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-RepinDeltas-c2e94c2
```

RepinDeltas: outdated pin

### IMP-graph\_test-TestRepinDeltas\_Current-860500f: TestRepinDeltas\_Current
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/repin_test.go
x-reqmd.source-line: 70
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-RepinDeltas-c2e94c2
```

RepinDeltas: current pin → no delta

### IMP-graph\_test-TestRepinDeltas\_Predated-d8fa80d: TestRepinDeltas\_Predated
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/repin_test.go
x-reqmd.source-line: 91
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-RepinDeltas-c2e94c2
```

RepinDeltas: predated pin → reported but not auto-fixable

### IMP-graph\_test-TestRepinDeltas\_UnpinnedNoPromote-98e2368: TestRepinDeltas\_UnpinnedNoPromote
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/repin_test.go
x-reqmd.source-line: 116
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-RepinDeltas-c2e94c2
```

RepinDeltas: unpinned ref (no promote) → empty

### IMP-graph\_test-TestRepinDeltas\_UnpinnedPromote-a30bced: TestRepinDeltas\_UnpinnedPromote
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/repin_test.go
x-reqmd.source-line: 137
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-RepinDeltas-c2e94c2
```

RepinDeltas: unpinned ref (promote) → unpinned delta

### IMP-graph\_test-TestRepinDeltas\_ExternalSkipped-6325c87: TestRepinDeltas\_ExternalSkipped
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/repin_test.go
x-reqmd.source-line: 172
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-RepinDeltas-c2e94c2
```

RepinDeltas: external upstream → skipped

### IMP-graph\_test-TestRepinDeltas\_UnversionedSkipped-24d2f96: TestRepinDeltas\_UnversionedSkipped
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/repin_test.go
x-reqmd.source-line: 193
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-RepinDeltas-c2e94c2
```

RepinDeltas: unversioned upstream → skipped

### IMP-graph\_test-TestRepinDeltas\_Suppressed-a3ca150: TestRepinDeltas\_Suppressed
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/repin_test.go
x-reqmd.source-line: 214
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-RepinDeltas-c2e94c2
```

RepinDeltas: suppressed node → skipped

### IMP-graph\_test-TestRepinDeltas\_QualifiedSourceRef-e928c8e: TestRepinDeltas\_QualifiedSourceRef
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/repin_test.go
x-reqmd.source-line: 235
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-RepinDeltas-c2e94c2
```

RepinDeltas: qualified ref (doc-id/ID~N) → SourceRef preserves the form

### IMP-graph\_test-TestRepinDeltas\_ResultSkipped-f14a86d: TestRepinDeltas\_ResultSkipped
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/repin_test.go
x-reqmd.source-line: 260
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-RepinDeltas-c2e94c2
```

RepinDeltas: result pseudo-requirement is not a source

### IMP-graph\_test-TestRepinDeltas\_SyntheticSkipped-24f7093: TestRepinDeltas\_SyntheticSkipped
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/repin_test.go
x-reqmd.source-line: 284
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-RepinDeltas-c2e94c2
```

RepinDeltas: any node from a Synthetic document is not a source
(future synthesized test-case nodes must not be rewritten in the report)

### IMP-graph\_test-TestRepinDeltas\_Sorted-ddd1237: TestRepinDeltas\_Sorted
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/repin_test.go
x-reqmd.source-line: 313
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-RepinDeltas-c2e94c2
```

RepinDeltas: stable sort by file → reqID → sourceRef

## internal/graph/verify\_test.go
### IMP-graph\_test-resultReq-fbde745: resultReq
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 13
x-reqmd.symbol-kind: function
trace: []
```

resultReq builds a synthesized verification-result pseudo-requirement
tracing to measureID with the given outcome.

### IMP-graph\_test-measureReq-794b472: measureReq
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 27
x-reqmd.symbol-kind: function
trace: []
```

measureReq builds a verification-measure requirement (an authored req
that a result traces to). By default it has status approved.

### IMP-graph\_test-resultDoc-55a42cb: resultDoc
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 39
x-reqmd.symbol-kind: function
trace: []
```

### IMP-graph\_test-measureDoc-66866b4: measureDoc
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 47
x-reqmd.symbol-kind: function
trace: []
```

### IMP-graph\_test-TestCheckResults\_NoResults\_NoOutcomeChecks-3a1b22b: TestCheckResults\_NoResults\_NoOutcomeChecks
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 55
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-CheckResults-f98967d
```

No result nodes → outcome checks are no-ops

### IMP-graph\_test-TestMissingVerdict\_MeasureWithoutResult-ccad749: TestMissingVerdict\_MeasureWithoutResult
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 74
x-reqmd.symbol-kind: function
trace: []
```

missing-verdict: approved measure with no result → WARNING

### IMP-graph\_test-TestMissingVerdict\_MeasureWithResult\_NoFinding-c5f09be: TestMissingVerdict\_MeasureWithResult\_NoFinding
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 97
x-reqmd.symbol-kind: function
trace: []
```

### IMP-graph\_test-TestMissingVerdict\_DraftMeasureSkipped-20f688f: TestMissingVerdict\_DraftMeasureSkipped
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 113
x-reqmd.symbol-kind: function
trace: []
```

### IMP-graph\_test-TestFailingVerdict\_MeasureFails-04c7e51: TestFailingVerdict\_MeasureFails
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 133
x-reqmd.symbol-kind: function
trace: []
```

failing-verdict: measure with fail result → ERROR

### IMP-graph\_test-TestFailingVerdict\_MeasurePasses\_NoFinding-584ebda: TestFailingVerdict\_MeasurePasses\_NoFinding
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 156
x-reqmd.symbol-kind: function
trace: []
```

### IMP-graph\_test-TestVersionPin\_StaleResultPin-84d772d: TestVersionPin\_StaleResultPin
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 176
x-reqmd.symbol-kind: function
trace: []
```

version-pin reuse: stale result pin triggers existing check

### IMP-graph\_test-staleResultReq-ef37d32: staleResultReq
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 203
x-reqmd.symbol-kind: function
trace: []
```

staleResultReq builds a result whose trace carries a ~N pin.

### IMP-graph\_test-bootTimeCase-3f1bc0f: bootTimeCase
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 221
x-reqmd.symbol-kind: const
trace: []
```

case-keyed chains (patterns C/D/E): cased results + synthesized TC nodes
bootTimeCase is the case key used by the case-keyed chain fixtures.

### IMP-graph\_test-casedResultReq-41bde99: casedResultReq
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 224
x-reqmd.symbol-kind: function
trace: []
```

casedResultReq builds a case-keyed result node tracing to the given target.

### IMP-graph\_test-tcNodeReq-237c305: tcNodeReq
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 237
x-reqmd.symbol-kind: function
trace: []
```

tcNodeReq builds a synthesized test-case node tracing to its verifies refs.

### IMP-graph\_test-TestCaseKeyedResult\_MeasureChecks-2b4ae57: TestCaseKeyedResult\_MeasureChecks
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 251
x-reqmd.symbol-kind: function
trace: []
```

A cased result (pattern C) attaches to an authored measure and flows
through the existing checks like any other result.

### IMP-graph\_test-TestSynthChain\_TCToRequirement\_NoBrokenRefs-02b40b6: TestSynthChain\_TCToRequirement\_NoBrokenRefs
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 282
x-reqmd.symbol-kind: function
trace: []
```

A synthesized test case (patterns D/E) traces to the requirements it
verifies; result → case → requirement edges resolve without broken refs.
The requirement's rolled-up evidence set includes the passing case, so no
missing-verdict fires (this is the roll-up behavior landed in the
evidence-set phase).

### IMP-graph\_test-TestRollup\_MixedEvidenceOneFailingFails-b155496: TestRollup\_MixedEvidenceOneFailingFails
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 328
x-reqmd.symbol-kind: function
trace: []
```

Mixed evidence on one requirement: a direct passing result (A) plus a
failing case on an authored downstream test (C). Any fail → fail.

### IMP-graph\_test-TestRollup\_DraftTCIgnored-87b3c65: TestRollup\_DraftTCIgnored
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 356
x-reqmd.symbol-kind: function
trace: []
```

A draft downstream test case does not contribute evidence; the requirement
stays missing-verdict and the message reports the ignored draft.

### IMP-graph\_test-TestRollup\_FilterExcludesTestsDoc-e6e2079: TestRollup\_FilterExcludesTestsDoc
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 383
x-reqmd.symbol-kind: function
trace: []
```

Under --filter, evidence is computed within the filtered subset: excluding
the tests document leaves the requirement uncovered without a false fail.

### IMP-graph\_test-TestVerdict\_DeferredDispositionSuppresses-7766189: TestVerdict\_DeferredDispositionSuppresses
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 413
x-reqmd.symbol-kind: function
trace: []
```

Deferred/rejected dispositions do not expect verification: neither
missing-verdict nor failing-verdict fires for them.

### IMP-graph\_test-TestMeasureVerdicts\_EvidenceCarriesDescription-c526d97: TestMeasureVerdicts\_EvidenceCarriesDescription
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/graph/verify_test.go
x-reqmd.source-line: 435
x-reqmd.symbol-kind: function
trace:
  - graph/IMP-graph-MeasureVerdicts-ce3940b
```

The evidence items surfaced by MeasureVerdicts carry the synthesized test
case's Markdown description, so exporters can render it inline.

