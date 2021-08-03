package main

import (
	"bytes"
	"compress/zlib"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dasio/base45"
	"github.com/fxamacker/cbor/v2"
	"github.com/veraison/go-cose"
)

type commonPayload struct {
	Issuer   string        `cbor:"1,keyasint"`
	IssuedAt int64         `cbor:"6,keyasint"`
	Expires  int64         `cbor:"4,keyasint"`
	Cert     map[int]hcert `cbor:"-260,keyasint"`
}

type hcert struct {
	DateOfBirth string            `json:"dob"`
	Name        map[string]string `json:"nam"`
	V           []vacInfo         `json:"v"`
	Version     string            `json:"ver"`
}
type vacInfo struct {
	TargetedDisease   string `json:"tg"`
	Vaccine           string `json:"vp"`
	Product           string `json:"mp"`
	Manufacturer      string `json:"ma"`
	DoseNumber        int    `json:"dn"`
	TotalDoses        int    `json:"sd"`
	DateOfVaccination string `json:"dt"`
	Country           string `json:"co"`
	Issuer            string `json:"is"`
	Identifier        string `json:"ci"`
}

type signedCWT struct {
	_           struct{} `cbor:",toarray"`
	Protected   []byte
	Unprotected interface{}
	Payload     []byte
	Signature   []byte
}

const finnishPK = "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEepKcLfnTZIej9gSNJVmR8sRYMMgztnG9h0ZGWx7D1X1g32V/GtJc55HkoH+vqkbkhKJnvDBJ1JdsbkKKmBJb2Q"

func parseCbor(data []byte) (*signedCWT, *commonPayload, error) {

	var cwt signedCWT
	err := cbor.Unmarshal(data, &cwt)
	if err != nil {
		return nil, nil, err
	}

	var cp commonPayload
	err = cbor.Unmarshal(cwt.Payload, &cp)
	if err != nil {
		return nil, nil, err
	}

	log.Debugf("issuer    : %s", cp.Issuer)
	log.Debugf("issued at : %s", time.Unix(cp.IssuedAt, 0))
	log.Debugf("expires at: %s", time.Unix(cp.Expires, 0))

	if cert, ok := cp.Cert[1]; !ok {
		return nil, nil, errors.New("no hcert found")
	} else if len(cert.V) != 1 {
		return nil, nil, fmt.Errorf("invalid number of certs %d", len(cert.V))
	}

	return &cwt, &cp, nil
}
func readData() ([]byte, error) {
	log.Println("reading data.txt...")
	return ioutil.ReadFile("data.txt")
}
func openData(data []byte) ([]byte, error) {
	// skip prefix "HC1:"
	data = data[4:]

	var deco = make([]byte, base45.DecodedLen(len(data)))
	_, err := base45.Decode(deco, data)
	if err != nil {
		return nil, err
	}

	b := bytes.NewReader(deco)
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}

	uncompressed := bytes.NewBuffer(make([]byte, 0))
	_, err = io.Copy(uncompressed, r)
	if err != nil {
		return nil, err
	}

	return uncompressed.Bytes(), nil
}

// verify return nil if verify is success
func verify(data []byte, signature []byte) error {
	pk, err := getPK()
	if err != nil {
		return err
	}
	verifier := cose.Verifier{
		PublicKey: pk,
		Alg:       cose.ES256,
	}

	sigMsg := cose.NewSign1Message()
	err = sigMsg.UnmarshalCBOR(data)
	if err != nil {
		return err
	}

	err = sigMsg.Verify([]byte{}, verifier)
	if err != nil {
		return err
	}

	return nil
}

func getPK() (*ecdsa.PublicKey, error) {
	fpk, err := base64.RawStdEncoding.DecodeString(finnishPK)
	if err != nil {
		return nil, err
	}
	key, err := x509.ParsePKIXPublicKey(fpk)
	if err != nil {
		return nil, err
	}
	return key.(*ecdsa.PublicKey), nil
}

func doValidation(data []byte) map[string]interface{} {

	start := time.Now()
	defer func() {
		log.Printf("parsing and validation process took %s", time.Since(start))
	}()
	cborData, err := openData(data)
	if err != nil {
		log.Errorf("failed to read data: %s", err)
		return nil
	}

	cwt, cp, err := parseCbor(cborData)
	if err != nil {
		log.Errorf("failed to parse cbor and get signature: %s", err)
		return nil
	}

	err = verify(cborData, cwt.Signature)
	if err != nil {
		log.Errorf("signature verification failed: %s", err)
		return nil
	}
	log.Printf("signature verified successfully")

	cert := cp.Cert[1]
	vac := cert.V[0]
	expiry := time.Unix(cp.Expires, 0)

	validStrs := []string{"certificate is trusted"}

	if time.Now().After(expiry) {
		validStrs = append(validStrs, "certificate is expired")
	}
	if vac.DoseNumber < vac.TotalDoses {
		validStrs = append(validStrs, "not enough doses")
	}
	return map[string]interface{}{
		"name":       cp.Cert[1].Name["gn"] + " " + cp.Cert[1].Name["fn"],
		"doses":      fmt.Sprintf("%d / %d", vac.DoseNumber, vac.TotalDoses),
		"issuer":     vac.Issuer,
		"expires":    expiry.String(),
		"birth_date": cert.DateOfBirth,
		"country":    vac.Country,
		"valid":      strings.Join(validStrs, ", "),
	}

}

func init() {
	log.SetLevel(log.DebugLevel)
}
