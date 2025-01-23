package core

import (
	"context"
	"sort"

	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

type onpremMappingService struct {
	hostMapFile string

	dvhAPI          service.DatavarehusAPI
	cloudStorageAPI service.CloudStorageAPI
}

type Host struct {
	Description string   `yaml:"description"`
	IPs         []string `yaml:"ips"`
	Port        string   `yaml:"port"`
}

var _ service.OnpremMappingService = &onpremMappingService{}

func (s *onpremMappingService) GetClassifiedHosts(ctx context.Context) (*service.ClassifiedHosts, error) {
	const op errs.Op = "onpremMappingService.GetClassifiedHosts"

	tnsHosts, err := s.dvhAPI.GetTNSNames(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	hostMap := map[string]Host{}
	err = s.cloudStorageAPI.GetObjectAndUnmarshalYAML(ctx, s.hostMapFile, &hostMap)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return s.sortClassifiedHosts(hostMap, tnsHosts), nil
}

func (s *onpremMappingService) sortClassifiedHosts(hostMap map[string]Host, dvhTNSHosts []service.TNSName) *service.ClassifiedHosts {
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

	sort.Slice(classifiedHosts.DVHHosts, func(i, j int) bool {
		return classifiedHosts.DVHHosts[i].TNSName < classifiedHosts.DVHHosts[j].TNSName
	})
	sort.Slice(classifiedHosts.OracleHosts, func(i, j int) bool {
		return classifiedHosts.OracleHosts[i].Host < classifiedHosts.OracleHosts[j].Host
	})
	sort.Slice(classifiedHosts.PostgresHosts, func(i, j int) bool {
		return classifiedHosts.PostgresHosts[i].Host < classifiedHosts.PostgresHosts[j].Host
	})
	sort.Slice(classifiedHosts.InformaticaHosts, func(i, j int) bool {
		return classifiedHosts.InformaticaHosts[i].Host < classifiedHosts.InformaticaHosts[j].Host
	})
	sort.Slice(classifiedHosts.UnclassifiedHosts, func(i, j int) bool {
		return classifiedHosts.UnclassifiedHosts[i].Host < classifiedHosts.UnclassifiedHosts[j].Host
	})

	return classifiedHosts
}

func NewOnpremMappingService(ctx context.Context, hostmapFile string, cloudStorageAPI service.CloudStorageAPI, dvhAPI service.DatavarehusAPI) *onpremMappingService {
	return &onpremMappingService{
		hostMapFile:     hostmapFile,
		dvhAPI:          dvhAPI,
		cloudStorageAPI: cloudStorageAPI,
	}
}
