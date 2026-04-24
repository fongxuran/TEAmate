---
description: 'Instructions for implementing Go controller-layer code following clean architecture and internal best practices.'
applyTo: 'api/internal/controller/*.go,!gqlgenerated.go,!**/*_gen.go,!**/mock_*.go'
---

# Controller Layer Guidelines

This instruction governs **business logic orchestration** inside the `api/internal/controller` package.  
Controllers are responsible for implementing domain logic, validation and coordination between repositories and handlers.

Refer also to:
- [Base Go standards](go-base.instructions.md)
- [Go API structure guidelines](go-structure.instructions.md)
- [Go Style Guide](go-style.instructions.md)

## Purpose

- Encapsulate business logic and domain rules
- Coordinate calls to repository, gateway, and handlers
- Validate data from datasource before processing
- Keep code composable, mockable, and testable
- Decouple API/handler layer from data and external dependencies

## Design Principles

✅ **Do**
- Organize by **domain** (e.g., `order`, `asset`, `incident`)
- Use **plural package names** (e.g., `orders`, not `order`)
- Define interfaces that expose domain operations
- Accept `context.Context` as the first parameter in every method
- Call repository methods for data access; do not embed SQLBoiler or DB logic
- Apply business logic, validation and orchestration in the controller
- Return domain models (`model.*`) or custom structs, not repository or ORM types
- Keep controller methods side-effect free except for intended business actions
- Place the controller interface and constructor in a single `new.go` file under the domain package

🚫 **Don’t**
- Don’t leak repository or ORM types outside the controller
- Don’t embed DB or transport logic here
- Don’t handle HTTP-specific errors or logging logic here
- Don’t duplicate business rules in handler or repository layers

## Package Skeleton

- Place constructor and interface in a single `new.go` file under the domain package.
    - Keep a single implementation type (e.g., `impl`) and inject dependencies there.

```go
package products

import (
	"context"

	"code.in.spdigital.sg/sp-digital/pencilcase/api/internal/model"
	"code.in.spdigital.sg/sp-digital/pencilcase/api/internal/repository"
)

// Controller represents the specification of this pkg
type Controller interface {
	// List gets a list of products from DB
	List(ctx context.Context) ([]model.Product, error)
	// Get gets a single product from DB
	Get(ctx context.Context, extID string) (model.Product, error)
	// Create creates a product in DB
	Create(ctx context.Context, inp CreateInput) (model.Product, error)
	// Delete deletes a product from DB by marking the status as deleted
	Delete(ctx context.Context, extID string) error
	// GetActiveCount gets the count of active products
	GetActiveCount(ctx context.Context) (int64, error)
}

// New initializes a new Controller instance and returns it
func New(repo repository.Registry) Controller {
	return impl{repo: repo}
}

type impl struct {
	repo repository.Registry
}
```

## Implementation Rules

- Each controller package should have a single `impl` struct in the `new.go` file.
- Define a `Controller` interface in the same `new.go` file, listing all domain operations.
- Controller interface and constructor must be defined in a single `new.go` file within the domain package. Do not create a separate interface.go.
- Group related functions logically:
    - One file per core function (e.g., `get.go`)
    - Or one file per small domain (e.g., `order.go` for List/Get).
- Register controllers in `registry.go` if required by the project pattern.

### Parameter and Return Types

- Use domain model structs for input/output.
- For queries, use filter structs from the repository layer.
- Add validation and business rules before calling repository methods.
- Propagate repository errors up, wrapping with context if needed.

### Business Logic & Validation

- Validate input agaist with data source and enforce domain invariants
- Apply business rules before/after repository calls
- Return errors for invalid input or failed invariants
- Do not duplicate validation in handler or repository layers

### Error Handling

- Propagate errors from repository layer
- Wrap errors with context using `fmt.Errorf` or `pkgerrors.WithStack` if needed

<!-- ### Testing & Performance

- Write table-driven tests for each controller function
- Mock repository dependencies for unit tests
- Ensure each test runs independently -->

## Example Implementation

```go
package products

import (
	"context"

	"code.in.spdigital.sg/sp-digital/pencilcase/api/internal/model"
	"code.in.spdigital.sg/sp-digital/pencilcase/api/internal/pkg/featuretoggle"
)

// CreateInput holds input params for creating the product
type CreateInput struct {
	Name  string
	Desc  string
	Price int64
}

// Create creates the product
func (i impl) Create(ctx context.Context, inp CreateInput) (model.Product, error) {
	product := model.Product{
		Name:        inp.Name,
		Description: inp.Desc,
		Status:      model.ProductStatusInactive,
		Price:       inp.Price,
	}

	if featuretoggle.IsProductV2Enabled() {
		product.Status = model.ProductStatusActive
	}

	product, err := i.repo.Inventory().CreateProduct(ctx, product)
	return product, err
}

```
