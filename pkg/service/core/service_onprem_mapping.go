package core

import (
	"context"
	"sort"

	"github.com/rs/zerolog"

	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

type onpremMappingService struct {
	hostMapFile string

	dvhAPI          service.DatavarehusAPI
	cloudStorageAPI service.CloudStorageAPI
	bucket          string

	log zerolog.Logger
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
	err = s.cloudStorageAPI.GetObjectAndUnmarshalYAML(ctx, s.bucket, s.hostMapFile, &hostMap)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return s.sortClassifiedHosts(hostMap, tnsHosts), nil
}

func (s *onpremMappingService) sortClassifiedHosts(hostMap map[string]Host, dvhTNSHosts []service.TNSName) *service.ClassifiedHosts {
	c := &service.ClassifiedHosts{
		Hosts: map[service.OnpremHostType][]*service.Host{},
	}

	tnsHosts := map[string]struct{}{}

	for _, tnsHost := range dvhTNSHosts {
		tnsHosts[tnsHost.Host] = struct{}{}

		c.Hosts[service.OnpremHostTypeTNS] = append(c.Hosts[service.OnpremHostTypeTNS], &service.Host{
			Name:        tnsHost.TnsName,
			Description: tnsHost.Description,
			Host:        tnsHost.Host,
		})
	}

	for host, data := range hostMap {
		hostType := service.OnpremHostType(data.Type)

		// Skip hosts that are children
		if len(data.Parent) > 0 {
			continue
		}

		// Skip TNS hosts
		if _, isTNSHost := tnsHosts[host]; isTNSHost {
			continue
		}

		if service.ValidOnpremHostType(hostType) {
			c.Hosts[hostType] = append(c.Hosts[hostType], &service.Host{
				Name:        host,
				Description: data.Description,
				Host:        host,
			})

			continue
		}

		s.log.Error().Msgf("Invalid host type: %s", data.Type)
	}

	for _, hosts := range c.Hosts {
		sort.Slice(hosts, func(i, j int) bool {
			return hosts[i].Host < hosts[j].Host
		})
	}

	return c
}

func NewOnpremMappingService(bucket, hostmapFile string, cloudStorageAPI service.CloudStorageAPI, dvhAPI service.DatavarehusAPI, log zerolog.Logger) *onpremMappingService {
	return &onpremMappingService{
		bucket:          bucket,
		hostMapFile:     hostmapFile,
		dvhAPI:          dvhAPI,
		cloudStorageAPI: cloudStorageAPI,
		log:             log,
	}
}
