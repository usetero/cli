# MCP Interface

Status: planned.

This interface should be treated as a transport adapter, not a new execution
model. The right way to build MCP here is to expose existing capabilities with
clear contracts, while preserving the same architecture boundaries used by TUI
and CLI surfaces.
If MCP starts to look like a parallel domain layer, the design has drifted.

## Design intent

MCP should translate typed tool/resource calls into existing service operations.
It should not add business logic that diverges from control plane behavior or
from existing command/runtime semantics.

## Non-negotiable constraints

Keep these constraints explicit as implementation starts:

1. transport is thin; policy remains upstream,
2. request/response schemas are explicit and versioned,
3. errors are mapped intentionally (not raw-internal leakage),
4. behavior is contract-tested at the interface boundary.

## Practical implementation guidance

When this interface is implemented, mirror the same composition style used in
`internal/cmd/root.go`: wire dependencies once, keep handlers narrow, and push
shared behavior into existing service/domain layers.
That keeps the MCP surface consistent with existing interfaces and avoids a
third policy implementation.
