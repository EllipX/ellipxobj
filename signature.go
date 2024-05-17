package ellipxobj

import (
	"bytes"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/KarpelesLab/jwt"
)

type SignedObject struct {
	Type       string       `json:"typ"`    // order | trade
	Object     string       `json:"object"` // base64url encoded json
	Signatures []*Signature `json:"sigs"`
}

type SignatureHeader struct {
	Type   string `json:"typ"` // order | trade
	Issuer string `json:"iss"` // name of signature issuer (if broker, broker id)
	Alg    string `json:"alg"` // jwt-style algorithm name
}

type Signature struct {
	Header    string `json:"header"` // base64url encoded header
	Signature string `json:"sig"`    // base64url encoded signature
}

func (s *SignedObject) Sign(rand io.Reader, issuer string, k crypto.Signer) error {
	algo, err := jwt.GetAlgoForSigner(k)
	if err != nil {
		return err
	}

	header := &SignatureHeader{
		Type:   s.Type,
		Issuer: issuer,
		Alg:    algo.String(),
	}

	headerJson, err := json.Marshal(header)
	if err != nil {
		return err
	}

	sig := &Signature{
		Header: base64.RawURLEncoding.EncodeToString(headerJson),
	}

	// this is very similar to jwt
	signString := &bytes.Buffer{}
	signString.WriteString(sig.Header)
	signString.WriteByte('.')
	signString.WriteString(s.Object)

	binsig, err := algo.Sign(rand, signString.Bytes(), k)
	if err != nil {
		return err
	}

	sig.Signature = base64.RawURLEncoding.EncodeToString(binsig)
	s.Signatures = append(s.Signatures, sig)

	return nil
}
