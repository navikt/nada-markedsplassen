package gcp

import (
	"context"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/kms"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.KMSAPI = (*kmsAPI)(nil)

type kmsAPI struct {
	ops kms.Operations
}

func (k *kmsAPI) Encrypt(ctx context.Context, id *service.KeyIdentifier, plaintext []byte) ([]byte, error) {
	const op errs.Op = "kmsAPI.Encrypt"

	ciphertext, err := k.ops.Encrypt(ctx, &kms.KeyIdentifier{
		Project:  id.Project,
		Location: id.Location,
		Keyring:  id.Keyring,
		KeyName:  id.KeyName,
	}, plaintext)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return ciphertext, nil
}

func (k *kmsAPI) Decrypt(ctx context.Context, id *service.KeyIdentifier, ciphertext []byte) ([]byte, error) {
	const op errs.Op = "kmsAPI.Decrypt"

	plaintext, err := k.ops.Decrypt(ctx, &kms.KeyIdentifier{
		Project:  id.Project,
		Location: id.Location,
		Keyring:  id.Keyring,
		KeyName:  id.KeyName,
	}, ciphertext)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return plaintext, nil
}

func NewKMSAPI(ops kms.Operations) service.KMSAPI {
	return &kmsAPI{
		ops: ops,
	}
}
