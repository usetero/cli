# MCP

Status: planned. This document defines the intended architecture so implementation stays consistent.

## Purpose

Expose a safe, typed subset of Tero capabilities to MCP-compatible clients.

## Design Constraints

1. Transport layer contains no business logic.
2. Reuse existing domain/service boundaries.
3. Define explicit request/response schemas.
4. Version behavior changes intentionally.

## Implementation Plan

1. Define tool/resource surface and capability boundaries.
2. Add transport adapter and registration.
3. Add contract tests for request validation and response shape.
4. Add error taxonomy mapping to MCP response semantics.

## Testing Expectations

1. Contract tests for schemas and validation.
2. Behavior tests for successful and failing calls.
3. No hidden coupling to UI-only types.
