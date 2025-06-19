package integration

import (
	"context"
	"fmt"
	gohttp "net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/billing/apiv1/billingpb"
	"github.com/navikt/nada-backend/pkg/cloudbilling"
	"github.com/navikt/nada-backend/pkg/datavarehus"
	"github.com/navikt/nada-backend/pkg/iamcredentials"
	"github.com/navikt/nada-backend/pkg/service/core/api/http"
	riverstore "github.com/navikt/nada-backend/pkg/service/core/queue/river"
	"github.com/navikt/nada-backend/pkg/service/core/storage/postgres"
	"github.com/navikt/nada-backend/pkg/worker/worker_args"
	crmv3 "google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/networksecurity/v1"
	"google.golang.org/genproto/googleapis/type/money"

	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"cloud.google.com/go/iam/apiv1/iampb"

	"github.com/navikt/nada-backend/pkg/artifactregistry"
	"golang.org/x/exp/maps"

	"github.com/navikt/nada-backend/pkg/worker"

	"github.com/navikt/nada-backend/pkg/database"
	"github.com/riverqueue/river"
	"google.golang.org/api/iam/v1"

	logpb "cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/navikt/nada-backend/pkg/cloudlogging"
	"github.com/stretchr/testify/require"
	ltype "google.golang.org/genproto/googleapis/logging/type"
	"google.golang.org/protobuf/types/known/timestamppb"

	"cloud.google.com/go/compute/apiv1/computepb"

	artifactRegistryEmulator "github.com/navikt/nada-backend/pkg/artifactregistry/emulator"
	cloudLoggingEmulator "github.com/navikt/nada-backend/pkg/cloudlogging/emulator"
	crm "github.com/navikt/nada-backend/pkg/cloudresourcemanager"
	crmEmulator "github.com/navikt/nada-backend/pkg/cloudresourcemanager/emulator"
	"github.com/navikt/nada-backend/pkg/computeengine"
	computeEmulator "github.com/navikt/nada-backend/pkg/computeengine/emulator"
	iamCredentialsEmulator "github.com/navikt/nada-backend/pkg/iamcredentials/emulator"
	"google.golang.org/api/cloudresourcemanager/v1"

	"cloud.google.com/go/workstations/apiv1/workstationspb"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/navikt/nada-backend/pkg/sa"
	serviceAccountEmulator "github.com/navikt/nada-backend/pkg/sa/emulator"
	"github.com/navikt/nada-backend/pkg/securewebproxy"
	secureWebProxyEmulator "github.com/navikt/nada-backend/pkg/securewebproxy/emulator"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core"
	"github.com/navikt/nada-backend/pkg/service/core/api/gcp"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/routes"
	"github.com/navikt/nada-backend/pkg/workstations"
	"github.com/navikt/nada-backend/pkg/workstations/emulator"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

const publicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAr0Jswk8lh4IOaWhpsa0i
JhZt0RS0NNyFrRLBZhkuuQP1nns+c3gnYssG4esWyKrJrma0McHBYRhtZZALB40Q
G2UXFVjNmStcotTh/iNpdDBAnoaYax7iq88IDBh5pqE3FrpPrjRdnDd/dIrx6lJS
zvaN5jZAw8x26GKN4HyiRvEfc9TNcbgrDuld84XcnEqZdMreJrkO6X9gxf4vpBM9
lPAuDqMFUn5DvQCvp9PI4Fal+vm3zDThFD/lE6bp/K/cau93C++XN47vhJn7oYV0
utg1eWFqGmxnsO3I85slL2F0BBIEC74bNyGu/8bXEPtP4fKiSO/2VEtzrEYG2/H/
twIDAQAB
-----END PUBLIC KEY-----
`

const privateKey = `-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCvQmzCTyWHgg5p
aGmxrSImFm3RFLQ03IWtEsFmGS65A/Weez5zeCdiywbh6xbIqsmuZrQxwcFhGG1l
kAsHjRAbZRcVWM2ZK1yi1OH+I2l0MECehphrHuKrzwgMGHmmoTcWuk+uNF2cN390
ivHqUlLO9o3mNkDDzHboYo3gfKJG8R9z1M1xuCsO6V3zhdycSpl0yt4muQ7pf2DF
/i+kEz2U8C4OowVSfkO9AK+n08jgVqX6+bfMNOEUP+UTpun8r9xq73cL75c3ju+E
mfuhhXS62DV5YWoabGew7cjzmyUvYXQEEgQLvhs3Ia7/xtcQ+0/h8qJI7/ZUS3Os
Rgbb8f+3AgMBAAECggEADpwuTjHO4nGtxeJrEowTprKD9m66vxVgchci1uIOimSR
cI68RrVOLebYQgkZCHgZrAI/z04O/YsjGNkIh66eGHU1lr/6aQ8Q/+St75kAIiwg
7EDdf+s+i4K3cbAF+XrDCY//3c6GZzlxKe6opWy2HoQLPD/APRJUb0xCoOOC8QA7
7cQX0E8n9+n1J76TzBDQgJnBBYC2G5V4eGdYgVgDsy3kaWpkFYsLTQWeSSeO8+qZ
QKPw6Yu0jQs812ajUUcMl53tKsZkpYJAnsPodTTw3lq5oJFF6XJ9Jf89DzhIayro
e7A/83zaDGlVfj5LxNSSF37ykKbToh/3ykefMuEioQKBgQDvF1Mu7jCptQH+hqff
h3idNJERYDNBjQ8m14mK1+mf5gz4czCeBtVY0WkDsobk8PHj+OtgawPxhun75AfA
qFwVKAQ1zoYP9rZDpqMkNys8Wiig7OBUEj4FbnGmnFP+kTEeG9LYxPHLA5ssPtGj
ycrrt76bY75Jr0cCEAvOqHWHTwKBgQC7p3JU/ZhsbwZHsqqnZSMd1BjcZfoPIrqy
TK4Eguo/GW4dfgOHsWnWtLHJ0TokZ6bdyJVoFciYLImFuUiA1SNNXlp1aBrNk24k
TnX00jeytax9DzIHsOHCS6ahrdbteiPLRhWTgC1MrLKPxM9erZq3IaBcAgZFS0mk
JIR/fslnGQKBgDLZMRXADpVpK51oIffGJf65GUkqvnvodhp6qIPg24zoLkYAqYxS
Q7l5/+2LYGj8XVVwsQ52dAY//S9XFdcBd2QAeLTA0X4/qA/HNtcS7J0PR6jB+Aup
PYuGK6GVib+QPXP70uHLMOlOQQgt7AP7fK6ZC26czfF51442v2waI7S9AoGAMpqI
GWU9mlgiQGls3bFHU/7jKWQSl8xMvlIxRyQqmRN5f1iBCTGNkgmuO/dBD5ooBHzX
1XayXl78QuRhKeTQHUgJasnFGJTeScoiwv+BZ57YQe08F5jaeHPAHq9rWyTpzCI9
JUaWcKvNhzmSljyIkUPvI4CkQkF4PVxfoqYFF9kCgYBUmL3Vx0rvE7h6ZKKlKJhL
LOUOLmXXOUsc64A5WeIpW7kuMaJJG3xeTQUjTe8py0Kif/UuXgawNSYNYLKKhNtO
HinoLzcgmwn4soz/TKYJXZOk1Czorxb1WH3uELTCh49fpwYIstJxuurB8sLqfMQn
TlNiEcMXKC9tdGiqi6SstA==
-----END PRIVATE KEY-----
`

func TestWorkstations(t *testing.T) {
	t.Parallel()

	log := zerolog.New(zerolog.NewConsoleWriter()).Level(zerolog.InfoLevel)
	ctx := context.Background()

	e := emulator.New(log)
	e.SetWorkstationConfigReplicaZones([]string{
		"europe-north1-a",
		"europe-north1-b",
	},
	)

	apiURL := e.Run()
	project := "test"
	location := "europe-north1"
	repository := "knast-images"
	clusterID := "clusterID"
	workstationHost := "ident.workstations.domain"
	slug := UserOneIdent

	c := NewContainers(t, log)
	defer c.Cleanup()

	pgCfg := c.RunPostgres(NewPostgresConfig())

	repo, err := database.New(
		pgCfg.ConnectionURL(),
		10,
		10,
	)
	assert.NoError(t, err)

	client := workstations.New(project, location, clusterID, apiURL, true, gohttp.DefaultClient)

	saEmulator := serviceAccountEmulator.New(log)
	saURL := saEmulator.Run()
	saClient := sa.NewClient(saURL, true)
	saEmulator.SetIamPolicy(
		fmt.Sprintf("%s@%s.iam.gserviceaccount.com", service.WorkstationServiceAccountID(slug), project),
		&iam.Policy{},
	)

	crmEmulator := crmEmulator.New(log)
	crmEmulator.SetPolicy(project, &cloudresourcemanager.Policy{
		Bindings: []*cloudresourcemanager.Binding{
			{
				Role:    "roles/owner",
				Members: []string{fmt.Sprintf("user:%s", UserOne.Email)},
			},
		},
	})
	crmURL := crmEmulator.Run()
	tagBindingClient := crmEmulator.TagBindingPolicyClient(
		&crmv3.ListEffectiveTagsResponse{
			EffectiveTags: []*crmv3.EffectiveTag{
				{
					TagValue:           "tagValues/281479612953454",
					NamespacedTagValue: "433637338589/drz-location/europe-north1",
					TagKey:             "tagKeys/281476476262619",
					NamespacedTagKey:   "433637338589/drz-location",
					TagKeyParentName:   "organizations/433637338589",
				},
			},
		},
		[]string{"europe-north1-a", "europe-north1-b"},
		gohttp.StatusOK,
		log,
	)
	crmClient := crm.NewClient(crmURL, true, tagBindingClient)

	swpEmulator := secureWebProxyEmulator.New(log)
	swpEmulator.SetURLList(project, location, service.GlobalURLAllowListName, &networksecurity.UrlList{
		Values: []string{
			"github.com/navikt/*",
		},
	})
	swpURL := swpEmulator.Run()
	swpClient := securewebproxy.New(swpURL, true)

	arEmulator := artifactRegistryEmulator.New(log)
	arEmulator.SetImages(map[string][]*artifactregistrypb.DockerImage{
		fmt.Sprintf("projects/%s/locations/%s/repositories/%s", project, location, repository): {
			{
				Name: "nginx",
				Uri:  "eu-north1-docker.pkg.dev/test/test/nginx@sha256:e9954c1fc875017be1c3e36eca16be2d9e9bccc4bf072163515467d6a823c7cf",
				Tags: []string{"latest"},
			},
			{
				Name: "something",
				Uri:  "eu-north1-docker.pkg.dev/test/test/something@sha256:e9954c1fc875017be1c3e36eca16be2d9e9bccc4bf072163515467d6a823c7cf",
				Tags: []string{"v1"},
			},
		},
	})
	arEmulator.SetIamPolicy(repository, &iampb.Policy{})
	arURL := arEmulator.Run()
	arClient := artifactregistry.New(arURL, true)

	ceEmulator := computeEmulator.New(log)
	ceEmulator.SetInstances(map[string][]*computepb.Instance{
		"europe-north1-a": {
			{
				Name: strToStrPtr("instance-1"),
				Labels: map[string]string{
					service.WorkstationConfigIDLabel: slug,
				},
				Id: uint64Ptr(12345),
				NetworkInterfaces: []*computepb.NetworkInterface{
					{
						NetworkIP: strToStrPtr("10.5.6.127"),
					},
				},
			},
		},
	})
	ceEmulator.SetFirewallPolicies(map[string]*computepb.FirewallPolicy{
		"test-europe-north1-onprem-access": {
			Name: strToStrPtr("onprem-access"),
			Rules: []*computepb.FirewallPolicyRule{
				{
					RuleName: strToStrPtr("rule-1"),
					TargetSecureTags: []*computepb.FirewallPolicyRuleSecureTag{
						{
							Name: strToStrPtr("test/rule-1/rule-1"),
						},
					},
				},
			},
		},
	})

	clEmulator := cloudLoggingEmulator.NewEmulator(log)
	clEmulator.AppendLogEntries([]*logpb.LogEntry{
		{
			LogName: "ehh",
			HttpRequest: &ltype.HttpRequest{
				RequestMethod: gohttp.MethodGet,
				RequestUrl:    "https://github.com/navikt",
				UserAgent:     "Chrome 91",
			},
			Timestamp: timestamppb.Now(),
		},
	})
	clURL, err := clEmulator.Start()
	require.NoError(t, err)
	clClient := cloudlogging.NewClient(clURL, true)
	clAPI := gcp.NewCloudLoggingAPI(clClient)

	computeURL := ceEmulator.Run()
	computeClient := computeengine.NewClient(computeURL, true)
	computeAPI := gcp.NewComputeAPI(Project, computeClient)

	workers := river.NewWorkers()

	config := worker.RiverConfig(&log, workers)
	config.TestOnly = true

	workstationsQueue := riverstore.NewWorkstationsQueue(config, repo)
	workstationsStorage := postgres.NewWorkstationsStorage(repo)

	arAPI := gcp.NewArtifactRegistryAPI(arClient, log)

	iamCredentialsEmulator := iamCredentialsEmulator.New(log)
	iamCredentialsEmulator.AddSigner("test@test-project.iam.gserviceaccount.com", publicKey, privateKey)
	iamCredentialsURL := iamCredentialsEmulator.Run()
	iamCredentialsClient := iamcredentials.New(iamCredentialsURL, true)

	datavarehusConfig := c.RunDatavarehus()
	datavarehusClient := datavarehus.New(
		fmt.Sprintf("http://%s", datavarehusConfig.HostPort),
		"",
		"",
	)

	router := TestRouter(log)
	crmAPI := gcp.NewCloudResourceManagerAPI(crmClient)
	gAPI := gcp.NewWorkstationsAPI(client)
	saAPI := gcp.NewServiceAccountAPI(saClient)
	swpAPI := gcp.NewSecureWebProxyAPI(log, swpClient)
	iamCredentialsAPI := gcp.NewIAMCredentialsAPI(iamCredentialsClient)
	datavarehusAPI := http.NewDatavarehusAPI(datavarehusClient, log)

	cloudBillingApi := cloudbilling.NewStaticClient(map[string]*billingpb.Sku{
		cloudbilling.SkuCpu: {
			SkuId: cloudbilling.SkuCpu,
			PricingInfo: []*billingpb.PricingInfo{
				{
					PricingExpression: &billingpb.PricingExpression{
						TieredRates: []*billingpb.PricingExpression_TierRate{
							{
								UnitPrice: &money.Money{
									Nanos: 123000000,
								},
							},
						},
					},
				},
			},
		},
		cloudbilling.SkuMemory: {
			SkuId: cloudbilling.SkuMemory,
			PricingInfo: []*billingpb.PricingInfo{
				{
					PricingExpression: &billingpb.PricingExpression{
						TieredRates: []*billingpb.PricingExpression_TierRate{
							{
								UnitPrice: &money.Money{
									Nanos: 234000000,
								},
							},
						},
					},
				},
			},
		},
	})
	workstationService := core.NewWorkstationService(
		project,
		project,
		location,
		"my-policy",
		"onprem-access",
		"my-bucket",
		"my-view",
		fmt.Sprintf("admin-sa@%s.iam.gserviceaccount.com", project),
		repository,
		project,
		"test@test-project.iam.gserviceaccount.com",
		"localhost",
		saAPI,
		crmAPI,
		swpAPI,
		gAPI,
		workstationsStorage,
		computeAPI,
		clAPI,
		workstationsQueue,
		arAPI,
		datavarehusAPI,
		iamCredentialsAPI,
		cloudBillingApi,
		log,
	)

	{
		h := handlers.NewWorkstationsHandler(workstationService)
		e := routes.NewWorkstationsEndpoints(log, h)
		f := routes.NewWorkstationsRoutes(e, InjectUser(UserOne))
		f(router)
	}

	err = worker.WorkstationAddWorkers(config, workstationService, repo)
	require.NoError(t, err)

	workstationWorker, err := worker.RiverClient(config, repo)
	require.NoError(t, err)

	err = workstationWorker.Start(ctx)
	require.NoError(t, err)
	defer workstationWorker.Stop(ctx)

	server := httptest.NewServer(router)

	fullyQualifiedConfigName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s", project, location, clusterID, slug)

	workstation := &service.WorkstationOutput{}

	t.Run("Get workstation options", func(t *testing.T) {
		hourlyCost, _ := cloudBillingApi.GetHourlyCostInNOKFromSKU(ctx)
		expected := &service.WorkstationOptions{
			MachineTypes: service.WorkstationMachineTypes(hourlyCost.CostForConfiguration),
			ContainerImages: []*service.WorkstationContainer{
				{
					Image:       "eu-north1-docker.pkg.dev/test/test/nginx:latest",
					Description: "nginx",
					Labels:      map[string]string{},
				},
			},
			GlobalURLAllowList: []string{
				"github.com/navikt/*",
			},
		}

		got := &service.WorkstationOptions{}

		NewTester(t, server).
			Get(ctx, "/api/workstations/options").
			HasStatusCode(gohttp.StatusOK).
			Expect(expected, got)
	})

	t.Run("Create workstation", func(t *testing.T) {
		expectedJob := &service.WorkstationJob{
			JobHeader: service.JobHeader{
				ID:        1,
				State:     service.JobStatePending,
				Duplicate: false,
				Errors:    []string{},
				Kind:      worker_args.WorkstationJobKind,
			},
			Name:           "User Userson",
			Email:          "user.userson@email.com",
			Ident:          "v101010",
			MachineType:    service.MachineTypeN2DStandard16,
			ContainerImage: service.ContainerImageVSCode,
		}

		subscribeChan, subscribeCancel := workstationWorker.Subscribe(river.EventKindJobCompleted)
		go func() {
			time.Sleep(65 * time.Second)
			subscribeCancel()
		}()

		job := &service.WorkstationJob{}

		NewTester(t, server).
			Post(ctx, service.WorkstationInput{
				MachineType:    service.MachineTypeN2DStandard16,
				ContainerImage: service.ContainerImageVSCode,
			}, "/api/workstations/job").
			HasStatusCode(gohttp.StatusAccepted).
			Expect(expectedJob, job, cmpopts.IgnoreFields(service.WorkstationJob{}, "StartTime", "EndTime"))

		event := <-subscribeChan
		assert.Equal(t, river.EventKindJobCompleted, event.Kind)

		expectedJob.State = service.JobStateCompleted

		NewTester(t, server).
			Get(ctx, fmt.Sprintf("/api/workstations/job/%d", job.ID)).
			HasStatusCode(gohttp.StatusOK).
			Expect(expectedJob, job, cmpopts.IgnoreFields(service.WorkstationJob{}, "StartTime", "EndTime"))

		expectedWorkstation := &service.WorkstationOutput{
			Slug:        slug,
			DisplayName: "User Userson (user.userson@email.com)",
			State:       service.Workstation_STATE_STARTING,
			Config: &service.WorkstationConfigOutput{
				MachineType:    service.MachineTypeN2DStandard16,
				Image:          service.ContainerImageVSCode,
				IdleTimeout:    2 * time.Hour,
				RunningTimeout: 12 * time.Hour,
			},
			Host: workstationHost,
		}

		workstation = &service.WorkstationOutput{}
		NewTester(t, server).
			Get(ctx, "/api/workstations/").
			HasStatusCode(gohttp.StatusOK).
			Expect(expectedWorkstation, workstation, cmpopts.IgnoreFields(service.WorkstationOutput{}, "CreateTime", "Config.UpdateTime", "StartTime", "Config.CreateTime", "Config.Env"))

		assert.Truef(t, maps.Equal(workstation.Config.Env, service.DefaultWorkstationEnv(slug, UserOne.Email, UserOneName)), "Expected %v, got %v", map[string]string{"WORKSTATION_NAME": slug}, workstation.Config.Env)
	})

	t.Run("Change running state", func(t *testing.T) {
		e.SetWorkstationState(fullyQualifiedConfigName, slug, workstationspb.Workstation_STATE_RUNNING)

		expected := &service.WorkstationOutput{
			Slug:        slug,
			DisplayName: "User Userson (user.userson@email.com)",
			Reconciling: false,
			UpdateTime:  nil,
			StartTime:   nil,
			State:       service.Workstation_STATE_RUNNING,
			Config: &service.WorkstationConfigOutput{
				UpdateTime:     nil,
				IdleTimeout:    2 * time.Hour,
				RunningTimeout: 12 * time.Hour,
				MachineType:    service.MachineTypeN2DStandard16,
				Image:          service.ContainerImageVSCode,
			},
			Host: workstationHost,
		}

		NewTester(t, server).
			Get(ctx, "/api/workstations/").Debug(os.Stdout).
			HasStatusCode(gohttp.StatusOK).
			Expect(expected, workstation, cmpopts.IgnoreFields(service.WorkstationOutput{}, "CreateTime", "StartTime", "Config.CreateTime", "Config.UpdateTime", "Config.Env"))
		assert.NotNil(t, workstation.StartTime)
	})

	t.Run("Get workstation logs", func(t *testing.T) {
		expected := &service.WorkstationLogs{
			ProxyDeniedHostPaths: []*service.LogEntry{
				{
					Timestamp: time.Time{},
					HTTPRequest: &service.HTTPRequest{
						URL: &url.URL{
							Host: "github.com",
							Path: "/navikt",
						},
						Method:    gohttp.MethodGet,
						UserAgent: "Chrome 91",
					},
				},
			},
		}

		got := &service.WorkstationLogs{}

		NewTester(t, server).
			Get(ctx, "/api/workstations/logs").
			HasStatusCode(gohttp.StatusOK).
			Expect(expected, got, cmpopts.IgnoreFields(service.LogEntry{}, "Timestamp", "HTTPRequest.URL.Scheme"))
	})

	t.Run("Update workstation", func(t *testing.T) {
		expectedJob := &service.WorkstationJob{
			JobHeader: service.JobHeader{
				ID:        2,
				State:     service.JobStatePending,
				Duplicate: false,
				Errors:    []string{},
				Kind:      worker_args.WorkstationJobKind,
			},
			Name:           "User Userson",
			Email:          "user.userson@email.com",
			Ident:          "v101010",
			MachineType:    service.MachineTypeN2DStandard32,
			ContainerImage: service.ContainerImageIntellijUltimate,
		}

		subscribeChan, subscribeCancel := workstationWorker.Subscribe(river.EventKindJobCompleted)
		go func() {
			time.Sleep(5 * time.Second)
			subscribeCancel()
		}()

		job := &service.WorkstationJob{}

		NewTester(t, server).
			Post(ctx, service.WorkstationInput{
				MachineType:    service.MachineTypeN2DStandard32,
				ContainerImage: service.ContainerImageIntellijUltimate,
			}, "/api/workstations/job").
			HasStatusCode(gohttp.StatusAccepted).
			Expect(expectedJob, job, cmpopts.IgnoreFields(service.WorkstationJob{}, "StartTime", "EndTime"))

		event := <-subscribeChan
		assert.Equal(t, river.EventKindJobCompleted, event.Kind)

		expectedJob.State = service.JobStateCompleted

		NewTester(t, server).
			Get(ctx, fmt.Sprintf("/api/workstations/job/%d", job.ID)).
			HasStatusCode(gohttp.StatusOK).
			Expect(expectedJob, job, cmpopts.IgnoreFields(service.WorkstationJob{}, "StartTime", "EndTime"))
	})

	t.Run("Get updated workstation", func(t *testing.T) {
		expected := &service.WorkstationOutput{
			Slug:        slug,
			DisplayName: "User Userson (user.userson@email.com)",
			Reconciling: false,
			StartTime:   nil,
			State:       service.Workstation_STATE_RUNNING,
			Config: &service.WorkstationConfigOutput{
				UpdateTime:     nil,
				IdleTimeout:    2 * time.Hour,
				RunningTimeout: 12 * time.Hour,
				MachineType:    service.MachineTypeN2DStandard32,
				Image:          service.ContainerImageIntellijUltimate,
			},
			Host: workstationHost,
		}

		NewTester(t, server).
			Get(ctx, "/api/workstations/").Debug(os.Stdout).
			HasStatusCode(gohttp.StatusOK).
			Expect(expected, workstation, cmpopts.IgnoreFields(service.WorkstationOutput{}, "CreateTime", "StartTime", "UpdateTime", "Config.CreateTime", "Config.UpdateTime", "Config.Env"))
		assert.NotNil(t, workstation.StartTime)
		assert.Truef(t, maps.Equal(workstation.Config.Env, service.DefaultWorkstationEnv(slug, UserOne.Email, UserOne.Name)), "Expected %v, got %v", map[string]string{"WORKSTATION_NAME": slug}, workstation.Config.Env)
	})

	t.Run("Update URLList", func(t *testing.T) {
		NewTester(t, server).
			Put(ctx, &service.WorkstationURLList{
				URLAllowList:           []string{"github.com/navikt", "github.com/navikt2"},
				DisableGlobalAllowList: true,
			}, "/api/workstations/urllist").Debug(os.Stdout).
			HasStatusCode(gohttp.StatusNoContent)
	})

	t.Run("Get workstation url list", func(t *testing.T) {
		expexted := &service.WorkstationURLList{
			URLAllowList:           []string{"github.com/navikt", "github.com/navikt2"},
			DisableGlobalAllowList: true,
		}

		got := &service.WorkstationURLList{}

		NewTester(t, server).Get(ctx, "/api/workstations/urllist").
			HasStatusCode(gohttp.StatusOK).
			Expect(expexted, got)
	})

	t.Run("Start workstation", func(t *testing.T) {
		expect := &service.WorkstationStartJob{
			JobHeader: service.JobHeader{
				ID:     3,
				State:  service.JobStatePending,
				Errors: []string{},
				Kind:   worker_args.WorkstationStartKind,
			},
			Ident: "v101010",
		}

		subscribeChan, subscribeCancel := workstationWorker.Subscribe(river.EventKindJobCompleted)
		go func() {
			time.Sleep(5 * time.Second)
			subscribeCancel()
		}()

		job := &service.WorkstationStartJob{}

		NewTester(t, server).
			Post(ctx, nil, "/api/workstations/start").
			HasStatusCode(gohttp.StatusAccepted).
			Expect(expect, job, cmpopts.IgnoreFields(service.JobHeader{}, "StartTime", "EndTime"))

		event := <-subscribeChan
		assert.Equal(t, river.EventKindJobCompleted, event.Kind)

		job.State = service.JobStateCompleted
	})

	t.Run("Update workstation onprem mapping hosts", func(t *testing.T) {
		NewTester(t, server).
			Put(ctx, &service.WorkstationOnpremAllowList{Hosts: []string{"host1", "host2", "host3"}}, "/api/workstations/onpremhosts").
			HasStatusCode(gohttp.StatusNoContent)
	})

	t.Run("Get workstation onprem mapping hosts", func(t *testing.T) {
		expected := &service.WorkstationOnpremAllowList{
			Hosts: []string{"host1", "host2", "host3"},
		}

		got := &service.WorkstationOnpremAllowList{}

		NewTester(t, server).
			Get(ctx, "/api/workstations/onpremhosts").
			HasStatusCode(gohttp.StatusOK).
			Expect(expected, got)
	})

	t.Run("Stop workstation", func(t *testing.T) {
		NewTester(t, server).
			Post(ctx, nil, "/api/workstations/stop").
			HasStatusCode(gohttp.StatusNoContent)
	})

	t.Run("Delete workstation", func(t *testing.T) {
		NewTester(t, server).
			Delete(ctx, "/api/workstations/").
			HasStatusCode(gohttp.StatusNoContent)
	})

	t.Run("Start workstation that does not exist", func(t *testing.T) {
		NewTester(t, server).
			Post(ctx, nil, "/api/workstations/start").
			HasStatusCode(gohttp.StatusNotFound)
	})

	t.Run("Stop workstation that does not exist", func(t *testing.T) {
		NewTester(t, server).
			Post(ctx, nil, "/api/workstations/stop").
			HasStatusCode(gohttp.StatusNotFound)
	})

	t.Run("Get workstation that does not exist", func(t *testing.T) {
		NewTester(t, server).
			Get(ctx, "/api/workstations/").
			HasStatusCode(gohttp.StatusNotFound)
		assert.Empty(t, e.GetWorkstationConfigs())
		assert.Empty(t, e.GetWorkstations())
	})

	t.Run("Get workstation jobs for user", func(t *testing.T) {
		expected := &service.WorkstationJobs{
			Jobs: []*service.WorkstationJob{
				{
					JobHeader: service.JobHeader{
						ID:     2,
						State:  service.JobStateCompleted,
						Errors: []string{},
						Kind:   worker_args.WorkstationJobKind,
					},
					Name:           "User Userson",
					Email:          "user.userson@email.com",
					Ident:          "v101010",
					MachineType:    service.MachineTypeN2DStandard32,
					ContainerImage: service.ContainerImageIntellijUltimate,
					Diff: map[string]*service.Diff{
						service.WorkstationDiffMachineType: {
							Added:   []string{"n2d-standard-32"},
							Removed: []string{"n2d-standard-16"},
						},
						service.WorkstationDiffContainerImage: {
							Added:   []string{"europe-north1-docker.pkg.dev/cloud-workstations-images/predefined/intellij-ultimate:latest"},
							Removed: []string{"europe-north1-docker.pkg.dev/cloud-workstations-images/predefined/code-oss:latest"},
						},
					},
				},
				{
					JobHeader: service.JobHeader{
						ID:     1,
						State:  service.JobStateCompleted,
						Errors: []string{},
						Kind:   worker_args.WorkstationJobKind,
					},
					Name:           "User Userson",
					Email:          "user.userson@email.com",
					Ident:          "v101010",
					MachineType:    service.MachineTypeN2DStandard16,
					ContainerImage: service.ContainerImageVSCode,
				},
			},
		}

		got := &service.WorkstationJobs{}

		NewTester(t, server).
			Get(ctx, "/api/workstations/job").
			HasStatusCode(gohttp.StatusOK).
			Expect(expected, got, cmpopts.IgnoreFields(service.JobHeader{}, "StartTime", "EndTime"))
	})

	t.Run("Get workstation start jobs for user", func(t *testing.T) {
		expected := &service.WorkstationStartJobs{
			Jobs: []*service.WorkstationStartJob{
				{
					JobHeader: service.JobHeader{
						ID:     3,
						State:  service.JobStateCompleted,
						Errors: []string{},
						Kind:   worker_args.WorkstationStartKind,
					},
					Ident: "v101010",
				},
			},
		}

		got := &service.WorkstationStartJobs{}

		NewTester(t, server).
			Get(ctx, "/api/workstations/start").
			HasStatusCode(gohttp.StatusOK).
			Expect(expected, got, cmpopts.IgnoreFields(service.JobHeader{}, "StartTime", "EndTime"))
	})

	notAllowedRouter := TestRouter(log)
	{
		h := handlers.NewWorkstationsHandler(workstationService)
		e := routes.NewWorkstationsEndpoints(log, h)
		f := routes.NewWorkstationsRoutes(e, InjectUser(UserTwo))
		f(notAllowedRouter)
	}

	notAllowedServer := httptest.NewServer(notAllowedRouter)

	t.Run("Create workstation not allowed for unauthorized user", func(t *testing.T) {
		NewTester(t, notAllowedServer).
			Post(ctx, service.WorkstationInput{
				MachineType:    service.MachineTypeN2DStandard16,
				ContainerImage: service.ContainerImageVSCode,
			}, "/api/workstations/job").
			HasStatusCode(gohttp.StatusForbidden)
	})
}
