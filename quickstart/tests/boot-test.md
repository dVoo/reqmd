# Boot Tests

## TEST-001: Boot Time Measurement
```attr
test-type: unit
requires-trace-from: []
trace: [software/SW-001, system/SYS-001]
```
The test shall measure boot time on reference hardware and assert it is ≤ 5 s.
