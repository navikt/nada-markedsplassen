package service

import "context"

type OnpremMappingService interface {
	GetClassifiedHosts(ctx context.Context) (*ClassifiedHosts, error)
}

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
	DVHHosts          []TNSHost `json:"dvhHosts"`
	OracleHosts       []Host    `json:"oracleHosts"`
	PostgresHosts     []Host    `json:"postgresHosts"`
	InformaticaHosts  []Host    `json:"informaticaHosts"`
	UnclassifiedHosts []Host    `json:"unclassifiedHosts"`
}
