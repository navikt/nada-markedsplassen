package service

import (
	"context"
	"errors"
)

var ErrDatavarehusInvalidDatabaseUser = errors.New("invalid database user")

type DatavarehusAPI interface {
	GetTNSNames(ctx context.Context) ([]TNSName, error)
	SendJWT(ctx context.Context, keyID, signedJWT string) error
}

type TNSName struct {
	TnsName     string
	Name        string
	Description string
	Host        string
	Port        string
	ServiceName string
}
