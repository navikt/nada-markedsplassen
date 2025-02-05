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
	Type        string   `yaml:"type"`
	Parent      string   `yaml:"parent"`
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

func (s *onpremMappingService) removeVIPNodes(hostMap map[string]Host) {
	for host, data := range hostMap {
		if data.Parent != "" {
			delete(hostMap, host)
		}
	}
}

func (s *onpremMappingService) sortClassifiedHosts(hostMap map[string]Host, dvhTNSHosts []service.TNSName) *service.ClassifiedHosts {
	s.removeVIPNodes(hostMap)

	classifiedHosts := &service.ClassifiedHosts{
		DVHHosts:          make([]service.TNSHost, 0),
		HTTPHosts:         make([]service.Host, 0),
		OracleHosts:       make([]service.Host, 0),
		PostgresHosts:     make([]service.Host, 0),
		InformaticaHosts:  make([]service.Host, 0),
		UnclassifiedHosts: make([]service.Host, 0),
	}

	for _, tnsHost := range dvhTNSHosts {
		if _, ok := hostMap[tnsHost.Host]; ok {
			classifiedHosts.DVHHosts = append(classifiedHosts.DVHHosts, service.TNSHost{
				Host:        tnsHost.Host,
				Description: tnsHost.Description,
				TNSName:     tnsHost.TnsName,
			})

			delete(hostMap, tnsHost.Host)
		}
	}

	for host, data := range hostMap {
		switch data.Type {
		case "oracle":
			classifiedHosts.OracleHosts = append(classifiedHosts.OracleHosts, service.Host{
				Host:        host,
				Description: data.Description,
			})
		case "postgres":
			classifiedHosts.PostgresHosts = append(classifiedHosts.PostgresHosts, service.Host{
				Host:        host,
				Description: data.Description,
			})
		case "informatica":
			classifiedHosts.InformaticaHosts = append(classifiedHosts.InformaticaHosts, service.Host{
				Host:        host,
				Description: data.Description,
			})
		case "http":
			classifiedHosts.HTTPHosts = append(classifiedHosts.HTTPHosts, service.Host{
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

	sort.Slice(classifiedHosts.DVHHosts, func(i, j int) bool {
		return classifiedHosts.DVHHosts[i].TNSName < classifiedHosts.DVHHosts[j].TNSName
	})
	sort.Slice(classifiedHosts.OracleHosts, func(i, j int) bool {
		return classifiedHosts.OracleHosts[i].Host < classifiedHosts.OracleHosts[j].Host
	})
	sort.Slice(classifiedHosts.HTTPHosts, func(i, j int) bool {
		return classifiedHosts.HTTPHosts[i].Host < classifiedHosts.HTTPHosts[j].Host
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

func NewOnpremMappingService(hostmapFile string, cloudStorageAPI service.CloudStorageAPI, dvhAPI service.DatavarehusAPI) *onpremMappingService {
	return &onpremMappingService{
		hostMapFile:     hostmapFile,
		dvhAPI:          dvhAPI,
		cloudStorageAPI: cloudStorageAPI,
	}
}
