package service

import "context"

type KMSAPI interface {
	Encrypt(ctx context.Context, id *KeyIdentifier, plaintext []byte) ([]byte, error)
	Decrypt(ctx context.Context, id *KeyIdentifier, ciphertext []byte) ([]byte, error)
}

type KeyIdentifier struct {
	Project  string
	Location string
	Keyring  string
	KeyName  string
}
