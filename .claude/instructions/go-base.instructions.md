---
description: 'Instructions for writing Go code following idiomatic Go practices and community standards'
applyTo: '**/*.go,**/go.mod,!gqlgenerated.go,!**/*_gen.go,!**/mock_*.go'
---

# Go Development Baseline Instructions

Follow idiomatic Go practices and community standards when writing Go code.
These instructions are based on:
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- SPD Engineering Golang best practices

## General Instructions

- Write simple, clear, and idiomatic Go code
- Favor clarity and simplicity over cleverness
- Keep the happy path left-aligned (minimize indentation)
- Prefer early return over if-else chains; use `if condition { return }` pattern to avoid else blocks
- Make the zero value useful
- Write self-documenting code with clear, descriptive names
- Document exported types, functions, methods, and packages
- Leverage the Go standard library when functionality exists instead of reinventing the wheel (e.g., use `strings.Builder` for string concatenation, `filepath.Join` for path construction)

## Type Safety and Language Features

### Type Definitions

- Define types to add meaning and type safety, e.g.:
    - `type Status string` with constants for different statuses
- Use struct tags for JSON, XML, database mappings
- Prefer explicit type conversions over implicit conversions, e.g.:
    - `int64(x)` instead of relying on automatic conversion
- Use type assertions carefully and check the second return value
- Prefer generics over unconstrained types; when an unconstrained type is truly needed, use the predeclared alias `any` instead of `interface{}` (Go 1.18+), e.g.:
    - `func Process[T any](input T) { ... }`

### Pointers vs Values

- Prefer value T. When we are passing an argument to a function/method or returning a value, 90% of the time we use value T.
- Only use *T when
    - Receiver’s value required to be changed (usually marshall or decoder use pointer by nature of it’s behaviour)
    - Be consistent within a type's method set, e.g., 3rd parties libraries required to use pointers like `sqlboiler`
    - The receiver object size is big. Justified by profiling
- Avoid using pointers to basic types (e.g., `*int`, `*string`) unless necessary for distinguishing between `not set (nil)` and `set to zero value`

### Interfaces and Composition

- Accept interfaces, return concrete types, e.g., `func Process(r io.Reader) []byte { ... }`
- Keep interfaces small (1-3 methods is ideal)
- Use embedding for composition, e.g.:
  ```
    type Logger struct{}
    
    func (Logger) Info(msg string)  { fmt.Println("INFO:", msg) }
    func (Logger) Error(msg string) { fmt.Println("ERROR:", msg) }
    
    type Service struct {
        Logger // embedded (no field name)
        Name   string
    }
    
    func (s Service) DoWork() {
        s.Info("starting " + s.Name) // promoted method from Logger
    }
  ```
- Define interfaces close to where they're used, not where they're implemented
- Don't export interfaces unless necessary

## Concurrency

### Goroutines

- Be cautious about creating goroutines in libraries; prefer letting the caller control concurrency
- If you must create goroutines in libraries, provide clear documentation and cleanup mechanisms
- Always know how a goroutine will exit
- Use `sync.WaitGroup` or channels to wait for goroutines
- Avoid goroutine leaks by ensuring cleanup
- When launching goroutines within a loop, always pass loop variables as parameters to the goroutine’s function.
  This ensures each goroutine receives its own copy of the variable, avoiding unintended sharing of the loop variable due to closure semantics.

### Channels

- Use channels to communicate between goroutines
- Don't communicate by sharing memory; share memory by communicating
- Close channels from the sender side, not the receiver
- Use buffered channels when you know the capacity
- Use `select` for non-blocking operations

### Synchronization

