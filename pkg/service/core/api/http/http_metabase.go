package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/navikt/nada-backend/pkg/errs"

	"github.com/navikt/nada-backend/pkg/service"
)

const (
	metabasePermissionGraphWrite = "write"
	metabasePermissionGraphNone  = "none"
	metabasePermissionGraphRead  = "read"
)

var _ service.MetabaseAPI = &metabaseAPI{}

type metabaseAPI struct {
	c           *http.Client
	password    string
	url         string
	username    string
	expiry      time.Time
	sessionID   string
	disableAuth bool
	endpoint    string
	log         zerolog.Logger
	debug       bool
}

func (c *metabaseAPI) request(ctx context.Context, method, path string, body interface{}, v interface{}) error {
	const op errs.Op = "metabaseAPI.request"

	err := c.ensureValidSession(ctx)
	if err != nil {
		return errs.E(op, err)
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return errs.E(errs.Internal, service.CodeInternalEncoding, op, err, service.ParamExternalRequest)
		}
	}

	res, err := c.performRequest(ctx, method, path, buf)
	if err != nil {
		return errs.E(op, err)
	}

	if c.debug {
		reqdump, _ := httputil.DumpRequestOut(res.Request, true)
		c.log.Info().Msg(string(reqdump))

		respdump, _ := httputil.DumpResponse(res, true)
		c.log.Info().Msg(string(respdump))
	}

	if res.StatusCode == http.StatusNotFound {
		return errs.E(errs.NotExist, service.CodeMetabase, op, err)
	}

	if res.StatusCode > 299 {
		errorMesgBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return errs.E(errs.IO, service.CodeMetabase, op, err)
		}

		c.log.Error().Fields(map[string]any{
			"error_message": string(errorMesgBytes),
			"method":        method,
			"path":          path,
		}).Msg("metabase_request")

		return errs.E(errs.IO, service.CodeMetabase, op, fmt.Errorf("%v %v: non 2xx status code, got: %v", method, path, res.StatusCode))
	}

	if v == nil {
		return nil
	}

	if err := json.NewDecoder(res.Body).Decode(v); err != nil {
		return errs.E(errs.Internal, service.CodeInternalDecoding, op, err, service.ParamExternalResponse)
	}

	return nil
}

func (c *metabaseAPI) performRequest(ctx context.Context, method, path string, buffer io.ReadWriter) (*http.Response, error) {
	const op errs.Op = "metabaseAPI.PerformRequest"

	req, err := http.NewRequestWithContext(ctx, method, c.url+path, buffer)
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeMetabase, op, err)
	}

	req.Header.Set("X-Metabase-Session", c.sessionID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.c.Do(req)
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeMetabase, op, err)
	}

	return resp, nil
}

