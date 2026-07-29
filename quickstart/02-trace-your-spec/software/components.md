# Software Components

## SW-001: Boot Manager
```attr
status: draft
requires-trace-from: [tests]
trace: [system/SYS-001]
```
The boot manager shall initialize the display driver and launch the UI shell within 4.5 seconds.

*Rationale:* Leaves 0.5 s margin for hardware variation.
