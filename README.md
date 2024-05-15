[![GoDoc](https://godoc.org/github.com/EllipX/ellipxobj?status.svg)](https://godoc.org/github.com/EllipX/ellipxobj)

# EllipX objects

These structures are shared among various parts of the system in order to ensure consistent structure, and can be used to communicate with APIs.

## Pairs

When a pair is mentioned, it comes in a specific order, and typically written with a `_` between the two elements as unlike standard pairs, there is no guarantee that elements of a given pair will be 3 characters long.

If the order is a buy order, it means exchanging the first element for the second. Sell order works the other way around.

