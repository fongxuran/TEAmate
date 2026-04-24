---
description: "Create or update a Go controller function following clean architecture and project standards."
agent: "agent"
---

# Create/Update Controller Function (Go)

Follow these instruction files (treat them as **always-on** for this task and subsequent turns in this session):
- [Go Controller Layer Instructions](../instructions/go-ctrl.instructions.md)
- [Go Base Instructions](../instructions/go-base.instructions.md)
- [Go Style Guide](../instructions/go-style.instructions.md)
- [Go API Structure Guidelines](../instructions/go-structure.instructions.md)

## What to do first (interactive)
Before writing code, ask me these questions if needed, then summarize the plan:


1) **domain_package** (singular, under `api/internal/controller/`, e.g. `orders`)
2) **operation**: `list|get|create|update|delete|custom`
3) **business rules**: any specific logic to implement
4) **repository_method**: which repository function to call (e.g., `List`, `Get`)
5) **function_name** (optional; otherwise propose)
6) **params** (e.g., `ctx context.Context`, `id:int64`, `payload:Order`, or controller-layer input struct)
7) **returns** (if not the default)
8) **validation**: any business rules, invariants, or checks
9) **side effects**: any events, logging, or notifications

After confirmation, print a short **plan** (files, signatures, tests/docs), then apply edits.


## Hard rules
- Follow controller layer instructions.
- Accept `context.Context` first in every method.
- Always define controller-layer input structs for function parameters; do not use repository structs in controller function parameters.
- Always define controller-layer output structs if needed; do not return repository or ORM types.
- Call repository methods for data access; do not embed SQLBoiler or DB logic.
- Apply business logic, validation, and orchestration in the controller.
- Do not leak repository or ORM types; use domain models.
- Keep controller methods side-effect free except for intended business actions.
- Place the controller interface and constructor in `new.go` only.
- Group related functions logically: one fn per file or a small cohesive group.

## Deliverables
- [ ] Interface updated in `api/internal/controller/{domain_package}/new.go`
- [ ] Implementation file in `api/internal/controller/{domain_package}/{snake_case(function_name)}.go`
<!-- - [ ] (Optional) Table-driven unit tests `*_test.go` covering business logic and error cases -->
- [ ] Brief top-of-file comment if behavior is non-trivial (validation, orchestration, tx)


## Example signatures (for reference)

In package `order`:
- list:   `List(ctx context.Context, input ListInput) ([]model.Order, error)`
- get:    `Get(ctx context.Context, id int64) (model.Order, error)`
- create: `Create(ctx context.Context, input CreateInput) (model.Order, error)`
- update: `Update(ctx context.Context, input UpdateInput) (model.Order, error)`
- delete: `Delete(ctx context.Context, id int64) error`
