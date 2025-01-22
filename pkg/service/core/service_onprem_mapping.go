package core

import (
	"context"

	"github.com/navikt/nada-backend/pkg/cs"
	"github.com/navikt/nada-backend/pkg/datavarehus"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"gopkg.in/yaml.v3"
)

type onpremMappingService struct {
	hostMapFilePath string

	dvhClient *datavarehus.Client
	ops       cs.Operations
}

type Host struct {
	Description string   `yaml:"description"`
	IPs         []string `yaml:"ips"`
	Port        string   `yaml:"port"`
}

var _ service.OnpremMappingService = &onpremMappingService{}

func (s *onpremMappingService) GetClassifiedHosts(ctx context.Context) (*service.ClassifiedHosts, error) {
	const op errs.Op = "onpremMappingService.GetClassifiedHosts"

	tnsHosts, err := s.dvhClient.GetTNSNames(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	obj, err := s.ops.GetObjectWithData(ctx, s.hostMapFilePath)
	if err != nil {
		return nil, errs.E(op, err)
	}

	hostMap := map[string]Host{}
	err = yaml.Unmarshal(obj.Data, &hostMap)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return s.sortClassifiedHosts(hostMap, tnsHosts), nil
}

func (s *onpremMappingService) sortClassifiedHosts(hostMap map[string]Host, dvhTNSHosts []datavarehus.TNSName) *service.ClassifiedHosts {
	classifiedHosts := &service.ClassifiedHosts{
		DVHHosts:          make([]service.TNSHost, 0),
		OracleHosts:       make([]service.Host, 0),
		PostgresHosts:     make([]service.Host, 0),
		InformaticaHosts:  make([]service.Host, 0),
		UnclassifiedHosts: make([]service.Host, 0),
	}

	for host, data := range hostMap {
		switch data.Port {
		case "1521":
			classifiedHosts.OracleHosts = append(classifiedHosts.OracleHosts, service.Host{
				Host:        host,
				Description: data.Description,
			})
		case "5432":
			classifiedHosts.PostgresHosts = append(classifiedHosts.PostgresHosts, service.Host{
				Host:        host,
				Description: data.Description,
			})
		case "6005-6120":
			classifiedHosts.InformaticaHosts = append(classifiedHosts.InformaticaHosts, service.Host{
				Host:        host,
				Description: data.Description,
			})
		default:
			classifiedHosts.UnclassifiedHosts = append(classifiedHosts.UnclassifiedHosts, service.Host{
				Host:        host,
				Description: data.Description,
			})
		}
	}

	for _, tnsHost := range dvhTNSHosts {
		if _, ok := hostMap[tnsHost.Host]; ok {
			classifiedHosts.DVHHosts = append(classifiedHosts.DVHHosts, service.TNSHost{
				Host:        tnsHost.Host,
				Description: tnsHost.Description,
				TNSName:     tnsHost.TnsName,
			})
		}
	}

	return classifiedHosts
}

func NewOnpremMappingService(ctx context.Context, hostMapFile, bucketName string, dvhClient *datavarehus.Client) *onpremMappingService {
	csClient, err := cs.New(ctx, bucketName)
	if err != nil {
		panic(err)
	}

	return &onpremMappingService{
		hostMapFilePath: hostMapFile,
		dvhClient:       dvhClient,
		ops:             csClient,
	}
}
