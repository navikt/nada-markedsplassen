package emulator

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"

	"cloud.google.com/go/iam/apiv1/iampb"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"cloud.google.com/go/workstations/apiv1/workstationspb"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Emulator struct {
	router                 *chi.Mux
	err                    error
	log                    zerolog.Logger
	server                 *httptest.Server
	replicaZones           []string
	storeWorkstationConfig map[string]*workstationspb.WorkstationConfig
	storeWorkstation       map[string]map[string]*workstationspb.Workstation
	storePolicies          map[string]*iampb.Policy
}

func New(log zerolog.Logger) *Emulator {
	e := &Emulator{
		router:                 chi.NewRouter(),
		log:                    log,
		storeWorkstationConfig: make(map[string]*workstationspb.WorkstationConfig),
		storeWorkstation:       make(map[string]map[string]*workstationspb.Workstation),
		storePolicies:          make(map[string]*iampb.Policy),
	}
	e.routes()
	return e
}

func (e *Emulator) SetWorkstationState(fullyQualifiedConfigName, slug string, state workstationspb.Workstation_State) {
	e.storeWorkstation[fullyQualifiedConfigName][slug].State = state

	if state == workstationspb.Workstation_STATE_RUNNING {
		e.storeWorkstation[fullyQualifiedConfigName][slug].StartTime = timestamppb.Now()
	}
}

func (e *Emulator) GetWorkstationConfigs() map[string]*workstationspb.WorkstationConfig {
	return e.storeWorkstationConfig
}

func (e *Emulator) GetWorkstations() map[string]map[string]*workstationspb.Workstation {
	return e.storeWorkstation
}

func (e *Emulator) routes() {
	e.router.With(e.debug).Post("/v1/projects/{project}/locations/{location}/workstationClusters/{cluster}/workstationConfigs", e.createWorkstationConfig)
	e.router.With(e.debug).Get("/v1/projects/{project}/locations/{location}/workstationClusters/{cluster}/workstationConfigs/{configName}", e.getWorkstationConfig)
	e.router.With(e.debug).Patch("/v1/projects/{project}/locations/{location}/workstationClusters/{cluster}/workstationConfigs/{configName}", e.updateWorkstationConfig)
	e.router.With(e.debug).Delete("/v1/projects/{project}/locations/{location}/workstationClusters/{cluster}/workstationConfigs/{configName}", e.deleteWorkstationConfig)
	e.router.With(e.debug).Post("/v1/projects/{project}/locations/{location}/workstationClusters/{cluster}/workstationConfigs/{configName}/workstations", e.createWorkstation)
	e.router.With(e.debug).Get("/v1/projects/{project}/locations/{location}/workstationClusters/{cluster}/workstationConfigs/{configName}/workstations/{name}", e.getWorkstation)
	e.router.With(e.debug).Post("/v1/projects/{project}/locations/{location}/workstationClusters/{cluster}/workstationConfigs/{configName}/workstations/{name}:start", e.startWorkstation)
	e.router.With(e.debug).Post("/v1/projects/{project}/locations/{location}/workstationClusters/{cluster}/workstationConfigs/{configName}/workstations/{name}:stop", e.stopWorkstation)
	e.router.With(e.debug).Get("/v1/projects/{project}/locations/{location}/workstationClusters/{cluster}/workstationConfigs/{configName}/workstations/{name}:getIamPolicy", e.getIamPolicy)
	e.router.With(e.debug).Post("/v1/projects/{project}/locations/{location}/workstationClusters/{cluster}/workstationConfigs/{configName}/workstations/{name}:setIamPolicy", e.setIamPolicy)
	e.router.With(e.debug).Get("/v1/projects/{project}/locations/{location}/operations/{operation}", e.getWorkstationOperation)
	e.router.With(e.debug).NotFound(e.notFound)
}

