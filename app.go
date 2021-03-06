package cvcert

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

func (v *vacInfo) VaccinationDate() (time.Time, error) {
	date, err := time.Parse(isoDate, v.DateOfVaccination)
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}

type cwtHeader struct {
	Algorithm     int    `cbor:"1,keyasint"`
	KeyIdentifier []byte `cbor:"4,keyasint"`
}

type signedCWT struct {
	_           struct{} `cbor:",toarray"`
	Protected   []byte
	Unprotected interface{}
	Payload     []byte
	Signature   []byte
}

const daysSinceVaccination = 14
const isoDate = "2006-01-02"

// not used anymore as we support all EU countries
const finnishPK = "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEepKcLfnTZIej9gSNJVmR8sRYMMgztnG9h0ZGWx7D1X1g32V/GtJc55HkoH+vqkbkhKJnvDBJ1JdsbkKKmBJb2Q=="

// parseCbor returns CWT (from which Signature is used) and commonPayload
// which includes issuer and certificate
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

func getKID(header []byte) (string, error) {
	var cheader cwtHeader
	err := cbor.Unmarshal(header, &cheader)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cheader.KeyIdentifier), nil
}

func ReadData() ([]byte, error) {
	log.Println("reading data.txt...")
	return ioutil.ReadFile("data.txt")
}

// openData decodes base64 input and then decompresses zlib data returning it
func openData(data []byte) ([]byte, error) {

	if !bytes.HasPrefix(data, []byte("HC1:")) {
		return nil, errors.New(`invalid data, should start with "HC1:"`)
	}

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

	decompressed := bytes.NewBuffer(make([]byte, 0))
	_, err = io.Copy(decompressed, r)
	if err != nil {
		return nil, err
	}

	return decompressed.Bytes(), nil
}

// verify return nil if verify is success
func verify(data []byte, signature []byte, kid string) error {
	pk, err := getPK(kid)
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

// getPK gets and returns PublicKey from constant.
// TODO: Support multiple pubkeys and read them from file or directly
// from the API
func getPK(kid string) (*ecdsa.PublicKey, error) {
	pk, ok := publicKeys[kid]
	if !ok {
		return nil, fmt.Errorf("no public key found with key id %s", kid)
	}
	fpk, err := base64.StdEncoding.DecodeString(pk)
	if err != nil {
		return nil, err
	}
	key, err := x509.ParsePKIXPublicKey(fpk)
	if err != nil {
		return nil, err
	}
	return key.(*ecdsa.PublicKey), nil
}

// doValidation decodes and parse cbor and validates signature
// It also validates vaccination date and health certificate expiration
// Returns string-interface map which can be feed to javascript. All values
// in the map shall be strings.
func DoValidation(data []byte) map[string]interface{} {

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

	kid, err := getKID(cwt.Protected)
	if err != nil {
		log.Errorf("failed to get KID from header: %s", err)
		return nil
	}

	err = verify(cborData, cwt.Signature, kid)
	if err != nil {
		log.Errorf("signature verification failed: %s", err)
		return nil
	}
	log.Printf("signature verified successfully")

	cert := cp.Cert[1]
	vac := cert.V[0]
	expiry := time.Unix(cp.Expires, 0)

	validStrs := []string{"certificate is trusted"}

	// 1. ensure cert not expired
	if time.Now().After(expiry) {
		validStrs = append(validStrs, "certificate is expired")
	}
	// 2. ensure enough doses
	if vac.DoseNumber < vac.TotalDoses {
		validStrs = append(validStrs, "not enough doses")
	}
	// 3. ensure at least 14 days since last dose
	vacDate, err := vac.VaccinationDate()
	if err != nil {
		validStrs = append(validStrs, "failed to parse vaccination date")
	} else if vacDate.AddDate(0, 0, daysSinceVaccination).After(time.Now()) {
		validStrs = append(validStrs, "not enough days since last dose")
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
