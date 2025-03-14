# Ellipxobj Development Guide

## Commands
```
make               # Format code with goimports and build project
make deps          # Get dependencies
make test          # Run all tests with verbose output
go test -v -run=TestAmount  # Run specific test file
go test -v -run=^TestAmount$/^TestValueConversion$  # Run specific test function
```

## Code Style Guidelines
- **Naming**: Use PascalCase for exported functions/types (e.g., `NewAmount`), camelCase for non-exported
- **Constructors**: Use `NewX` prefix for constructor functions (e.g., `NewAmount`, `NewOrder`)
- **Receivers**: Use pointer receivers for state modification (e.g., `func (a *Amount) SetExp()`), non-pointer for read-only methods
- **Method Chaining**: Return `*self` to enable method chaining pattern (e.g., `a.SetExp(5).SetRaw(42)`)
- **Error Handling**: Define specific error types in errors.go, use clear error checks and return paths
- **Documentation**: Document public APIs with usage examples and parameter/return explanations
- **Tests**: Compare expected vs actual outputs with descriptive error messages
- **JSON Handling**: Implement custom Marshaler/Unmarshaler interfaces for complex types
- **Precision**: Use appropriate exponent handling for financial calculations (see Amount type)
- **Imports**: Group standard library first, then third-party, then internal packages
- **Helper Functions**: Use generic helpers like `must[T any](v T, err error)` for test code

## Project Structure
Go module: github.com/EllipX/ellipxobj - Financial exchange objects (Amount, Order, Pair, Trade, TimeId, etc.)