- Use `sync.Mutex` for protecting shared state
- Keep critical sections small
- Use `sync.RWMutex` when you have many readers
- Choose between channels and mutexes based on the use case: use channels for communication, mutexes for protecting state
- Use `sync.Once` for one-time initialization
- WaitGroup usage by Go version:
    - If `go >= 1.25` in `go.mod`, use the new `WaitGroup.Go` method ([documentation](https://pkg.go.dev/sync#WaitGroup)):
```
    var wg sync.WaitGroup
    wg.Go(task1)
    wg.Go(task2)
    wg.Wait()
```
- If `go < 1.25`, use the classic `Add`/`Done` pattern

## Error Handling Patterns

### Error Handling

- Check errors immediately after the function call
- Don't ignore errors using `_` unless you have a good reason (document why)
- Wrap errors with context using `fmt.Errorf` with `%w` verb
- Create custom error types when you need to check for specific errors
- Place error returns as the last return value
- Name error variables `err`
- Keep error messages lowercase and don't end with punctuation

### Creating Errors

- Use `errors.New` for simple static errors
- Use `fmt.Errorf` for dynamic errors
- Export error variables for sentinel errors, e.g., `var ErrNotFound = errors.New("not found")`
- Use `errors.Is` and `errors.As` for error checking

### Error Propagation

- Add context when propagating errors up the stack, e.g.:
    - `return fmt.Errorf("failed to update due to: %w", err) // propagate with context`
- Don't log and return errors (choose one), e.g.:
```
    // Bad:
    if err != nil {
        monitor.Errorf("failed to update due to: %w", err) // Don't log the same error you're returning
        return fmt.Errorf("failed to update due to: %w", err)
    }
```
- Handle errors at the appropriate level
- Use of `pkgerrors.WithStack()`
    - Adding `WithStack()` allows us to trace back where is the origin of this error
    - Always add this at the bottom layer, not on each caller or its caller’s caller
    - All library or shared errors should be wrapped with stack traces, e.g.:
```
    slice, err := orm.Products(qms...).All(ctx, i.pgConn)
    if err != nil {
        return nil, pkgerrors.WithStack(err)
    }  
```
- Always double check not to introduce double stack traces

## Context

### Context Creation

- Use `context.Context` for cancellation, timeouts, and passing request-scoped values
- Use `context.Background()` for top-level contexts (e.g., main, init)
- Always `defer cancel()` when you create it

### Context Values

- Only for request-scoped data that’s needed across APIs (e.g., request ID)
- Don’t stuff configs, big structs, or optional params into context
- Define unexported, distinct key types to avoid collisions

### Context Propagation

- Always accept `context.Context` as the first parameter in functions, avoid storing it in structs
- Any function that does I/O, blocks, or might take noticeable time should accept a `context.Context`
- Don't use `context.TODO()` or `nil` context in production paths

## Logs

### Logging Practices

- Be consistent in log format and structure
- Include relevant context (request ID, user ID) in logs, e.g., `monitor.Infof("[Login] user [%s] logged in", userID)`
- Use appropriate log levels (Debug, Info, Error)
- **Avoid logging sensitive information**
- Log errors with context, but avoid logging and returning the same error
- Log the right amount of information; avoid excessive logging or under-logging
- Utilize `monitoring` package from `Athena` library with Graylog integration
    - Use `monitor.Infof` for general logging
    - Use `monitor.Errorf` for error logging that requires Sentry error alert
    - Use `monitor.WithTag` or `monitor.WithTags` to add custom tags for better log filtering and searching in Graylog, e.g.:
```
    monitor := monitoring.FromContext(ctx)
    monitor = monitor.WithTag(ctx, "user_id", userID)
      
    ctx = monitoring.SetInContext(ctx, monitor) // Don't forget to set the new monitor back to the context if you want the tags to be available downstream
    doSomething(ctx)
```

## Performance Optimization

### Memory Management

- Minimize allocations in hot paths
- Reuse objects when possible (consider `sync.Pool`)
- Use value receivers for small structs
- Preallocate slices when size is known
- Avoid unnecessary conversions

### I/O: Readers and Buffers

- Most `io.Reader` streams are consumable once; reading advances state. Do not assume a reader can be re-read without special handling
- If you must read data multiple times, buffer it once and recreate readers on demand:
    - Use `io.ReadAll` (or a limited read) to obtain `[]byte`, then create fresh readers via `bytes.NewReader(buf)` or `bytes.NewBuffer(buf)` for each reuse
    - For strings, use `strings.NewReader(s)`; you can `Seek(0, io.SeekStart)` on `*bytes.Reader` to rewind
- For HTTP requests, do not reuse a consumed `req.Body`. Instead:
    - Keep the original payload as `[]byte` and set `req.Body = io.NopCloser(bytes.NewReader(buf))` before each send
    - Prefer configuring `req.GetBody` so the transport can recreate the body for redirects/retries: `req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(buf)), nil }`
- To duplicate a stream while reading, use `io.TeeReader` (copy to a buffer while passing through) or write to multiple sinks with `io.MultiWriter`
- Reusing buffered readers: call `(*bufio.Reader).Reset(r)` to attach to a new underlying reader; do not expect it to “rewind” unless the source supports seeking
- For large payloads, avoid unbounded buffering; consider streaming, `io.LimitReader`, or on-disk temporary storage to control memory

- Use `io.Pipe` to stream without buffering the whole payload:
    - Write to `*io.PipeWriter` in a separate goroutine while the reader consumes
    - Always close the writer; use `CloseWithError(err)` on failures
    - `io.Pipe` is for streaming, not rewinding or making readers reusable

- **Warning:** When using `io.Pipe` (especially with multipart writers), all writes must be performed in strict, sequential order. Do not write concurrently or out of order—multipart boundaries and chunk order must be preserved. Out-of-order or parallel writes can corrupt the stream and result in errors.

- Streaming multipart/form-data with `io.Pipe`:
    - `pr, pw := io.Pipe()`; `mw := multipart.NewWriter(pw)`; use `pr` as the HTTP request body
    - Set `Content-Type` to `mw.FormDataContentType()`
    - In a goroutine: write all parts to `mw` in the correct order; on error `pw.CloseWithError(err)`; on success `mw.Close()` then `pw.Close()`
    - Do not store request/in-flight form state on a long-lived client; build per call
    - Streamed bodies are not rewindable; for retries/redirects, buffer small payloads or provide `GetBody`

## Security Best Practices

### Input Validation

- Validate all external input
- Use strong typing to prevent invalid states
- Sanitize data before using in SQL queries, Use parameterized queries or ORM to prevent SQL injection
- Be careful with file paths from user input
- Validate and escape data for different contexts (HTML, SQL, shell)

### Cryptography

- Use standard library crypto packages
- Don't implement your own cryptography
- Use crypto/rand for random number generation
- Store passwords using bcrypt, scrypt, or argon2 (consider golang.org/x/crypto for additional options)
- Use TLS for network communication

## Common Pitfalls to Avoid

- Not checking errors
- Ignoring race conditions
- Creating goroutine leaks
- Not using defer for cleanup
- Modifying maps concurrently
- Not understanding nil interfaces vs nil pointers
- Forgetting to close resources (files, connections)
- Using global variables unnecessarily
- Over-using unconstrained types (e.g., `any`); prefer specific types or generic type parameters with constraints. If an unconstrained type is required, use `any` rather than `interface{}`
- Not considering the zero value of types
- **Creating duplicate `package` declarations** - this is a compile error; always check existing files before adding package declarations
