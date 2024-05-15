package ellipxobj

type SignedOrder struct {
	Order      string       `json:"order"` // json-encoded
	Signatures []*Signature `json:"sigs"`
}

type Signature struct {
	Issuer    string `json:"iss"` // name of signature issuer (if broker, broker id)
	Signature string `json:"sig"` // base64url encoded signature
}
