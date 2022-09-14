package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"

	"golang.org/x/crypto/ssh"
)

func EncryptData(data []byte, pubKey string) ([]byte, error) {
	pk, err := os.ReadFile(pubKey)
	if err != nil {
		return nil, err
	}
	parsed, _, _, _, err := ssh.ParseAuthorizedKey(pk)
	if err != nil {
		return nil, err
	}

	parsedCryptoKey := parsed.(ssh.CryptoPublicKey)
	pubCrypto := parsedCryptoKey.CryptoPublicKey()

	pub := pubCrypto.(*rsa.PublicKey)
	hash := sha256.New()
	datalen := len(data)
	step := pub.Size() - 2*hash.Size() - 2

	var encryptedBytes []byte
	for start := 0; start < datalen; start += step {
		finish := start + step
		if finish > datalen {
			finish = datalen
		}
		var encryptedBlock []byte
		encryptedBlock, err = rsa.EncryptOAEP(hash, rand.Reader, pub, data[start:finish], nil)
		if err != nil {
			return nil, err
		}
		encryptedBytes = append(encryptedBytes, encryptedBlock...)
	}
	return encryptedBytes, err
}

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
		var decryptedBlock []byte
		decryptedBlock, err = rsa.DecryptOAEP(hash, rand.Reader, key, data[start:finish], nil)
		if err != nil {
			return nil, err
		}
		decryptedBytes = append(decryptedBytes, decryptedBlock...)
	}
	return decryptedBytes, err
}
