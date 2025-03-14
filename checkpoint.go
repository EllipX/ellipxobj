package ellipxobj

// Checkpoint represents a snapshot of an order book at a specific point in time.
// Checkpoints can be used for verification, recovery, or synchronization
// of order book state between different systems.
//
// Checkpoints form a chain through PrevEpoch and PrevHash fields,
// allowing validation of the integrity of the order book history.
type Checkpoint struct {
	Pair       PairName `json:"pair"`      // The trading pair this checkpoint belongs to
	Epoch      uint64   `json:"epoch"`     // Current checkpoint sequence number
	PrevEpoch  uint64   `json:"prev"`      // Previous checkpoint sequence number (for chain validation)
	PrevHash   []byte   `json:"prev_hash"` // Hash of the previous checkpoint (for integrity verification)
	Point      TimeId   `json:"point"`     // Timestamp of when this checkpoint was created
	OrderSum   []byte   `json:"in_sum"`    // Cryptographic hash representing all orders in the book
	OrderCount uint64   `json:"in_cnt"`    // Total number of orders included in this checkpoint
	Bids       []*Order `json:"bids"`      // Buy orders in the book (sorted by price, highest first)
	Asks       []*Order `json:"asks"`      // Sell orders in the book (sorted by price, lowest first)
}
