package jenga

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
)

func SignSha256DataWithPrivateKey(data, privateKeyPath string) (string, error) {

	signer, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return "", err
	}
	signed, err := signer.SignSHA256([]byte(data))
	if err != nil {
		return "", err
	}
	sig := base64.StdEncoding.EncodeToString(signed)
	return sig, nil
}

func SignDataWithPrivateKey(data, privateKeyPath string) (string, error) {

	signer, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return "", err

	}
	signed, err := signer.Sign([]byte(data))
	sig := base64.StdEncoding.EncodeToString(signed)
	return sig, nil
}

// A Signer is can create signatures that verify against a public key.
type Signer interface {
	// Sign returns raw signature for the given data. This method
	// will apply the hash specified for the key type to the data.
	Sign(data []byte) ([]byte, error)
	SignSHA256(data []byte) ([]byte, error)
}

type rsaPrivateKey struct {
	*rsa.PrivateKey
}

func loadPrivateKey(path string) (Signer, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParsePrivateKey(data)
}

// parsePublicKey parses a PEM encoded private key.
func ParsePrivateKey(pemBytes []byte) (Signer, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("crypto: no key found")
	}

	var rawkey interface{}
	switch block.Type {
	case "RSA PRIVATE KEY":
		rsa, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rawkey = rsa
	default:
		return nil, fmt.Errorf("crypto: unsupported private key type %q", block.Type)
	}
	return newSignerFromKey(rawkey)
}
func newSignerFromKey(k interface{}) (Signer, error) {
	var sshKey Signer
	switch t := k.(type) {
	case *rsa.PrivateKey:
		sshKey = &rsaPrivateKey{t}
	default:
		return nil, fmt.Errorf("crypto: unsupported key type %T", k)
	}
	return sshKey, nil
}

// Signs directly the data.
func (r *rsaPrivateKey) Sign(data []byte) ([]byte, error) {
	return rsa.SignPKCS1v15(nil, r.PrivateKey, 0, data)
}

// Sign signs data with rsa-sha256
func (r *rsaPrivateKey) SignSHA256(data []byte) ([]byte, error) {
	h := sha256.New()
	h.Write(data)
	d := h.Sum(nil)
	return rsa.SignPKCS1v15(rand.Reader, r.PrivateKey, crypto.SHA256, d)
}
