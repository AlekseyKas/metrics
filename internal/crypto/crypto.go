package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

//decrypt data
func DecryptData(data []byte, privKey string) ([]byte, error) {
	pk, err := os.ReadFile(privKey)
	if err != nil {
		return nil, err
	}
	pem, _ := pem.Decode(pk)

	rsaPriv, err := x509.ParsePKCS8PrivateKey(pem.Bytes)
	if err != nil {
		return nil, err
	}

	if rsaPriv == nil {
		return nil, errors.New("nil private key")
	}
	hash := sha256.New()
	key := rsaPriv.(*rsa.PrivateKey)
	dataLen := len(data)
	step := 384
	var decryptedBytes []byte
	for start := 0; start < dataLen; start += step {
		finish := start + step
		if finish > dataLen {
			finish = dataLen
		}
		decryptedBlock, err := rsa.DecryptOAEP(hash, rand.Reader, key, data[start:finish], nil)
		if err != nil {
			return nil, err
		}
		decryptedBytes = append(decryptedBytes, decryptedBlock...)
	}
	return decryptedBytes, err
}
