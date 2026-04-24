---
description: 'Instructions for writing Go code following the Go community style best practices.'
applyTo: '**/*.go,!gqlgenerated.go,!**/*_gen.go,!**/mock_*.go'
---

# Go Style Instructions

Follow idiomatic Go practices and community standards when writing Go code.
These instructions are based on:
- [Effective Go](https://go.dev/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide/tree/master)
- [Google Go Style Guide](https://google.github.io/styleguide/go/index.html)
- SPD GDS standards and conventions

## General Instructions

- **Be Consistent**
- Avoid lines of code that require readers to scroll horizontally, but no hard limit, recommend around 99 characters.
-

## Import

### Import Group Ordering

- There should be two import groups:
    - Standard library imports
    - Everything else
```
// Good:
import (
    "fmt"
    "os"
      
    "code.in.spdigital.sg/sp-digital/bobcat/api/internal/model"
    "go.uber.org/atomic"
    "golang.org/x/sync/errgroup"
      
)
  ```

### Import Aliasing

- Import aliasing must be used if the package name does not match the last element of the import path
```
    // Good:
    import (
        "net/http"

        client "example.com/client-go"
        trace "example.com/trace/v2"
    )
```
- In all other scenarios, import aliases should be avoided unless there is a direct conflict between imports
```
    // Bad:
    import (
        "fmt"
        "os"
        runtimetrace "runtime/trace"

        nettrace "golang.net/x/trace"
    )
```
```
    // Good:
    import (
        "fmt"
        "os"
        "runtime/trace"
    
        nettrace "golang.net/x/trace"
    )
```

## Naming Conventions

### Packages

- Use lowercase, single-word package names such as:
    - time (provides functionality for measuring and displaying time)
    - http (provides HTTP client and server implementations)
- Avoid underscores, hyphens, or mixedCaps
- Choose names that describe what the package provides, not what it contains
- Avoid generic names like `util`, `common`, or `base`
- Package names should be singular, not plural in most cases

#### Package Declaration Rules (CRITICAL):
- **NEVER duplicate `package` declarations** - each Go file must have exactly ONE `package` line
- When editing an existing `.go` file:
    - **PRESERVE** the existing `package` declaration - do not add another one
    - If you need to replace the entire file content, start with the existing package name
- When creating a new `.go` file:
    - **BEFORE writing any code**, check what package name other `.go` files in the same directory use
    - Use the SAME package name as existing files in that directory
    - If it's a new directory, use the directory name as the package name
    - Write **exactly one** `package <name>` line at the very top of the file
- When using file creation or replacement tools:
    - **ALWAYS verify** the target file doesn't already have a `package` declaration before adding one
    - If replacing file content, include only ONE `package` declaration in the new content
    - **NEVER** create files with multiple `package` lines or duplicate declarations
- When declaring package name in the `handler`, `controller`, or `repository` layers:
    - Use plural names (e.g., `assets`, `incidents`) as these layers often deal with collections of entities, and singular form is a reserved word or keyword.
- When declaring package name in the `model` layer:
    - Use singular names (e.g., `asset`, `incident`) as these layers typically represent individual entities.


### Variables and Functions

- Use mixedCaps or MixedCaps (camelCase) rather than underscores
- Keep names short but descriptive
- Use single-letter variables only for very short scopes (like loop indices within 7 lines)
- Exported names start with a capital letter
- Unexported names start with a lowercase letter
- Use uppercase acronyms (e.g., `URL`, `HTTP`, `ID`)
- Group related variables together using var blocks for better organization and readability.

### Function and Method Names

- Avoid repetition, consider the context in which the name will be read to avoid redundancy, e.g.:
```
    // Bad:
    package yamlconfig
    
    func ParseYAMLConfig(input string) (*Config, error)
```
```
    // Good:
    package yamlconfig
    
    func Parse(input string) (*Config, error)
```
- For methods, do not repeat the name of the method receiver
```
    // Bad:
    func (c *Config) WriteConfigTo(w io.Writer) (int64, error)
```
```
    // Good:
    func (c *Config) WriteTo(w io.Writer) (int64, error)
```
- Do not repeat the names of variables passed as parameters
```
    // Bad:
    func OverrideFirstWithSecond(dest, source *Config) error
```
```
    // Good:
    func Override(dest, source *Config) error
```
- Do not repeat the names and types of the return values
```
    // Bad:
    func TransformToJSON(input *Config) *jsonconfig.Config
```
```
    // Good:
    func Transform(input *Config) *jsonconfig.Config
```

### Interfaces

- Name interfaces with -er suffix when possible (e.g., `Reader`, `Writer`, `Formatter`)
- Single-method interfaces should be named after the method (e.g., `Read` → `Reader`)
- Keep interfaces small and focused

### Constants

- Use MixedCaps for exported constants
- Use mixedCaps for unexported constants
- Group related constants using `const` blocks
- Consider using typed constants for better type safety (e.g., prefer `const StatusActive Status = "ACTIVE"` not `const StatusActive = "ACTIVE"`)
- Use uppercase with underscores for enum-like constant values (e.g., `PENDING_APPROVAL`, `APPROVED`)


## Commentary

- Strive for self-documenting code; prefer clear variable names, function names, and code structure over comments
- Write comments only when necessary to explain complex logic, business rules, or non-obvious behavior
- Write comments in complete sentences in English by default
- Translate comments to other languages only upon specific user request
- Start sentences with the name of the thing being described
- Package comments should start with "Package [name]"
- Use line comments (`//`) for most comments
- Use block comments (`/* */`) sparingly, mainly for package documentation
- Document why, not what, unless the what is complex
- Avoid emoji in comments and code

### Formatting

- Always use `gofmt` to format code
- Use `goimports` to manage imports automatically
- Keep line length reasonable (no hard limit, but consider readability)
- Add blank lines to separate logical groups of code