func (e *Emulator) debug(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request, err := httputil.DumpRequest(r, true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		e.log.Debug().Str("request", string(request)).Msg("request")

		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)

		response, err := httputil.DumpResponse(rec.Result(), true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		e.log.Debug().Str("response", string(response)).Msg("response")

		for k, v := range rec.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(rec.Code)
		w.Write(rec.Body.Bytes())
	})
}

func (e *Emulator) GetIamPolicies() map[string]*iampb.Policy {
	return e.storePolicies
}

func (e *Emulator) setIamPolicy(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil
		return
	}

	projectId, cluster, location, configName, name := chi.URLParam(r, "project"), chi.URLParam(r, "cluster"), chi.URLParam(r, "location"), chi.URLParam(r, "configName"), chi.URLParam(r, "name")

	fullyQualifiedWorkstationName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s/workstation/%s",
		projectId,
		location,
		cluster,
		configName,
		name,
	)

	_, found := e.storePolicies[fullyQualifiedWorkstationName]
	if !found {
		http.Error(w, "not exists", http.StatusNotFound)
		return
	}

	req := &iampb.SetIamPolicyRequest{}
	if err := parseRequest(r, req); err != nil {
		e.log.Error().Err(err).Msg("error parsing request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	e.storePolicies[fullyQualifiedWorkstationName] = req.Policy

	if err := response(w, req.Policy); err != nil {
		e.log.Error().Err(err).Msg("error writing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) getIamPolicy(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil
		return
	}

	projectId, cluster, location, configName, name := chi.URLParam(r, "project"), chi.URLParam(r, "cluster"), chi.URLParam(r, "location"), chi.URLParam(r, "configName"), chi.URLParam(r, "name")

	fullyQualifiedWorkstationName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s/workstation/%s",
		projectId,
		location,
		cluster,
		configName,
		name,
	)

	policy, found := e.storePolicies[fullyQualifiedWorkstationName]
	if !found {
		http.Error(w, "not exists", http.StatusNotFound)
		return
	}

	if err := response(w, policy); err != nil {
		e.log.Error().Err(err).Msg("error writing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) stopWorkstation(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil
		return
	}

	projectId, cluster, location, configName, name := chi.URLParam(r, "project"), chi.URLParam(r, "cluster"), chi.URLParam(r, "location"), chi.URLParam(r, "configName"), chi.URLParam(r, "name")
	fullyQualifiedConfigName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s", projectId, location, cluster, configName)

	req, found := e.storeWorkstation[fullyQualifiedConfigName][name]
	if !found {
		http.Error(w, "not exists", http.StatusNotFound)
		return
	}

	req.State = workstationspb.Workstation_STATE_STOPPED
	req.UpdateTime = timestamppb.Now()

	if err := longRunningResponse(w, req, projectId, location); err != nil {
		e.log.Error().Err(err).Msg("error writing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) SetWorkstationConfigReplicaZones(replicaZones []string) {
	e.replicaZones = replicaZones
}

func (e *Emulator) startWorkstation(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil
		return
	}

	projectId, cluster, location, configName, name := chi.URLParam(r, "project"), chi.URLParam(r, "cluster"), chi.URLParam(r, "location"), chi.URLParam(r, "configName"), chi.URLParam(r, "name")
	fullyQualifiedConfigName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s", projectId, location, cluster, configName)

	req, found := e.storeWorkstation[fullyQualifiedConfigName][name]
	if !found {
		http.Error(w, "not exists", http.StatusNotFound)
		return
	}

	req.State = workstationspb.Workstation_STATE_RUNNING
	req.UpdateTime = timestamppb.Now()

	if err := longRunningResponse(w, req, projectId, location); err != nil {
		e.log.Error().Err(err).Msg("error writing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) createWorkstationConfig(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil
		return
	}

	workstationConfigID := r.URL.Query().Get("workstationConfigId")

	req := &workstationspb.WorkstationConfig{}
	if err := parseRequest(r, req); err != nil {
		e.log.Error().Err(err).Msg("error parsing request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.CreateTime = timestamppb.Now()

	projectId, cluster, location := chi.URLParam(r, "project"), chi.URLParam(r, "cluster"), chi.URLParam(r, "location")
	fullyQualifiedName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s", projectId, location, cluster, workstationConfigID)

	if _, found := e.storeWorkstationConfig[fullyQualifiedName]; found {
		http.Error(w, "already exists", http.StatusConflict)
		return
	}

	e.storeWorkstationConfig[fullyQualifiedName] = req
	e.storeWorkstation[fullyQualifiedName] = make(map[string]*workstationspb.Workstation)

	if err := longRunningResponse(w, req, projectId, location); err != nil {
		e.log.Error().Err(err).Msg("error writing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) getWorkstationConfig(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil
		return
	}

	projectId, cluster, location, configName := chi.URLParam(r, "project"), chi.URLParam(r, "cluster"), chi.URLParam(r, "location"), chi.URLParam(r, "configName")
	fullyQualifiedConfigName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s", projectId, location, cluster, configName)

	req, found := e.storeWorkstationConfig[fullyQualifiedConfigName]
	if !found {
		http.Error(w, "not exists", http.StatusNotFound)
		return
	}

	req.ReplicaZones = e.replicaZones

	if err := response(w, req); err != nil {
		e.log.Error().Err(err).Msg("error writing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) updateWorkstationConfig(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil
		return
	}

	projectId, cluster, location, configName := chi.URLParam(r, "project"), chi.URLParam(r, "cluster"), chi.URLParam(r, "location"), chi.URLParam(r, "configName")
	fullyQualifiedName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s", projectId, location, cluster, configName)

	storedReq, found := e.storeWorkstationConfig[fullyQualifiedName]
	if !found {
		http.Error(w, "not exists", http.StatusNotFound)
		return
	}

	req := &workstationspb.WorkstationConfig{}
	if err := parseRequest(r, req); err != nil {
		e.log.Error().Err(err).Msg("error parsing request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updateMask := strings.Split(r.URL.Query().Get("updateMask"), ",")
	for _, field := range updateMask {
		switch field {
		case "host.gceInstance.machineType":
			storedReq.GetHost().GetGceInstance().MachineType = req.GetHost().GetGceInstance().MachineType
		case "container.image":
			storedReq.GetContainer().Image = req.GetContainer().Image
		case "annotations":
			storedReq.Annotations = req.Annotations
		case "readinessChecks":
			storedReq.ReadinessChecks = req.ReadinessChecks
		case "container.env":
			storedReq.GetContainer().Env = req.GetContainer().Env
		}
	}

	storedReq.UpdateTime = timestamppb.Now()

	for _, w := range e.storeWorkstation[fullyQualifiedName] {
		w.UpdateTime = timestamppb.Now()
	}

	if err := longRunningResponse(w, storedReq, projectId, location); err != nil {
		e.log.Error().Err(err).Msg("error writing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) deleteWorkstationConfig(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil
		return
	}

	projectId, cluster, location, configName := chi.URLParam(r, "project"), chi.URLParam(r, "cluster"), chi.URLParam(r, "location"), chi.URLParam(r, "configName")
	fullyQualifiedName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s", projectId, location, cluster, configName)

	delete(e.storeWorkstationConfig, fullyQualifiedName)
	delete(e.storeWorkstation, fullyQualifiedName)

	if err := longRunningResponse(w, &workstationspb.WorkstationConfig{}, projectId, location); err != nil {
		e.log.Error().Err(err).Msg("error writing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) getWorkstation(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil
		return
	}

	projectId, cluster, location, configName, name := chi.URLParam(r, "project"), chi.URLParam(r, "cluster"), chi.URLParam(r, "location"), chi.URLParam(r, "configName"), chi.URLParam(r, "name")
	fullyQualifiedConfigName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s", projectId, location, cluster, configName)

	req, found := e.storeWorkstation[fullyQualifiedConfigName][name]
	if !found {
		http.Error(w, "not exists", http.StatusNotFound)
		return
	}

	if err := response(w, req); err != nil {
		e.log.Error().Err(err).Msg("error writing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) createWorkstation(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil
		return
	}

	workstationID := r.URL.Query().Get("workstationId")

	req := &workstationspb.Workstation{}
	if err := parseRequest(r, req); err != nil {
		e.log.Error().Err(err).Msg("error parsing request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.State = workstationspb.Workstation_STATE_STARTING
	req.CreateTime = timestamppb.Now()
	req.Host = "ident.workstations.domain"
	req.Reconciling = false

	projectId, cluster, location, configName := chi.URLParam(r, "project"), chi.URLParam(r, "cluster"), chi.URLParam(r, "location"), chi.URLParam(r, "configName")
	fullyQualifiedConfigName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s", projectId, location, cluster, configName)

	if _, found := e.storeWorkstationConfig[fullyQualifiedConfigName]; !found {
		http.Error(w, "not exists", http.StatusNotFound)
		return
	}

	if _, found := e.storeWorkstation[fullyQualifiedConfigName][workstationID]; found {
		http.Error(w, "already exists", http.StatusConflict)
		return
	}

	e.storeWorkstation[fullyQualifiedConfigName][workstationID] = req

	fullyQualifiedWorkstationName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s/workstation/%s",
		projectId,
		location,
		cluster,
		configName,
		workstationID,
	)

	e.storePolicies[fullyQualifiedWorkstationName] = &iampb.Policy{}

	if err := longRunningResponse(w, req, projectId, location); err != nil {
		e.log.Error().Err(err).Msg("error writing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) getWorkstationOperation(w http.ResponseWriter, r *http.Request) {
	projectID, location, operationID := chi.URLParam(r, "project"), chi.URLParam(r, "location"), chi.URLParam(r, "operation")
	op := &longrunningpb.Operation{
		Name: fmt.Sprintf("projects/%s/locations/%s/operations/%s", projectID, location, operationID),
		Done: true,
	}
	if err := response(w, op); err != nil {
		e.log.Error().Err(err).Msg("error writing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) GetRouter() *chi.Mux {
	return e.router
}

func (e *Emulator) Run() string {
	e.log.Info().Msg("starting cloud workstation emulator")
	e.server = httptest.NewServer(e)
	return e.server.URL
}

func (e *Emulator) Reset() {
	e.server.Close()
}

func (e *Emulator) SetError(err error) {
	e.err = err
}

func (e *Emulator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.router.ServeHTTP(w, r)
}

func (e *Emulator) notFound(w http.ResponseWriter, r *http.Request) {
	request, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	e.log.Warn().Str("request", string(request)).Msg("not found")
	http.Error(w, "not found", http.StatusNotFound)
}

func parseRequest(r *http.Request, req proto.Message) error {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	return protojson.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}.Unmarshal(bytes, req)
}

func response(w http.ResponseWriter, v proto.Message) error {
	w.Header().Set("Content-Type", "application/json")

	bytes, err := protojson.Marshal(v)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)

	return err
}

func longRunningResponse(w http.ResponseWriter, msg proto.Message, project, location string) error {
	into := &anypb.Any{}

	if err := anypb.MarshalFrom(into, msg, proto.MarshalOptions{}); err != nil {
		return err
	}

	op := &longrunningpb.Operation{
		Name:   fmt.Sprintf("projects/%s/locations/%s/operations/operation-1737978043998-62cae8bf0a1b4-a3c0e2df-4f80f96b", project, location),
		Done:   true,
		Result: &longrunningpb.Operation_Response{Response: into},
	}

	w.Header().Set("Content-Type", "application/json")

	bytes, err := protojson.Marshal(op)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)

	return err
}
