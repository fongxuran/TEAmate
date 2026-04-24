---
description: 'Instructions for writing Go code in the model layer following best practices.'
applyTo: 'api/internal/model/*.go,!gqlgenerated.go,!**/*_gen.go,!**/mock_*.go'
---

# Model

Contains code that defines the data structures used across the application.
Models represent the core business entities and are used for data transfer between different layers of the application (e.g., handler, controller, repository).
Models should be designed to be independent of any specific storage or transport protocols, focusing solely on the business domain.

- Refer to the [base.instructions.md](go-base.instructions.md) for general Go coding standards.
- Refer to the [structure.instructions.md](go-structure.instructions.md) -> `api/internal/model/` section for folder structure guidelines.

## General Instructions

- Use **singular nouns** for the file and model name (e.g., `product.go` for a `Product` model)
- Define models as Go structs with clear and descriptive field names and appropriate data types
- **Shouldn't use JSON tags in model structs**; JSON tags should be defined in handler or gateway layer structs for external request/response payloads
- Create related constants, types, and helper functions in the same file as the model for better organization
- Implement methods on the model struct for common operations (e.g., validation, formatting)
- Ensure models are well-documented with comments explaining their purpose and usage

## Example

```go
package model

import (
    "time"
)

// ProductStatus represents the status of the product
type ProductStatus string

const (
	// ProductStatusActive means the product is active
	ProductStatusActive ProductStatus = "ACTIVE"
	// ProductStatusInactive means the product is inactive
	ProductStatusInactive ProductStatus = "INACTIVE"
	// ProductStatusDeleted means the product is deleted. This is for archival only
	ProductStatusDeleted ProductStatus = "DELETED"
)

// AllProductStatus is a list of all ProductStatus
var AllProductStatus = []ProductStatus{ProductStatusActive, ProductStatusInactive, ProductStatusDeleted}

// String converts to string value
func (p ProductStatus) String() string {
	return string(p)
}

// IsValid checks if plan status is valid
func (p ProductStatus) IsValid() bool {
	return p == ProductStatusActive || p == ProductStatusInactive || p == ProductStatusDeleted
}

// Product represents the product to be sold
type Product struct {
	ID          int
	ExternalID  string
	Name        string
	Description string
	Status      ProductStatus
	Price       int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
```
