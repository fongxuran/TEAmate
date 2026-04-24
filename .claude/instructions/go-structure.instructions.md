---
description: 'Instructions for Go project structure with Bobcat template and best practices.'
applyTo: '**/*.go,!gqlgenerated.go,!**/*_gen.go,!**/mock_*.go'
---

# Go Project Structure Instructions

Follow Bobcat project template which inspired by Uber Clean Architecture when writing Go code.
These instructions are based on:
- [Bobcat - api](https://code.in.spdigital.sg/sp-digital/bobcat/api)
- [Uber Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

## Bobcat Project Structure
All golang code should typically exist in ./api or ./api-go

- `api` - Holds backend code in Golang.
    - `cmd` - Contains executable binary root packages.
        - `banner` - Sample binary to print banner in console.
        - `serverd` - Main API server binary.
            - `router.go` - Single file listing all HTTP ingress routes.
        - `job` - Batch jobs binary.
        - `single/multi partition consumer` - Sample kafka consumer binary.
        - `common` - Shared code across all binaries.
        - `datagen` - Sample load producer binary.
    - `data` - Contains SQL files for direct DB manipulation.
        - `migrations` - Holds all the migration files used by [migrate](https://github.com/golang-migrate/migrate).
        - `seed` - Holds all seeding data.
    - `internal`
        - `model` - [Entities](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html). Holds the business logic entities along with business rules. No API protocol/storage code should be in here. Does not call anyone.
        - `handler` - [Presenters & part of the controller](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html). Contains API protocol related code for ingress calls. No business logic except for simple validation. Only calls controller.
            - `gql` - Contains code for ingress GraphQL calls.
                - `authenticated` - Contains code for authenticated GraphQL calls.
                    - `schema` - Contains GraphQL schema definitions.
                - `public` - Contains code for public GraphQL calls.
                - `m2m` - Contains code for machine-to-machine GraphQL calls.
                - `mod` - Contains input/output structs shareable across the other sub-packages.
                - `dataloader` - Contains code for Graphql batching and caching database calls to avoid N+1 query problem.
            - `grpc` - Contains code for ingress gRPC calls.
            - `kafka` - Contains code for ingress Kafka calls. No need for subpackages here.
            - `rest` - Contains code for ingress REST calls.
                - `system` - Contains code for system endpoints such as health check, metrics, etc.
                - `admin` - Contains code for admin portal REST calls.
                - `authenticated` - Contains code for authenticated REST calls.
                    - `scan` - Example sub-package for scan domain related authenticated REST calls.
                - `public` - Contains code for public REST calls.
                - `m2m` - Contains code for machine-to-machine REST calls.
        - `controller` - [Controller & part of the presenter](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html). Contains business logic. No API protocol/storage code should be in here. Only calls repository or gateway.
            - sub-packages here are domain specific. e.g.:
            - `products` - Example sub-package for product domain related business logic.
            - `orders` - Example sub-package for order domain related business logic.
        - `repository` - [Repository](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html). Contains code to talk to storage layer such as Database, Cache, BLOB store, etc.
            - `orm` - Contains ORM related code such as SQLBoiler generated models and queries.
            - `generator` - Contains code to generate snowflake IDs for DB primary keys. No business logic except for ID generation. Does not call anyone.
            - sub-packages here are domain specific. e.g.:
            - `orders` - Example sub-package for order domain related business logic.
        - `gateway` - [Gateway](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html). Contains API protocol related code for egress calls to other microservices or message bus. No business logic except for outbound validation. Does not call anyone.
            - sub-packages here are destination specific. e.g.:
            - `avaas` - Example sub-package for AVAAS microservice related egress calls.
        - `pkg` - Helper/library code used across all packages, but internal to this repository only.
    - `pkg` - Helper/library code used across packages, but intended to be moved out to common libraries such as athena.
    - `tools` - Helper/library code used across packages, but intended to be moved out to common libraries such as athena.
    - `go.mod/go.sum` - Go code's external dependencies.
    - `vendor` - Holds vendored packages. Ideally we would like to not commit this, but due to a lack of Go dependency caching server in SPD, we will be committing this.
    - `sqlboiler.yaml` - [SQLBoiler](https://github.com/aarondl/sqlboiler) configs.
    - `.mockery.yaml` - [Mockery](https://github.com/vektra/mockery) configs.


### CMD
This portion deals with the entry points of the application, such as the server, background jobs, and other executable components. It typically contains the main function and any related setup or configuration code.

### Internal
This is the core of the application, containing all the business logic, handlers, repositories, models, and utilities. It is organized into subdirectories based on functionality, such as handlers for different types

folders:
- `handler`: Contains the logic for handling incoming requests, whether they are gRPC, GraphQL, REST.
- `repository`: Contains the data access layer, responsible for interacting with the database or other storage systems.
- `utils`: Contains utility functions that can be used across the application.
- `controller`: Contains the business logic that sits between the handlers and repositories.
- `model`: Contains the domain models and data structures used throughout the application.
- `gateway`: Contains code for interacting with external services or APIs. 

### Data
This directory contains database migration files and database-related scripts.

## Technical Requirements
Know the technical requirements for the implementation, always think about the flow of data, eg
New user story requires a new data field to be shown in the UI for substation, this would entail handler requiring export of the new field in the response struct, questions to ask
1. does the handler already export the relevant struct?
2. does the model already have the relevant field, if no add a new field to the model which is used internally
2. does the field exist on the ORM and database, if no:
  a. migration file to add the new column to the relevant table
  b. repository requiring the new field to be added to the output struct
4. ask about the data source if current implementation does not have the data, and plan for how to populate the new field with data from the relevant source, user may say to implement extraction only

folder flow: DB -> repository -> controller -> handler -> UI
data structure flow: DB -> DB ORM -> model -> handler export -> UI
*conversion between DB and model types happens in the repository layer, conversion between model and handler types happens in the controller layer

In another example would be to update a form in the UI to include a new field, which would flow in the opposite direction and the input fields will have to be edited to accept the new data.

When implementing a new API, check if API can fit into existing handlers/repo/controller or if a new one is required. eg asking for a new GetSubstationList API with new filters, it should go under existing substation handler/repo/controller, do not create new ones. Unless user is asking for new data requiring new models and DB tables, then new handler/repo/controller may be required.

