package kms_test

import (
	"context"
	"github.com/navikt/nada-backend/pkg/kms"
	"github.com/navikt/nada-backend/pkg/kms/emulator"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	ctx := context.Background()

	log := zerolog.New(zerolog.NewConsoleWriter())
	e := emulator.New(log)

	id := &kms.KeyIdentifier{
		Project:  "test-project",
		Location: "europe-north1",
		Keyring:  "test-keyring",
		KeyName:  "test-key",
	}

	e.AddSymmetricKey(id.Project, id.Location, id.Keyring, id.KeyName, []byte("7b483b28d6e67cfd3b9b5813a286c763"))

	url := e.Run()
	c := kms.NewClient("", url, true)
	plaintext := []byte("this is a test message")
	ciphertext, err := c.Encrypt(ctx, id, plaintext)
	require.NoError(t, err)

	decrypted, err := c.Decrypt(ctx, id, ciphertext)
	require.NoError(t, err)

	require.Equal(t, plaintext, decrypted, "decrypted text should match original plaintext")
}
