# MCP

MCP support is planned. Treat this as a design placeholder until implementation lands.

## Intended Role

Expose Tero capabilities to MCP-compatible hosts while reusing internal services.

## Design Constraints

1. No business logic in transport handlers.
2. Reuse existing domain/service boundaries.
3. Keep request/response contracts explicit and versionable.

## Initial Implementation Plan

1. Define tool/resource surface area.
2. Add transport adapter and capability registration.
3. Add contract tests for request validation and response shape.

