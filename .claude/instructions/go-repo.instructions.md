---
description: 'Instructions for implementing Go repository-layer code following SQLBoiler and internal best practices.'
applyTo: 'api/internal/repository/*.go,!gqlgenerated.go,!**/*_gen.go,!**/mock_*.go'
---

# Repository Layer Guidelines

This instruction governs **data-access implementation** inside the `api/internal/repository` package.  
Repositories are responsible for abstracting database operations, providing a clean, domain-oriented API to the service layer.

Refer also to:
- [Base Go standards](go-base.instructions.md)
- [Go API structure guidelines](go-structure.instructions.md)
- [Go Style Guide](go-style.instructions.md)

## Purpose

- Encapsulate and reuse database logic behind clear interfaces.
- Decouple upper layers from specific DB engines.
- Keep code composable, mockable, and testable.
- Follow SQLBoiler best practices for performance and maintainability.

## Design Principles

✅ **Do**
- Organize by **domain** (e.g., `products`, `orders`, `inventory`).
- Use **plural package names** (e.g., `users`, not `user`).
- Define interfaces that expose CRUD and domain-specific queries.
- Convert ORM entities (`orm.*`) → domain models (`model.*`) **before returning**.
- Ensure all functions accept a `context.Context`.
- Keep repository methods **side-effect free** except for intended data writes.

🚫 **Don’t**
- Don’t leak ORM structs outside the repository.
- Don’t embed `sqlboiler` queries directly in service or handler layers.
- Don’t handle HTTP-specific errors or logging logic here.
- Don’t silently swallow `sql.ErrNoRows`; return a defined `ErrNotFound`.

## Package Skeleton

- Place constructor and interface in a single `new.go` file under same package.
    - Keep a single implementation type (e.g., `impl`) and inject dependencies there.
```go
package inventory

import (
	"context"

	"code.in.spdigital.sg/sp-digital/athena/db/pg"
	"code.in.spdigital.sg/sp-digital/pencilcase/api/internal/model"
)

// Repository provides the specification of the functionality provided by this pkg
type Repository interface {
	// ListProducts gets a list of products from DB
	ListProducts(context.Context, ProductsFilter) ([]model.Product, error)
	// CreateProduct saves product in DB
	CreateProduct(context.Context, model.Product) (model.Product, error)
	// UpdateProductStatus updates the product status in DB
	UpdateProductStatus(context.Context, model.Product) (model.Product, error)
}

// New returns an implementation instance satisfying Repository
func New(pgConn pg.ContextExecutor) Repository {
	return impl{pgConn: pgConn}
}

type impl struct {
	pgConn      pg.ContextExecutor
}
```

## Implementation Rules

- Each repository package should have a single `impl` struct (e.g., `type impl struct { pgConn boil.ContextExecutor }`) in the `new.go` file.
- Define a `Repository` interface in the same `new.go` file, listing all methods can be used by the upper layers.
- Repository interface and constructor must be defined in a single new.go file within the domain package. Do not create a separate interface.go.
- Group related functions logically:
    - One file per core function (e.g., `get_product.go`)
    - Or one file per small domain (e.g., `products.go` for List/Get).
- Register repositories in `registry.go` if required by the project pattern.

### Parameter and Return Types

- For CRUD:
    - Use domain model structs for input/output.
- For queries:
    - Define lightweight filter structs (e.g., `ProductsFilter`) beside query functions.
- Return `([]model.Product, error)` or `(model.Product, error)` — not pointers unless necessary.
- Add pagination, sorting, and optional filters gracefully:
  ```go
  if filter.Status != nil {
      qms = append(qms, orm.ProductWhere.Status.IN(filter.Status))
  }
  ```

### Query Practices

- Always prefer **SQLBoiler query mods (qm.*)** over manual SQL.
- Explicitly list columns when doing partial updates with `boil.Whitelist`.
- Use `qm.For("UPDATE")` for locking semantics where required.
- Handle nil/empty filter fields safely before appending to query mods.

### Error & Transaction Handling

- Place domain-specific error variables (e.g., `ErrNotFound`, `ErrCacheMiss`) in api/internal/repository/{domain}/errors.go.
- Export errors as package-level variables, with clear comments.
- Reference these errors in repository implementations (e.g., return `ErrNotFound` for not found cases).
  **Do not duplicate** error definitions in multiple files; centralize in `errors.go` per domain package.
- Wrap errors using `pkgerrors.WithStack(err)`  for better traceability if error is from the vendor library.
- Check `sql.ErrNoRows` explicitly:
  ```go
  if errors.Is(err, sql.ErrNoRows) {
	  return model.Product{}, ErrNotFound
  }
  ```

### Testing & Performance

- Write table-driven tests for each repository function.
- Use in-memory or test DB; mock external dependencies (e.g., Redis).
- Benchmark critical queries under realistic dataset sizes (go test -bench).
- Ensure each test runs independently and cleans up DB state.

## Example Implementation