func (c *metabaseAPI) FindUserByEmail(ctx context.Context, email string) (*service.MetabaseUser, error) {
	const op errs.Op = "metabaseAPI.FindUserByEmail"

	email = strings.ToLower(email)

	users, err := c.GetUsers(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	for _, u := range users {
		if strings.ToLower(u.Email) == email {
			return &u, nil
		}
	}

	return nil, errs.E(errs.NotExist, service.CodeMetabase, op, fmt.Errorf("user %s not found", email), service.ParamUser)
}

func (c *metabaseAPI) GetUsers(ctx context.Context) ([]service.MetabaseUser, error) {
	const op errs.Op = "metabaseAPI.GetUser"

	var users struct {
		Data []service.MetabaseUser
	}

	err := c.request(ctx, http.MethodGet, "/user", nil, &users)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return users.Data, nil
}

func (c *metabaseAPI) CreateUser(ctx context.Context, email string) (*service.MetabaseUser, error) {
	const op errs.Op = "metabaseAPI.CreateUser"

	payload := map[string]string{"email": email}
	var user service.MetabaseUser

	err := c.request(ctx, http.MethodPost, "/user", payload, &user)
	if err != nil {
		return nil, errs.E(op, fmt.Errorf("creating user %s: %w", email, err), service.ParamUser)
	}

	return &user, nil
}

func (c *metabaseAPI) ensureValidSession(ctx context.Context) error {
	const op errs.Op = "metabaseAPI.EnsureValidSession"

	if c.sessionID != "" && c.expiry.After(time.Now()) {
		return nil
	}

	payload := fmt.Sprintf(`{"username": "%s", "password": "%s"}`, c.username, c.password)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/session", strings.NewReader(payload))
	if err != nil {
		return errs.E(errs.IO, service.CodeMetabase, op, fmt.Errorf("creating request: %w", err))
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := c.c.Do(req)
	if err != nil {
		return errs.E(errs.IO, service.CodeMetabase, op, fmt.Errorf("performing request: %w", err))
	}

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return errs.E(errs.IO, service.CodeMetabase, op, fmt.Errorf("not statuscode 200 OK when creating session, got: %v: %v", res.StatusCode, string(b)))
	}

	var session struct {
		ID string `json:"id"`
	}

	err = json.NewDecoder(res.Body).Decode(&session)
	if err != nil {
		return errs.E(errs.IO, service.CodeMetabase, op, err, service.ParamExternalResponse)
	}

	c.sessionID = session.ID
	c.expiry = time.Now().Add(24 * time.Hour)

	return nil
}

func (c *metabaseAPI) Databases(ctx context.Context) ([]service.MetabaseDatabase, error) {
	const op errs.Op = "metabaseAPI.Databases"

	v := struct {
		Data []struct {
			Details struct {
				DatasetID string `json:"dataset-id"`
				ProjectID string `json:"project-id"`
				NadaID    string `json:"nada-id"`
				SAEmail   string `json:"sa-email"`
			} `json:"details"`
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}{}

	if err := c.request(ctx, http.MethodGet, "/database", nil, &v); err != nil {
		return nil, errs.E(op, err)
	}

	var ret []service.MetabaseDatabase
	for _, db := range v.Data {
		ret = append(ret, service.MetabaseDatabase{
			ID:        db.ID,
			Name:      db.Name,
			DatasetID: db.Details.DatasetID,
			ProjectID: db.Details.ProjectID,
			NadaID:    db.Details.NadaID,
			SAEmail:   db.Details.SAEmail,
		})
	}

	return ret, nil
}

func (c *metabaseAPI) Database(ctx context.Context, dbID int) (*service.MetabaseDatabase, error) {
	const op errs.Op = "metabaseAPI.Database"

	v := struct {
		Details struct {
			DatasetID string `json:"dataset-filters-patterns"`
			ProjectID string `json:"project-id"`
			NadaID    string `json:"nada-id"`
			SAEmail   string `json:"sa-email"`
		} `json:"details"`
		ID   int    `json:"id"`
		Name string `json:"name"`
	}{}

	if err := c.request(ctx, http.MethodGet, fmt.Sprintf("/database/%d", dbID), nil, &v); err != nil {
		return nil, errs.E(op, err)
	}

	return &service.MetabaseDatabase{
		ID:        v.ID,
		Name:      v.Name,
		DatasetID: v.Details.DatasetID,
		ProjectID: v.Details.ProjectID,
		NadaID:    v.Details.NadaID,
		SAEmail:   v.Details.SAEmail,
	}, nil
}

type NewDatabase struct {
	AutoRunQueries bool    `json:"auto_run_queries"`
	Details        Details `json:"details"`
	Engine         string  `json:"engine"`
	IsFullSync     bool    `json:"is_full_sync"`
	Name           string  `json:"name"`
}

type UpdateDatabase struct {
	AutoRunQueries bool    `json:"auto_run_queries"`
	Details        Details `json:"details"`
	Engine         string  `json:"engine"`
	IsFullSync     bool    `json:"is_full_sync"`
	IsOnDemand     bool    `json:"is_on_demand"`
	Name           string  `json:"name"`
}

type Details struct {
	DatasetID          string `json:"dataset-id"`
	ProjectID          string `json:"project-id"`
	ServiceAccountJSON string `json:"service-account-json"`
	NadaID             string `json:"nada-id"`
	SAEmail            string `json:"sa-email"`
	Endpoint           string `json:"endpoint,omitempty"`
	EnableAuth         *bool  `json:"enable-auth,omitempty"`
}

func (c *metabaseAPI) CreateDatabase(ctx context.Context, team, name, saJSON, saEmail string, ds *service.BigQuery) (int, error) {
	const op errs.Op = "metabaseAPI.CreateDatabase"

	dbs, err := c.Databases(ctx)
	if err != nil {
		return 0, errs.E(op, err)
	}

	if dbID, exists := dbExists(dbs, ds.DatasetID.String()); exists {
		return dbID, nil
	}

	var enableAuth *bool = nil
	if c.disableAuth {
		enableAuth = new(bool) // false
	}

	db := NewDatabase{
		Name: strings.Split(team, "@")[0] + ": " + name,
		Details: Details{
			DatasetID:          ds.Dataset,
			ProjectID:          ds.ProjectID,
			ServiceAccountJSON: saJSON,
			NadaID:             ds.DatasetID.String(),
			SAEmail:            saEmail,
			Endpoint:           c.endpoint,
			EnableAuth:         enableAuth,
		},
		Engine:         "bigquery-cloud-sdk",
		IsFullSync:     true,
		AutoRunQueries: true,
	}
	var v struct {
		ID int `json:"id"`
	}
	err = c.request(ctx, http.MethodPost, "/database", db, &v)
	if err != nil {
		c.log.Debug().Fields(map[string]any{
			"team":        team,
			"name":        name,
			"sa":          saEmail,
			"endpoint":    c.endpoint,
			"enable_auth": enableAuth,
		}).Msg("creating_database")
		return 0, errs.E(op, err)
	}

	return v.ID, nil
}

func (c *metabaseAPI) UpdateDatabase(ctx context.Context, dbID int, saJSON, saEmail string) error {
	const op errs.Op = "metabaseAPI.UpdateDatabase"

	existing, err := c.Database(ctx, dbID)
	if err != nil {
		return errs.E(op, err)
	}

	db := UpdateDatabase{
		Name: existing.Name,
		Details: Details{
			DatasetID:          existing.DatasetID,
			ProjectID:          existing.ProjectID,
			NadaID:             existing.NadaID,
			ServiceAccountJSON: saJSON,
			SAEmail:            saEmail,
			Endpoint:           c.endpoint,
		},
		Engine:         "bigquery-cloud-sdk",
		IsFullSync:     true,
		AutoRunQueries: true,
		IsOnDemand:     false,
	}

	err = c.request(ctx, http.MethodPut, fmt.Sprintf("/database/%d", dbID), db, nil)
	if err != nil {
		c.log.Debug().Fields(map[string]any{
			"name":     db.Name,
			"sa":       saEmail,
			"endpoint": c.endpoint,
		}).Msg("updating_database")
		return errs.E(op, err)
	}

	return nil
}

func (c *metabaseAPI) HideTables(ctx context.Context, ids []int) error {
	const op errs.Op = "metabaseAPI.HideTables"

	t := struct {
		IDs            []int  `json:"ids"`
		VisibilityType string `json:"visibility_type"`
	}{
		IDs:            ids,
		VisibilityType: "hidden",
	}

	err := c.request(ctx, http.MethodPut, "/table", t, nil)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (c *metabaseAPI) ShowTables(ctx context.Context, ids []int) error {
	const op errs.Op = "metabaseAPI.ShowTables"

	t := struct {
		IDs            []int   `json:"ids"`
		VisibilityType *string `json:"visibility_type"`
	}{
		IDs:            ids,
		VisibilityType: nil,
	}

	err := c.request(ctx, http.MethodPut, "/table", t, nil)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (c *metabaseAPI) Tables(ctx context.Context, dbID int, includeHidden bool) ([]service.MetabaseTable, error) {
	const op errs.Op = "metabaseAPI.Tables"

	var v struct {
		Tables []service.MetabaseTable `json:"tables"`
	}

	url := fmt.Sprintf("/database/%d/metadata", dbID)
	if includeHidden {
		url += "?include_hidden=true"
	}

	if err := c.request(ctx, http.MethodGet, url, nil, &v); err != nil {
		return nil, errs.E(op, err)
	}

	var ret []service.MetabaseTable
	for _, t := range v.Tables {
		ret = append(ret, service.MetabaseTable{
			Name:   t.Name,
			ID:     t.ID,
			Fields: t.Fields,
		})
	}

	return ret, nil
}

func (c *metabaseAPI) DeleteDatabase(ctx context.Context, id int) error {
	const op errs.Op = "metabaseAPI.DeleteDatabase"

	err := c.request(ctx, http.MethodDelete, fmt.Sprintf("/database/%d", id), nil, nil)
	if err != nil {
		if errs.KindIs(errs.NotExist, err) {
			return nil
		}

		return errs.E(op, err)
	}

	return nil
}

func (c *metabaseAPI) AutoMapSemanticTypes(ctx context.Context, dbID int) error {
	const op errs.Op = "metabaseAPI.AutoMapSemanticTypes"

	tables, err := c.Tables(ctx, dbID, false)
	if err != nil {
		return errs.E(op, err)
	}

	for _, t := range tables {
		for _, f := range t.Fields {
			switch f.DatabaseType {
			case "STRING":
				if err := c.MapSemanticType(ctx, f.ID, "type/Name"); err != nil {
					return err
				}
			case "TIMESTAMP":
				if err := c.MapSemanticType(ctx, f.ID, "type/CreationTimestamp"); err != nil {
					return err
				}
			case "DATE":
				if err := c.MapSemanticType(ctx, f.ID, "type/CreationDate"); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *metabaseAPI) MapSemanticType(ctx context.Context, fieldID int, semanticType string) error {
	const op errs.Op = "metabaseAPI.MapSemanticType"

	payload := map[string]string{"semantic_type": semanticType}
	err := c.request(ctx, http.MethodPut, "/field/"+strconv.Itoa(fieldID), payload, nil)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (c *metabaseAPI) GetPermissionGroups(ctx context.Context) ([]service.MetabasePermissionGroup, error) {
	const op errs.Op = "metabaseAPI.GetPermissionGroups"

	groups := []service.MetabasePermissionGroup{}

	err := c.request(ctx, http.MethodGet, "/permissions/group", nil, &groups)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return groups, nil
}

func (c *metabaseAPI) GetOrCreatePermissionGroup(ctx context.Context, name string) (int, error) {
	const op errs.Op = "metabaseAPI.GetOrCreatePermissionGroup"

	groups, err := c.GetPermissionGroups(ctx)
	if err != nil {
		return 0, errs.E(op, err)
	}

	for _, g := range groups {
		if g.Name == name {
			return g.ID, nil
		}
	}

	gid, err := c.CreatePermissionGroup(ctx, name)
	if err != nil {
		return 0, errs.E(op, err)
	}

	return gid, nil
}

func (c *metabaseAPI) CreatePermissionGroup(ctx context.Context, name string) (int, error) {
	const op errs.Op = "metabaseAPI.CreatePermissionGroup"

	group := service.MetabasePermissionGroup{}
	payload := map[string]string{"name": name}
	if err := c.request(ctx, http.MethodPost, "/permissions/group", payload, &group); err != nil {
		return 0, errs.E(op, fmt.Errorf("creating group '%s': %w", name, err))
	}

	return group.ID, nil
}

func (c *metabaseAPI) GetPermissionGroup(ctx context.Context, groupID int) ([]service.MetabasePermissionGroupMember, error) {
	const op errs.Op = "metabaseAPI.GetPermissionGroup"

	g := service.MetabasePermissionGroup{}
	err := c.request(ctx, http.MethodGet, fmt.Sprintf("/permissions/group/%d", groupID), nil, &g)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return g.Members, nil
}

func (c *metabaseAPI) RemovePermissionGroupMember(ctx context.Context, memberID int) error {
	const op errs.Op = "metabaseAPI.RemovePermissionGroupMember"

	err := c.request(ctx, http.MethodDelete, fmt.Sprintf("/permissions/membership/%d", memberID), nil, nil)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (c *metabaseAPI) AddPermissionGroupMember(ctx context.Context, groupID, userID int) error {
	const op errs.Op = "metabaseAPI.AddPermissionGroupMember"

	payload := map[string]int{"group_id": groupID, "user_id": userID}
	err := c.request(ctx, http.MethodPost, "/permissions/membership", payload, nil)
	if err != nil {
		return errs.E(op, fmt.Errorf("creating group %d for user %d: %w", groupID, userID, err))
	}

	return nil
}

func (c *metabaseAPI) GetPermissionGraphForGroup(ctx context.Context, groupID int) (*service.PermissionGraphGroups, error) {
	const op errs.Op = "metabaseAPI.GetPermissionGraphForGroup"

	permissionGraphGroup := service.PermissionGraphGroups{}
	err := c.request(ctx, http.MethodGet, fmt.Sprintf("/permissions/graph/group/%d", groupID), nil, &permissionGraphGroup)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &permissionGraphGroup, nil
}

func (c *metabaseAPI) RestrictAccessToDatabase(ctx context.Context, groupID int, databaseID int) error {
	const op errs.Op = "metabaseAPI.RestrictAccessToDatabase"

	var permissionGraph struct {
		Groups   map[string]map[string]service.PermissionGroup `json:"groups"`
		Revision int                                           `json:"revision"`
	}

	err := c.request(ctx, http.MethodGet, fmt.Sprintf("/permissions/graph/group/%d", groupID), nil, &permissionGraph)
	if err != nil {
		return errs.E(op, err)
	}

	_, hasGroup := permissionGraph.Groups[strconv.Itoa(groupID)]
	if !hasGroup {
		return errs.E(errs.IO, op, fmt.Errorf("group %d not found in permission graph", groupID))
	}

	if groupID != service.MetabaseAllUsersGroupID {
		// When adding a new restricted database the corresponding permission group should not have any existing permissions.
		// Therefore we remove all existing permissions in the permission graph for this group
		for dbID := range permissionGraph.Groups[strconv.Itoa(groupID)] {
			permissionGraph.Groups[strconv.Itoa(groupID)][dbID] = service.PermissionGroup{
				ViewData:      "unrestricted",
				CreateQueries: "no",
				DataModel:     &service.DataModelPermission{Schemas: "none"},
				Download:      &service.DownloadPermission{Schemas: "none"},
				Details:       "no",
			}
		}
	}

	permissionGraph.Groups[strconv.Itoa(groupID)][strconv.Itoa(databaseID)] = service.PermissionGroup{
		ViewData:      "unrestricted",
		CreateQueries: "query-builder-and-native",
		DataModel:     &service.DataModelPermission{Schemas: "all"},
		Download:      &service.DownloadPermission{Schemas: "full"},
		Details:       "no",
	}

	if err := c.request(ctx, http.MethodPut, "/permissions/graph", permissionGraph, nil); err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (c *metabaseAPI) OpenAccessToDatabase(ctx context.Context, databaseID int) error {
	const op errs.Op = "metabaseAPI.OpenAccessToDatabase"

	err := c.RestrictAccessToDatabase(ctx, service.MetabaseAllUsersGroupID, databaseID)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (c *metabaseAPI) DeletePermissionGroup(ctx context.Context, groupID int) error {
	const op errs.Op = "metabaseAPI.DeletePermissionGroup"

	if groupID <= 0 {
		return nil
	}

	err := c.request(ctx, http.MethodDelete, fmt.Sprintf("/permissions/group/%d", groupID), nil, nil)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (c *metabaseAPI) ArchiveCollection(ctx context.Context, colID int) error {
	const op errs.Op = "metabaseAPI.ArchiveCollection"

	var collection struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		ID          int    `json:"id"`
		Archived    bool   `json:"archived"`
	}

	if err := c.request(ctx, http.MethodGet, "/collection/"+strconv.Itoa(colID), nil, &collection); err != nil {
		return errs.E(op, err)
	}

	collection.Archived = true

	if err := c.request(ctx, http.MethodPut, "/collection/"+strconv.Itoa(colID), collection, nil); err != nil {
		return errs.E(op, err)
	}

	return nil
}

type CollectionID struct {
	IntID    int
	StringID string
	IsString bool
}

func (c *CollectionID) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		c.IsString = true

		return json.Unmarshal(data, &c.StringID)
	}

	return json.Unmarshal(data, &c.IntID)
}

type Collection struct {
	ID          CollectionID `json:"id,omitempty"`
	Name        string       `json:"name,omitempty"`
	Description string       `json:"description,omitempty"`
	IsPersonal  bool         `json:"is_personal,omitempty"`
	IsSample    bool         `json:"is_sample,omitempty"`
}

func (c *metabaseAPI) GetCollections(ctx context.Context) ([]*service.MetabaseCollection, error) {
	const op errs.Op = "metabaseAPI.GetCollections"

	var raw []Collection

	if err := c.request(ctx, http.MethodGet, "/collection/", nil, &raw); err != nil {
		return nil, errs.E(op, err)
	}

	var collections []*service.MetabaseCollection
	for _, col := range raw {
		if col.ID.IsString {
			c.log.Debug().Msgf("collection id is string: %s, skipping", col.ID.StringID)

			continue
		}

		if col.IsPersonal || col.IsSample {
			c.log.Debug().Msgf("skipping personal or sample collection: %s", col.Name)

			continue
		}

		collections = append(collections, &service.MetabaseCollection{
			ID:          col.ID.IntID,
			Name:        col.Name,
			Description: col.Description,
		})
	}

	return collections, nil
}

func (c *metabaseAPI) UpdateCollection(ctx context.Context, collection *service.MetabaseCollection) error {
	const op errs.Op = "metabaseAPI.UpdateCollection"

	col := Collection{
		Name:        collection.Name,
		Description: collection.Description,
	}

	err := c.request(ctx, http.MethodPut, fmt.Sprintf("/collection/%d", collection.ID), col, nil)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (c *metabaseAPI) CreateCollection(ctx context.Context, name string) (int, error) {
	const op errs.Op = "metabaseAPI.CreateCollection"

	collection := struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}{
		Name:        name,
		Description: "Collection for " + name,
		Color:       "#509EE3",
	}

	var response struct {
		ID int `json:"id"`
	}

	err := c.request(ctx, http.MethodPost, "/collection", collection, &response)
	if err != nil {
		return 0, errs.E(op, err)
	}

	return response.ID, nil
}

func (c *metabaseAPI) SetCollectionAccess(ctx context.Context, groupID int, collectionID int, removeAllUsersAccess bool) error {
	const op errs.Op = "metabaseAPI.SetCollectionAccess"

	var cPermissions struct {
		Revision int                          `json:"revision"`
		Groups   map[string]map[string]string `json:"groups"`
	}

	err := c.request(ctx, http.MethodGet, "/collection/graph", nil, &cPermissions)
	if err != nil {
		return errs.E(op, err)
	}

	group, hasGroup := cPermissions.Groups[strconv.Itoa(groupID)]
	if !hasGroup {
		return errs.E(errs.IO, service.CodeMetabase, op, fmt.Errorf("group %d not found in permission graph for collections", groupID))
	}

	_, hasCollection := group[strconv.Itoa(collectionID)]
	if !hasCollection {
		return errs.E(errs.IO, service.CodeMetabase, op, fmt.Errorf("collection %d not found in permission graph for group %d", collectionID, groupID))
	}

	cPermissions.Groups[strconv.Itoa(groupID)][strconv.Itoa(collectionID)] = metabasePermissionGraphWrite

	if removeAllUsersAccess {
		cPermissions.Groups[strconv.Itoa(service.MetabaseAllUsersGroupID)][strconv.Itoa(collectionID)] = metabasePermissionGraphNone
	}

	err = c.request(ctx, http.MethodPut, "/collection/graph", cPermissions, nil)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (c *metabaseAPI) CreateCollectionWithAccess(ctx context.Context, groupID int, name string, removeAllUsersAccess bool) (int, error) {
	const op errs.Op = "metabaseAPI.CreateCollectionWithAccess"

	cid, err := c.CreateCollection(ctx, name)
	if err != nil {
		return 0, errs.E(op, err)
	}

	if err := c.SetCollectionAccess(ctx, groupID, cid, removeAllUsersAccess); err != nil {
		return cid, errs.E(op, err)
	}

	return cid, nil
}

func dbExists(dbs []service.MetabaseDatabase, nadaID string) (int, bool) {
	for _, db := range dbs {
		if db.NadaID == nadaID {
			return db.ID, true
		}
	}

	return 0, false
}

func NewMetabaseHTTP(url, username, password, endpoint string, disableAuth, debug bool, log zerolog.Logger) *metabaseAPI {
	return &metabaseAPI{
		c: &http.Client{
			Timeout: time.Second * 300, //nolint:gomnd
		},
		url:         url,
		password:    password,
		username:    username,
		endpoint:    endpoint,
		disableAuth: disableAuth,
		log:         log,
		debug:       debug,
	}
}
