package service

import "context"

type OnpremMappingService interface {
	GetClassifiedHosts(ctx context.Context) (*ClassifiedHosts, error)
}

type Host struct {
	Name        string
	Description string
	Host        string
}

type ClassifiedHosts struct {
	Hosts map[OnpremHostType][]*Host `json:"hosts"`
}

type OnpremHostType string

const (
	OnpremHostTypeOracle      OnpremHostType = "oracle"
	OnpremHostTypePostgres    OnpremHostType = "postgres"
	OnpremHostTypeInformatica OnpremHostType = "informatica"
	OnpremHostTypeSFTP        OnpremHostType = "sftp"
	OnpremHostTypeHTTP        OnpremHostType = "http"
	OnpremHostTypeTDV         OnpremHostType = "tdv"
	OnpremHostTypeSMTP        OnpremHostType = "smtp"
	OnpremHostTypeTNS         OnpremHostType = "tns"
	OnpremHostTypeCloudSQL    OnpremHostType = "cloudsql"
)

func ValidOnpremHostType(hostType OnpremHostType) bool {
	switch hostType {
	case OnpremHostTypeOracle:
		fallthrough
	case OnpremHostTypePostgres:
		fallthrough
	case OnpremHostTypeInformatica:
		fallthrough
	case OnpremHostTypeSFTP:
		fallthrough
	case OnpremHostTypeHTTP:
		fallthrough
	case OnpremHostTypeTDV:
		fallthrough
	case OnpremHostTypeSMTP:
		fallthrough
	case OnpremHostTypeTNS:
		fallthrough
	case OnpremHostTypeCloudSQL:
		return true
	}

	return false
}
