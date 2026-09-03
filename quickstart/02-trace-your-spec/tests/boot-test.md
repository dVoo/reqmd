# Boot Tests

## TEST-001: Boot Time Measurement
```attr
status: approved
verify: Test
test-type: unit
disposition: implemented
trace: [software/SW-001, system/SYS-001]
```
The test shall measure boot time on reference hardware and assert it is ≤ 5 s.

## TEST-002: Boot Sequence Design Review
```attr
status: approved
verify: Review
test-type: unit
disposition: implemented
trace: [system/SYS-001]
```
A design review shall confirm the boot sequence covers all required initialization steps.

*Rationale:* Not every verification is automated. Manual methods (Review, Inspection, Analysis, Demonstration) are documented as measures and their results recorded as markdown outside the spec tree.