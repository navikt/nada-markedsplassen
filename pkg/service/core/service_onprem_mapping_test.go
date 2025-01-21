package core

import (
	"testing"

	"github.com/navikt/nada-backend/pkg/datavarehus"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/stretchr/testify/assert"
)

func TestSortClassifiedHosts(t *testing.T) {
	testCases := []struct {
		name        string
		hostMap     map[string]Host
		dvhTNSHosts []datavarehus.TNSName
		expect      *service.ClassifiedHosts
	}{
		{
			name: "Sort hosts",
			hostMap: map[string]Host{
				"oracle-host-scan.domain.no": {
					Description: "Oracle host scan",
					IPs: []string{
						"1.2.3.4",
						"5.6.7.8",
					},
					Port: "1521",
				},
				"oracle-host-vip.domain.no": {
					Description: "Oracle VIP host",
					IPs: []string{
						"1.2.3.4",
					},
					Port: "1521",
				},
				"oracle-non-vip-host.domain.no": {
					Description: "Oracle non VIP host",
					IPs: []string{
						"4.3.2.1",
					},
					Port: "1521",
				},
				"postgres-host.domain.no": {
					Description: "Postgres host",
					IPs: []string{
						"11.22.33.44",
					},
					Port: "5432",
				},
				"informatica.domain.no": {
					Description: "Informatica host",
					IPs: []string{
						"44.33.22.11",
					},
					Port: "6005-6120",
				},
			},
			dvhTNSHosts: []datavarehus.TNSName{
				{
					Host:        "oracle-host-scan.domain.no",
					TnsName:     "host",
					Description: "host with tns",
				},
			},
			expect: &service.ClassifiedHosts{
				DVHHosts: []service.TNSHost{
					{
						Host:        "oracle-host-scan.domain.no",
						Description: "host with tns",
						TNSName:     "host",
					},
				},
				OracleHosts: []service.Host{
					{
						Host:        "oracle-host-scan.domain.no",
						Description: "Oracle host scan",
					},
					{
						Host:        "oracle-host-vip.domain.no",
						Description: "Oracle VIP host",
					},
					{
						Host:        "oracle-non-vip-host.domain.no",
						Description: "Oracle non VIP host",
					},
				},
				PostgresHosts: []service.Host{
					{
						Host:        "postgres-host.domain.no",
						Description: "Postgres host",
					},
				},
				InformaticaHosts: []service.Host{
					{
						Host:        "informatica.domain.no",
						Description: "Informatica host",
					},
				},
				UnclassifiedHosts: []service.Host{},
			},
		},
	}

	onpremMappingService := &onpremMappingService{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			classifiedHosts := onpremMappingService.sortClassifiedHosts(tc.hostMap, tc.dvhTNSHosts)
			assert.Equal(t, tc.expect, classifiedHosts)
		})
	}
}