```go
package inventory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"code.in.spdigital.sg/sp-digital/athena/cache/redis"
	"code.in.spdigital.sg/sp-digital/bobcat/api/internal/model"
	"code.in.spdigital.sg/sp-digital/bobcat/api/internal/repository/generator"
	"code.in.spdigital.sg/sp-digital/bobcat/api/internal/repository/orm"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	pkgerrors "github.com/pkg/errors"
)

const (
	cacheKeyObjTypeActiveProductsCount = "active-products-count"
	cacheKeyActiveProductsCount        = "all"
)

// ProductsFilter holds filters for getting products list
type ProductsFilter struct {
	ExtID    string
	Status   []model.ProductStatus
	WithLock bool
}

// ListProducts gets a list of products from DB
func (i impl) ListProducts(ctx context.Context, filter ProductsFilter) ([]model.Product, error) {
	qms := []qm.QueryMod{
		qm.OrderBy(orm.ProductColumns.ID + " DESC"),
	}
	if filter.Status != nil {
		status := make([]string, len(filter.Status))
		for idx, s := range filter.Status {
			status[idx] = s.String()
		}
		qms = append(qms, orm.ProductWhere.Status.IN(status))
	}
	if filter.ExtID != "" {
		qms = append(qms, orm.ProductWhere.ExternalID.EQ(filter.ExtID))
	}
	if filter.WithLock {
		qms = append(qms, qm.For("UPDATE"))
	}
	slice, err := orm.Products(qms...).All(ctx, i.pgConn)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	result := make([]model.Product, len(slice))
	for idx, o := range slice {
		result[idx] = model.Product{
			ID:          o.ID,
			ExternalID:  o.ExternalID,
			Name:        o.Name,
			Description: o.Description,
			Status:      model.ProductStatus(o.Status),
			Price:       o.Price,
			CreatedAt:   o.CreatedAt,
			UpdatedAt:   o.UpdatedAt,
		}
	}

	return result, nil
}

// CreateProduct saves product in DB
func (i impl) CreateProduct(ctx context.Context, p model.Product) (model.Product, error) {
	id, err := generator.ProductIDSNF.Generate()
	if err != nil {
		return p, pkgerrors.WithStack(err)
	}
	o := orm.Product{
		// ID:          id, // TODO: Switch to snowflake after changing column to BIGINT non-serial
		ExternalID:  fmt.Sprint(id), // TODO: Should be snowflake as well and be int64
		Name:        p.Name,
		Description: p.Description,
		Status:      p.Status.String(),
		Price:       p.Price,
	}

	if err := o.Insert(ctx, i.pgConn, boil.Infer()); err != nil {
		return p, pkgerrors.WithStack(err)
	}

	p.ID = o.ID
	p.CreatedAt = o.CreatedAt
	p.UpdatedAt = o.UpdatedAt

	return p, nil
}

// UpdateProductStatus updates the product status in DB
func (i impl) UpdateProductStatus(ctx context.Context, p model.Product) (model.Product, error) {
	o, err := orm.Products(orm.ProductWhere.ID.EQ(p.ID)).One(ctx, i.pgConn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return p, ErrNotFound
		}
		return p, pkgerrors.WithStack(err)
	}

	o.Status = p.Status.String()
	rows, err := o.Update(ctx, i.pgConn, boil.Whitelist(orm.ProductColumns.Status, orm.ProductColumns.UpdatedAt))
	if err != nil {
		return p, pkgerrors.WithStack(err)
	}

	if rows != 1 {
		return p, pkgerrors.WithStack(fmt.Errorf("%w, found: %d", ErrUnexpectedRowsFound, rows))
	}

	p.UpdatedAt = o.UpdatedAt

	return p, nil
}

// GetActiveProductsCountFromDB gets active products count from DB
func (i impl) GetActiveProductsCountFromDB(ctx context.Context) (int64, error) {
	result, err := orm.Products(orm.ProductWhere.Status.EQ(model.ProductStatusActive.String())).Count(ctx, i.pgConn)
	return result, pkgerrors.WithStack(err)
}

// GetCachedActiveProductsCount gets active products count from cache
func (i impl) GetCachedActiveProductsCount(ctx context.Context) (int64, error) {
	count, err := i.redisClient.GetInt64(ctx, cacheKeyObjTypeActiveProductsCount, cacheKeyActiveProductsCount)
	if err != nil {
		if errors.Is(err, redis.ErrNilReply) {
			return 0, ErrCacheMiss
		}
		return 0, pkgerrors.WithStack(err)
	}

	return count, nil
}

// CacheActiveProductsCount caches active products count in cache
func (i impl) CacheActiveProductsCount(ctx context.Context, count int64) error {
	return pkgerrors.WithStack(
		i.redisClient.Set(
			ctx,
			cacheKeyObjTypeActiveProductsCount,
			cacheKeyActiveProductsCount,
			count,
			5*60, // 5 minutes
		),
	)
}
```
