package service

import "context"

type TNSHost struct {
	Host        string
	Description string
	TNSName     string
}

type Host struct {
	Description string
	Host        string
}

type ClassifiedHosts struct {
	DVHHosts          []TNSHost
	OracleHosts       []Host
	PostgresHosts     []Host
	InformaticaHosts  []Host
	UnclassifiedHosts []Host
}

type OnpremMappingService interface {
	GetClassifiedHosts(ctx context.Context) (*ClassifiedHosts, error)
}
