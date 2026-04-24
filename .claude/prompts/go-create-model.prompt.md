---
agent: agent
description: "Create a new model in the Go API model layer following best practices."
---

# Create/Update Model (Go)

Follow these instruction files (treat them as **always-on** for this task and subsequent turns in this session):
- [Go Model Instructions](../instructions/go-model.instructions.md)
- [Go Base Instructions](../instructions/go-base.instructions.md)
- [Go Style Guide](../instructions/go-style.instructions.md)
- [Go API Structure Guidelines](../instructions/go-structure.instructions.md)

## What to do first (interactive)

Before writing code, ask me these questions if needed, then summarize the plan:
1) **table/entity**: SQLBoiler entity/table (e.g., `products`)
2) **file name**: model file name (default: `<singular>.go`, e.g., `product.go`)
3) **struct name**: domain model name (default: `Product`)
4) **fields & types**: which columns to include, custom type mappings (enums, timestamps, decimals)
5) **nullability policy**: pointers vs custom nullable types for optional columns
6) **validation**: required fields, ranges, format checks; `IsValid()` expectations

After confirmation, print a short **plan** (file path, struct shape, helper methods, tests/docs), then apply edits.

## Hard rules

- **Do not modify** generated ORM (`api/internal/repository/orm/**`).
- Model types live in `api/internal/model/` and **mirror domain needs**, not raw DB layouts.
- Prefer value types where practical; use pointers only for truly optional/nullable fields.
- Time fields: use `time.Time` (UTC); document zero-time semantics if applicable.
- Add `IsValid()` (and small validators) for critical invariants; **no DB or HTTP logic** here.
- Keep the model **free of side effects**; no logging or I/O.


## Implementation tips

- Map ORM → model intentionally: include only fields needed by services/handlers. Include all the fields in the ORM if unsure.
- For nullable DB columns, choose either pointer fields or a small wrapper type (follow repo policy).
- Keep zero values meaningful; prefer explicit constructors only when invariants require them.
- If adding IDs, document ID type (`int64`, Snowflake, UUID) and string render rules.
- Keep files small and cohesive; one domain model per file (`product.go`).

## Deliverables

- [ ] Model file at `api/internal/model/{file}.go` with the primary struct and helpers
- [ ] Inline **doc comments** on exported types/fields and any non-obvious semantics
- [ ] (Optional) Small table-driven unit tests `api/internal/model/{file}_test.go` for `IsValid()`/other methods
- [ ] If enums added, include a brief comment listing allowed values and parsing rules
