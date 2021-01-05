package core

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"golang.org/x/crypto/ripemd160"
)

// An address is a hash put through a base58 encoder. That hash is made of three parts. The first byte is the version, and the last 4 bytes are a checksum.
// Everything in between is a sha512, RIPEMD160 of the public key

const (
	// checksumLen is the length of the checksum in bytes. The checksum is appended to the end of a publicKeyHash, and
	// it allows us to verify a publicKeyHash
	checksumLen = 4
	// version of the publicKey
	version = 0x00
)

// Wallet is an instance of a single Wallet
type Wallet struct {
	// PrivateKey is an instance of ecdsa.PrivateKey. This contains both the public and private key
	PrivateKey ecdsa.PrivateKey
	PublicKey []byte
}

// CreateWallet creates a single wallet, consisting of a public and private key.
func CreateWallet() (Wallet, error) {
	var (
		wallet Wallet
		err error
	)

	wallet.PrivateKey, wallet.PublicKey, err = NewKeyPair()
	return  wallet, err
}

// NewKeyPair generates a new ecdsa keypair. The publicKey is actually two []byte appended together.
// In practice, when verifying a signature with the publicKey, split them back up into x and y values
func NewKeyPair() (ecdsa.PrivateKey, []byte, error) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		fmt.Printf("error creating wallet keypair: %v", err)
		return ecdsa.PrivateKey{}, nil, err
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey, err
}

// GetAddress gets a Base58Encoded address. This address is the user's Blemflarck address.
func (w Wallet) GetAddress() ([]byte, error) {
	// create a public key hash
	hashPubKey, err := HashPublicKey(w.PublicKey)
	if err != nil {
		return nil, err
	}
	// add the version to the beginning of that hash
	versionPayload := append([]byte{version}, hashPubKey...)
	// create a checksum with that hash+version
	checksum := CreateChecksum(versionPayload)
	// the full payload is version+hash+checksum
	fullPayload := append(versionPayload, checksum...)
	// an address is a base58 version+hash+checksum
	return Base58Encode(fullPayload), nil
}

// HashPublicKey hashes a public key. First it runs a sha512 hashing on the key, then sends that output through a RIPEMD160 Hasher.
func HashPublicKey(pubKey []byte) ([]byte, error) {
	hash := sha512.Sum512(pubKey)
	RIPEMDHasher := ripemd160.New()
	_, err := RIPEMDHasher.Write(hash[:])
	if err != nil {
		return nil, err
	}

	return RIPEMDHasher.Sum(nil), nil
}

// CreateChecksum creates a checksum for a version+hash(of pub key) combo.
// It simply double hashes the payload, and returns the last 4 bytes.
func CreateChecksum(payload []byte) []byte {
	hashA := sha512.Sum512(payload)
	hashB := sha512.Sum512(hashA[:])
	return hashB[:checksumLen]
}

// CheckValidAddress checks if a wallet address is indeed a valid address.
// It does so by reversing the process used to create the address, and then checks the checksums against each other.
// First it base58 decodes the address. Then separates the version, pubKeyHash, and checksum.
// Then it checks if the checksum of the version+hash matches the address's checksum
func CheckValidAddress(address []byte) bool {
	decoded := Base58Decode(address)
	// checksum is the last 4 bytes of the decoded address
	checksum := decoded[len(decoded)-checksumLen:]
	// Create a checksum based off of the decodedAddress minus the checksum. That value should be equal to the checksum on the decodedAddress.
	targetChecksum := CreateChecksum(decoded[:len(decoded)-checksumLen])
	return bytes.Compare(checksum, targetChecksum) == 0
}