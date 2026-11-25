# Package: verify\_test

<!-- reqmd: generated -->

## internal/verify/verify\_test.go
### IMP-verify\_test-writeFixture-e0c0c1b: writeFixture
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 15
x-reqmd.symbol-kind: function
trace: []
```

writeFixture writes fixture content to a file, failing the test on error.

### IMP-verify\_test-ctrfReport-79379c6: ctrfReport
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 27
x-reqmd.symbol-kind: function
trace: []
```

CTRF parser
ctrfReport builds a minimal CTRF JSON report with the given tests.

### IMP-verify\_test-TestLoadCTRF\_PassFail-b91f1a4: TestLoadCTRF\_PassFail
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 37
x-reqmd.symbol-kind: function
trace: []
```

### IMP-verify\_test-TestLoadCTRF\_UninstrumentedTestSkippedSilently-d9022ec: TestLoadCTRF\_UninstrumentedTestSkippedSilently
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 78
x-reqmd.symbol-kind: function
trace: []
```

An uninstrumented test (no extra.x-reqmd block) is skipped silently:
the old per-test "no x-reqmd.id" warning is gone so suites can adopt
reqmd incrementally without spamming every uninstrumented test.

### IMP-verify\_test-TestLoadCTRF\_UnboundResultWarns-96e81b8: TestLoadCTRF\_UnboundResultWarns
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 102
x-reqmd.symbol-kind: function
trace: []
```

A test with an x-reqmd block but nothing bindable (no id, no verifies)
yields an unbound-result warning.

### IMP-verify\_test-TestLoadCTRF\_VersionPinPreserved-f1c65fe: TestLoadCTRF\_VersionPinPreserved
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 127
x-reqmd.symbol-kind: function
trace: []
```

### IMP-verify\_test-TestLoadCTRF\_FlakyIsInconclusive-ce6325e: TestLoadCTRF\_FlakyIsInconclusive
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 149
x-reqmd.symbol-kind: function
trace: []
```

### IMP-verify\_test-TestMergeLatest\_LatestByVerifiedAt-2266f70: TestMergeLatest\_LatestByVerifiedAt
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 173
x-reqmd.symbol-kind: function
trace:
  - verify/IMP-verify-MergeLatest-79e3025
```

Multi-file merge

### IMP-verify\_test-TestMergeLatest\_PinStrippedFromKey-f51f924: TestMergeLatest\_PinStrippedFromKey
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 187
x-reqmd.symbol-kind: function
trace:
  - verify/IMP-verify-MergeLatest-79e3025
```

### IMP-verify\_test-TestLoadResults\_NonCTRFJsonSkipped-1349326: TestLoadResults\_NonCTRFJsonSkipped
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 202
x-reqmd.symbol-kind: function
trace:
  - verify/IMP-verify-LoadResults-e444d2c
```

Auto-detection (non-CTRF json skipped)

### IMP-verify\_test-TestLoadResults\_PlainJsonCTRFShapeAccepted-f530cff: TestLoadResults\_PlainJsonCTRFShapeAccepted
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 221
x-reqmd.symbol-kind: function
trace:
  - verify/IMP-verify-LoadResults-e444d2c
```

### IMP-verify\_test-TestLoadCTRF\_ResultsArrayRejected-fc5eec4: TestLoadCTRF\_ResultsArrayRejected
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 238
x-reqmd.symbol-kind: function
trace: []
```

### IMP-verify\_test-TestLoadCTRF\_ResultsArrayJsonSkippedSilently-e97acae: TestLoadCTRF\_ResultsArrayJsonSkippedSilently
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 250
x-reqmd.symbol-kind: function
trace: []
```

### IMP-verify\_test-TestLoadCTRF\_MalformedJSONError-23d5f1e: TestLoadCTRF\_MalformedJSONError
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 265
x-reqmd.symbol-kind: function
trace: []
```

### IMP-verify\_test-TestSynthesize\_ResultNodeIDAndTrace-0191e26: TestSynthesize\_ResultNodeIDAndTrace
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 281
x-reqmd.symbol-kind: function
trace:
  - verify/IMP-verify-Synthesize-d4769b6
```

Synthesize

### IMP-verify\_test-TestIsResultNode-d92525e: TestIsResultNode
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 307
x-reqmd.symbol-kind: function
trace:
  - verify/IMP-verify-IsResultNode-1bd2c9b
```

### IMP-verify\_test-TestLoadCTRF\_IDWithCase\_RecordsCaseKey-709880c: TestLoadCTRF\_IDWithCase\_RecordsCaseKey
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 327
x-reqmd.symbol-kind: function
trace: []
```

v2 binding (patterns C/D/E)

### IMP-verify\_test-TestLoadCTRF\_CaseAndVerifies\_SynthesizesCaseBinding-9262656: TestLoadCTRF\_CaseAndVerifies\_SynthesizesCaseBinding
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 353
x-reqmd.symbol-kind: function
trace: []
```

### IMP-verify\_test-TestLoadCTRF\_VerifiesOnly\_ImplicitCaseFromSuiteName-838ccf7: TestLoadCTRF\_VerifiesOnly\_ImplicitCaseFromSuiteName
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 389
x-reqmd.symbol-kind: function
trace: []
```

### IMP-verify\_test-TestSynthesize\_CaseKeyedCreatesTCAndCasedResults-02d0ab5: TestSynthesize\_CaseKeyedCreatesTCAndCasedResults
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 425
x-reqmd.symbol-kind: function
trace:
  - verify/IMP-verify-Synthesize-d4769b6
```

### IMP-verify\_test-TestIsTestCaseNode-4bdc873: TestIsTestCaseNode
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/verify/verify_test.go
x-reqmd.source-line: 473
x-reqmd.symbol-kind: function
trace:
  - verify/IMP-verify-IsTestCaseNode-5f32cca
```

