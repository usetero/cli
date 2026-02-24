# Domain Package

Shared types for the Tero CLI. Every layer — TUI, CLI, MCP, chat tools — imports domain for its data models. Domain never imports from any of them.

## How It Works

### Two Layers: Raw and Rich

Every entity has two types. Take policies:

- **`PolicyCard`** is the raw database DTO. Strings, nullable floats, raw JSON. It mirrors the SQLite row exactly. No methods, no parsing. This is what `sqlc` generates and what queries return.

- **`Policy`** is the rich model. Typed enums (`PolicyAction`, `PolicyStatus`, `PolicySeverity`), parsed analysis (`PolicyAnalysis` interface), parsed log examples, and derived display values. This is what every consumer actually works with.

The bridge is a constructor: `ParsePolicy(card *PolicyCard) *Policy`. It parses JSON, converts strings to typed enums, and extracts derived data. Parse once, use everywhere.

```
SQLite row → PolicyCard (raw) → ParsePolicy() → Policy (rich) → any consumer
```

### Display Methods Live on the Rich Model

`Policy` has methods that return human-readable derived values:

```go
p.Headline()    // "Sample — keep ~1% of volume"
p.Mechanism()   // "Approving keeps 1 in every ~8,837 events."
p.Pitch()       // First sentence of rationale
p.CostPerYear() // "~$16.3k/yr" or ""
p.Impact()      // Before/after metrics, or nil
```

These are **not rendering** — they're data derivation. Any consumer wants the same headline string whether it's the TUI card, a CLI summary, an MCP response, or AI context. The principle: domain answers "what to display." Renderers answer "where and how."

If you're computing a derived value from `Policy` fields and it would be useful to more than one consumer, it belongs here as a method. If it's specific to a particular layout (padding, column widths, lipgloss styles), it belongs in the renderer.

### The PolicyAnalysis Interface

Each policy category (health_checks, bot_traffic, pii_leakage, etc.) has its own analysis struct with category-specific fields. They all implement `PolicyAnalysis`:

```go
type PolicyAnalysis interface {
    Category() string      // "bot_traffic"
    Subtitle() string      // "~38% bot/crawler traffic"
    Rationale() string     // Full rationale text
    ActionDetail() string  // "3 duplicate fields"
    RelevantKeys() []string // ["http.user_agent"]
}
```

Categories that only have a rationale embed `baseAnalysis` and get default implementations for free. Categories with structured data (fields, pairs, levels) override the methods that matter.

Adding a new category:
1. Add the analysis struct to the appropriate `policy_analysis_*.go` file
2. Embed `baseAnalysis`, override methods that have real data
3. Add a case to `parseAnalysis()` in `policy_analysis.go`

## File Organization

```
policy.go                       Policy struct + ParsePolicy constructor
policy_card.go                  PolicyCard — raw DB DTO
policy_display.go               Display methods on Policy (Headline, Mechanism, etc.)
policy_analysis.go              PolicyAnalysis interface + parseAnalysis switch
policy_analysis_waste.go        8 waste category types
policy_analysis_quality.go      4 quality category types
policy_analysis_compliance.go   4 compliance category types + SensitiveField
```

## Rules

1. **Domain never imports rendering packages.** No lipgloss, no theme, no styles. If you need `lipgloss` you're in the wrong layer.
2. **PolicyCard is immutable.** Never add methods to it. It's a data transfer object.
3. **Typed enums over raw strings.** `PolicyAction`, `PolicyStatus`, `PolicySeverity`, `CategoryType` — use them. Add new ones when a string field has a known set of values.
4. **Parse once.** `ParsePolicy` does all JSON parsing. Consumers never unmarshal analysis JSON themselves.
