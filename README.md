[![GoDoc](https://godoc.org/github.com/EllipX/ellipxobj?status.svg)](https://godoc.org/github.com/EllipX/ellipxobj)

# EllipX Objects

A Go package providing core data types and structures for EllipX cryptocurrency exchange, with a focus on precision, correctness, and performance.

## Overview

The `ellipxobj` package implements essential objects for the EllipX cryptocurrency exchange with precise decimal handling, order processing, and time-ordered event tracking. It's designed to be used across various components of the EllipX platform to ensure consistent data structures and behavior.

## Key Components

### Amount

Fixed-point decimal implementation with arbitrary precision for cryptocurrency calculations, avoiding floating-point errors. Features include:

- Mathematical operations (add, subtract, multiply, divide)
- Configurable exponent handling
- Precise comparison operations
- Serialization to/from various formats (JSON, strings)

### Order

Represents cryptocurrency trading orders with comprehensive parameter support:

- Buy/sell (bid/ask) directionality
- Price and quantity specifications with high precision
- Order lifecycle management
- Special execution constraints via flags
- Status tracking

### Pair

Trading pair representation (e.g., `BTC_USD`):

- Standardized format for cryptocurrency pairs
- Utilities for extracting base and quote currencies
- Consistent formatting across the exchange

When a pair is mentioned, it uses a specific order and is typically written with a `_` between the two elements. Unlike standard pairs in some systems, there is no guarantee that elements of a given pair will be 3 characters long.

For buy orders, this means exchanging the first element for the second. Sell orders work in the opposite direction.

### Trade

Records completed transactions between matched orders:

- Captures details about executed exchanges
- Maintains references to the orders involved
- Records precise execution price and quantity

### TimeId

Unique timestamp-based identifier with nanosecond precision for accurately ordering events, particularly useful when multiple events occur simultaneously.

### Checkpoint

Provides snapshot capabilities for order book states at specific points in time for verification, recovery, or synchronization between exchange components.

## Usage

These objects form the foundation for the EllipX cryptocurrency exchange platform and can be used to:

- Interact with EllipX APIs
- Process and validate trade data
- Build integrations with the EllipX exchange
- Develop tools that work with EllipX order and trade data

## Development

```bash
make               # Format code with goimports and build project
make deps          # Get dependencies
make test          # Run all tests with verbose output
```

## License

See the [LICENSE](LICENSE) file for details.