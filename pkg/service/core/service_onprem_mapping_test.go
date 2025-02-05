package core

import (
	"testing"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/stretchr/testify/assert"
)

func TestSortClassifiedHosts(t *testing.T) {
	testCases := []struct {
		name        string
		hostMap     map[string]Host
		dvhTNSHosts []service.TNSName
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
					Type: "oracle",
				},
				"oracle-host-vip.domain.no": {
					Description: "Oracle VIP host",
					IPs: []string{
						"1.2.3.4",
					},
					Port:   "1521",
					Type:   "oracle",
					Parent: "oracle-host-scan.domain.no",
				},
				"oracle-non-vip-host.domain.no": {
					Description: "Oracle non VIP host",
					IPs: []string{
						"4.3.2.1",
					},
					Port: "1521",
					Type: "oracle",
				},
				"postgres-host.domain.no": {
					Description: "Postgres host",
					IPs: []string{
						"11.22.33.44",
					},
					Port: "5432",
					Type: "postgres",
				},
				"informatica.domain.no": {
					Description: "Informatica host",
					IPs: []string{
						"44.33.22.11",
					},
					Port: "6005-6120",
					Type: "informatica",
				},
				"address.nav.no": {
					Description: "HTTP host",
					IPs: []string{
						"123.123.123.123",
					},
					Port: "443",
					Type: "http",
				},
				"sftp.nav.no": {
					Description: "SFTP host",
					IPs: []string{
						"255.255.255.255",
					},
					Port: "22",
					Type: "sftp",
				},
			},
			dvhTNSHosts: []service.TNSName{
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
				HTTPHosts: []service.Host{
					{
						Host:        "address.nav.no",
						Description: "HTTP host",
					},
				},
				UnclassifiedHosts: []service.Host{
					{
						Host:        "sftp.nav.no",
						Description: "SFTP host",
					},
				},
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
