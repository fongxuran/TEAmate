---
agent: 'agent'
description: 'Create or update a Go repository function using SQLBoiler and repo standards.'
---

# Create/Update Repository Function (Go)

Follow these instruction files:
- [Go Repository Layer Instructions](../instructions/go-repo.instructions.md)
- [Go Base Instructions](../instructions/go-base.instructions.md)
- [Go Style Guide](../instructions/go-style.instructions.md)
- [Go API Structure Guidelines](../instructions/go-structure.instructions.md)

> For this task **and subsequent chat turns in this session**, continue to apply the rules above without being reminded.

## What to do first (interactive)
Before writing code, ask me these questions, then summarize the plan:

1) **domain_package** (plural, under `api/internal/repository/`, e.g. `products`)
2) **operation**: `list|get|create|update|delete|count|exists|custom`
3) **model**: domain model name (e.g., `Product`)
4) **table/entity**: SQLBoiler entity/table (e.g., `products`)
5) **function_name** (optional; otherwise propose)
6) **filters** (if any): fields & types, pagination, sorting allow-list
7) **params** (e.g., `id:int64`, `payload:Product`)
8) **returns** (if not the default)
9) **lock_mode** (optional): e.g., `UPDATE`

After confirming, print a short plan (files, signatures, tests), then apply edits.

## Hard rules
- Follow repository layer instructions.
- Accept `context.Context` first in every method.
- Do **not** return ORM entities; convert `orm.*` → `model.*` before returning.
- Map `sql.ErrNoRows` → `ErrNotFound`; wrap errors with stack traces.
- Apply WHERE clauses **only** for provided (non-zero) filters.
- Guard pagination (max page size) and **allow-list** sort fields/directions.
- Use SQLBoiler `qm` mods for queries; use `qm.For("<LOCK>")` only if `lock_mode` is given.
- For partial updates, use `boil.Whitelist(...)` with explicit columns.
- Keep filenames coherent: one fn per file (`get_product.go`) or a small cohesive group (`products.go`).
- Create an `errors.go` file with a package-level errors like `ErrNotFound` if it doesn't exist.
- Place the repository interface and constructor in `new.go` only.
- Register in `registry.go` if the project pattern requires it.

## Deliverables
- [ ] Interface updated in `api/internal/repository/{domain_package}/new.go`
- [ ] Implementation file in `api/internal/repository/{domain_package}/{snake_case(function_name)}.go`
- [ ] Optional filter struct co-located with the function
<!-- - [ ] Table-driven unit tests `*_test.go` covering success / not-found / error -->
- [ ] Brief top-of-file comment if behavior is non-trivial (locking, tx)

## Example signatures (for reference)

In package `products`:
- list:   `List(ctx context.Context, filter ListFilter) ([]model.Product, error)`
- get:    `Get(ctx context.Context, id int64) (model.Product, error)`
- create: `Create(ctx context.Context, product model.Product) (model.Product, error)`
- update: `Update(ctx context.Context, product model.Product) (model.Product, error)`
- delete: `Delete(ctx context.Context, id int64) error`
- count:  `Count(ctx context.Context, filter CountFilter) (int64, error)`
- exists: `Exists(ctx context.Context, id int64) (bool, error)`
