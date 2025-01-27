module github.com/navikt/nada-backend

go 1.23.3

replace sourcegraph.com/sourcegraph/appdash => github.com/sourcegraph/appdash v0.0.0-20211028080628-e2786a622600

replace sourcegraph.com/sourcegraph/appdash-data => github.com/sourcegraph/appdash-data v0.0.0-20151005221446-73f23eafcf67

require (
	cloud.google.com/go/artifactregistry v1.16.1
	cloud.google.com/go/bigquery v1.66.0
	cloud.google.com/go/compute v1.31.1
	cloud.google.com/go/iam v1.3.1
	cloud.google.com/go/logging v1.13.0
	cloud.google.com/go/longrunning v0.6.4
	cloud.google.com/go/storage v1.50.0
	cloud.google.com/go/workstations v1.1.3
	github.com/btcsuite/btcutil v1.0.2
	github.com/containers/image/v5 v5.33.1
	github.com/coreos/go-oidc/v3 v3.12.0
	github.com/docker/cli v27.5.1+incompatible
	github.com/emersion/go-sasl v0.0.0-20241020182733-b788ff22d5a6
	github.com/emersion/go-smtp v0.21.3
	github.com/fsouza/fake-gcs-server v1.52.1
	github.com/gilcrest/diygoapi v0.53.0
	github.com/go-chi/chi/v5 v5.2.0
	github.com/go-chi/cors v1.2.1
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/goccy/bigquery-emulator v0.6.6
	github.com/goccy/go-json v0.10.4
	github.com/golang-jwt/jwt/v4 v4.5.1
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/go-cmp v0.6.0
	github.com/google/uuid v1.6.0
	github.com/gosimple/slug v1.15.0
	github.com/jackc/pgx/v5 v5.7.2
	github.com/jarcoal/httpmock v1.3.1
	github.com/lib/pq v1.10.9
	github.com/lithammer/shortuuid/v4 v4.2.0
	github.com/mileusna/useragent v1.3.5
	github.com/mitchellh/mapstructure v1.5.0
	github.com/ory/dockertest/v3 v3.11.0
	github.com/pkg/errors v0.9.1
	github.com/pressly/goose/v3 v3.24.1
	github.com/prometheus/client_golang v1.20.5
	github.com/riverqueue/river v0.15.0
	github.com/riverqueue/river/riverdriver/riverdatabasesql v0.15.0
	github.com/riverqueue/river/rivertype v0.15.0
	github.com/rs/zerolog v1.33.0
	github.com/samber/slog-zerolog/v2 v2.7.3
	github.com/sebdah/goldie/v2 v2.5.5
	github.com/slack-go/slack v0.15.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.19.0
	github.com/sqlc-dev/pqtype v0.3.0
	github.com/sqlc-dev/sqlc v1.28.0
	github.com/stretchr/testify v1.10.0
	golang.org/x/exp v0.0.0-20250106191152-7588d65b2ba8
	golang.org/x/oauth2 v0.25.0
	google.golang.org/api v0.218.0
	google.golang.org/genproto v0.0.0-20250106144421-5f5ef82da422
	google.golang.org/grpc v1.70.0
	google.golang.org/protobuf v1.36.3
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cel.dev/expr v0.19.1 // indirect
	cloud.google.com/go v0.118.0 // indirect
	cloud.google.com/go/auth v0.14.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.7 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	cloud.google.com/go/monitoring v1.22.1 // indirect
	cloud.google.com/go/pubsub v1.45.3 // indirect
	dario.cat/mergo v1.0.1 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/DataDog/go-hll v1.0.2 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.25.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.49.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.49.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/apache/arrow/go/v10 v10.0.1 // indirect
	github.com/apache/arrow/go/v15 v15.0.2 // indirect
	github.com/apache/thrift v0.21.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20241223141626-cff3c89139a3 // indirect
	github.com/containerd/continuity v0.4.5 // indirect
	github.com/containers/libtrust v0.0.0-20230121012942-c1716e8a8d01 // indirect
	github.com/containers/ocicrypt v1.2.1 // indirect
	github.com/containers/storage v1.56.1 // indirect
	github.com/creack/pty v1.1.20 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-farm v0.0.0-20240924180020-3414d57e47da // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/dlclark/regexp2 v1.11.4 // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker v27.4.1+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.8.2 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dop251/goja v0.0.0-20241024094426-79f3a7efcdbd // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.32.3 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.1.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-jose/go-jose/v4 v4.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.23.0 // indirect
	github.com/go-sourcemap/sourcemap v2.1.4+incompatible // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/goccy/go-yaml v1.15.13 // indirect
	github.com/goccy/go-zetasql v0.5.5 // indirect
	github.com/goccy/go-zetasqlite v0.19.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/flatbuffers v24.12.23+incompatible // indirect
	github.com/google/pprof v0.0.0-20241210010833-40e02aabc2ad // indirect
	github.com/google/renameio/v2 v2.0.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/gorilla/handlers v1.5.2 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/asmfmt v1.3.2 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/linkedin/goavro/v2 v2.13.0 // indirect
	github.com/magiconair/properties v1.8.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mattn/go-sqlite3 v1.14.24 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/minio/asm2plan9s v0.0.0-20200509001527-cdd76441f9d8 // indirect
	github.com/minio/c2goasm v0.0.0-20190812172519-36a3d3bbc4f3 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/sys/capability v0.4.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/user v0.3.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/opencontainers/runc v1.2.4 // indirect
	github.com/opencontainers/runtime-spec v1.2.0 // indirect
	github.com/parquet-go/parquet-go v0.24.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pkg/xattr v0.4.10 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.61.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/riverqueue/river/riverdriver v0.15.0 // indirect
	github.com/riverqueue/river/rivershared v0.15.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/sagikazarmark/locafero v0.6.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/samber/lo v1.47.0 // indirect
	github.com/samber/slog-common v0.18.1 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.33.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.58.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.58.0 // indirect
	go.opentelemetry.io/otel v1.33.0 // indirect
	go.opentelemetry.io/otel/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/sdk v1.33.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/trace v1.33.0 // indirect
	go.uber.org/goleak v1.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/time v0.9.0 // indirect
	golang.org/x/tools v0.29.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	gonum.org/v1/gonum v0.15.1 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
