package ellipxobj

type SignedObject struct {
	Type       string       `json:"type"`   // order | trade
	Object     string       `json:"object"` // base64 encoded json
	Signatures []*Signature `json:"sigs"`
}

type Signature struct {
	Issuer    string `json:"iss"` // name of signature issuer (if broker, broker id)
	Signature string `json:"sig"` // base64url encoded signature
}
