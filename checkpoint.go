package ellipxobj

type Checkpoint struct {
	Pair       PairName `json:"pair"`
	Epoch      uint64   `json:"epoch"`
	PrevEpoch  uint64   `json:"prev"`
	PrevHash   []byte   `json:"prev_hash"`
	OrderSum   []byte   `json:"in_sum"`
	OrderCount uint64   `json:"in_cnt"`
	Bids       []*Order `json:"bids"`
	Asks       []*Order `json:"asks"`
}
