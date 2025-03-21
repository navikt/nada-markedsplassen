package integration

import (
	"context"
	http2 "net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/navikt/nada-backend/pkg/cloudstorage"
	"github.com/navikt/nada-backend/pkg/cloudstorage/emulator"
	"github.com/navikt/nada-backend/pkg/datavarehus"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core"
	"github.com/navikt/nada-backend/pkg/service/core/api/gcp"
	"github.com/navikt/nada-backend/pkg/service/core/api/http"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/routes"
	"github.com/rs/zerolog"
)

const (
	onpremMapping = `
postgres.database.no:
  description: "Postgres database"
  ips:
  - "11.22.33.44"
  port: 5432
  type: postgres
informatica.database.no:
  description: "Informatica database"
  ips:
  - "111.222.255.255"
  port: 6005-6120
  type: informatica
dm08-scan.adeo.no:
  description: "Oracle database dm08-scan.adeo.no"
  ips:
  - "1.2.3.4"
  port: 1521
  type: oracle
a01dbfl036.adeo.no:
  description: "Oracle database a01dbfl036.adeo.no"
  ips:
  - "44.33.22.11"
  port: 1521
  type: oracle
dmv34-scan.adeo.no:
  description: "Oracle database dmv34-scan.adeo.no"
  ips:
  - "44.33.22.11"
  port: 1521
  type: oracle
address.nav.no:
  description: "HTTP host"
  ips:
  - "123.123.123.123"
  port: 443
  type: http
`
)

func TestOnpremMapping(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	log := zerolog.New(os.Stdout)

	bucketName := "onprem-mapping"
	mappingObjectName := "onprem_mapping.yaml"

	c := NewContainers(t, log)
	defer c.Cleanup()

	cfg := c.RunDatavarehus()

	dvhClient := datavarehus.New(cfg.ConnectionURL(), "client_id", "client_secret")
	dvhAPI := http.NewDatavarehusAPI(dvhClient, log)

	e := emulator.New(t, []fakestorage.Object{
		{
			ObjectAttrs: fakestorage.ObjectAttrs{
				BucketName: bucketName,
				Name:       mappingObjectName,
			},
			Content: []byte(onpremMapping),
		},
	})
	defer e.Cleanup()

	client := cloudstorage.NewFromClient(e.Client())
	cloudStorageAPI := gcp.NewCloudStorageAPI(client, log)

	onpremMappingService := core.NewOnpremMappingService(
		bucketName,
		mappingObjectName,
		cloudStorageAPI,
		dvhAPI,
		log,
	)

	r := TestRouter(log)
	{
		h := handlers.NewOnpremMappingHandler(onpremMappingService)
		e := routes.NewOnpremMappingEndpoints(log, h)
		f := routes.NewOnpremMappingRoutes(e)

		f(r)
	}

	server := httptest.NewServer(r)
	defer server.Close()

	t.Run("TNS name hosts are extracted from mapping_config file", func(t *testing.T) {
		expect := &service.ClassifiedHosts{
			Hosts: map[service.OnpremHostType][]*service.Host{
				service.OnpremHostTypeHTTP: {
					{
						Name:        "address.nav.no",
						Description: "HTTP host",
						Host:        "address.nav.no",
					},
				},
				service.OnpremHostTypeInformatica: {
					{
						Name:        "informatica.database.no",
						Description: "Informatica database",
						Host:        "informatica.database.no",
					},
				},
				service.OnpremHostTypePostgres: {
					{
						Name:        "postgres.database.no",
						Description: "Postgres database",
						Host:        "postgres.database.no",
					},
				},
				service.OnpremHostTypeTNS: {
					{
						Name:        "DVH-I",
						Description: "Live datamart mot DVH-P som benyttes til konsum av dataobjekter gjort tilgjengelig i IAPP-sonen.",
						Host:        "a01dbfl036.adeo.no",
					},
					{
						Name:        "DVH-P",
						Description: "Produksjonsdatabasen til Nav-datavarehus. Den benyttes til ETL/lagring og konsum av dataobjekter i fagsystemsonen.",
						Host:        "dm08-scan.adeo.no",
					},
					{
						Name:        "DVH-R",
						Description: "Kopi av produksjon (synkroniseres hver natt) som benyttes til test og utvikling.",
						Host:        "dmv34-scan.adeo.no",
					},
					{
						Name:        "DVH-U",
						Description: "Kopi av produksjon (manuelt oppdatert) som benyttes til utvikling.",
						Host:        "dmv34-scan.adeo.no",
					},
					{
						Name:        "DVH-Q",
						Description: "Kopi av produksjon (manuelt oppdatert) som benyttes til test av deploy.",
						Host:        "dmv38-scan.adeo.no",
					},
				},
			},
		}

		into := &service.ClassifiedHosts{}

		NewTester(t, server).
			Get(ctx, "/api/onpremMapping").
			HasStatusCode(http2.StatusOK).Expect(expect, into)
	})
}
