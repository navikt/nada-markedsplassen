package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
)

var _ service.UserService = &userService{}

type userService struct {
	knastADGroups         []string
	accessStorage         service.AccessStorage
	pollyStorage          service.PollyStorage
	tokenStorage          service.TokenStorage
	storyStorage          service.StoryStorage
	dataProductStorage    service.DataProductsStorage
	insightProductStorage service.InsightProductStorage
	naisConsoleStorage    service.NaisConsoleStorage
	log                   zerolog.Logger
}

func (s *userService) GetUserData(ctx context.Context, user *service.User) (*service.UserInfo, error) {
	const op = "userService.GetUserData"

	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeUserMissing, op, fmt.Errorf("no user found in context"))
	}

	userData := &service.UserInfo{
		Name:            user.Name,
		Email:           user.Email,
		Ident:           user.Ident,
		IsKnastUser:     user.IsKnastUser,
		GoogleGroups:    user.GoogleGroups,
		LoginExpiration: user.Expiry,
		AllGoogleGroups: user.AllGoogleGroups,
		AzureGroups:     user.AzureGroups,
	}

	for _, grp := range user.GoogleGroups {
		// Skip all-users group
		if grp.Email == "all-users@nav.no" {
			continue
		}

		proj, err := s.naisConsoleStorage.GetTeamProject(ctx, grp.Email)
		if err != nil {
			s.log.Debug().Err(err).Msg("getting team project")

			continue
		}

		userData.GcpProjects = append(userData.GcpProjects, service.GCPProject{
			ID:   proj.ProjectID,
			Name: proj.Slug,
			Group: &service.GoogleGroup{
				Name:  proj.Slug,
				Email: grp.Email,
			},
		})
	}

	teamSlugs := teamNamesFromGroups(userData.GcpProjects)
	tokens, err := s.tokenStorage.GetNadaTokensForTeams(ctx, teamSlugs)
	if err != nil {
		return nil, errs.E(op, errs.Parameter("team"), err)
	}
	userData.NadaTokens = tokens

	dpwithds, dar, err := s.dataProductStorage.GetDataproductsWithDatasetsAndAccessRequests(ctx, []uuid.UUID{}, userData.GoogleGroups.Emails())
	if err != nil {
		return nil, errs.E(op, err)
	}

	for _, dpds := range dpwithds {
		userData.Dataproducts = append(userData.Dataproducts, dpds.Dataproduct)
	}

	for _, ar := range dar {
		ar.Polly, err = s.addPollyDoc(ctx, &ar.AccessRequest)
		if err != nil {
			return nil, errs.E(op, err)
		}
		userData.AccessRequestsAsGranter = append(userData.AccessRequestsAsGranter, ar)
	}

	owned, granted, serviceAccountGranted, err := s.dataProductStorage.GetAccessibleDatasets(ctx, userData.GoogleGroups.Emails(), "user:"+strings.ToLower(user.Email))
	if err != nil {
		return nil, errs.E(op, err)
	}

	userData.Accessable = service.AccessibleDatasets{
		Owned:                 owned,
		Granted:               granted,
		ServiceAccountGranted: serviceAccountGranted,
	}

	dbStories, err := s.storyStorage.GetStoriesWithTeamkatalogenByGroups(ctx, user.GoogleGroups.Emails())
	if err != nil {
		return nil, errs.E(op, err)
	}

	userData.Stories = dbStories

	dbProducts, err := s.insightProductStorage.GetInsightProductsByGroups(ctx, user.GoogleGroups.Emails())
	if err != nil {
		return nil, errs.E(op, err)
	}

	for _, p := range dbProducts {
		userData.InsightProducts = append(userData.InsightProducts, *p)
	}

	groups := []string{strings.ToLower(user.Email)}
	for _, g := range user.GoogleGroups {
		groups = append(groups, strings.ToLower(g.Email))
	}

	accessRequestSQLs, err := s.accessStorage.ListAccessRequestsForOwner(ctx, groups)
	if err != nil {
		if !errs.KindIs(errs.NotExist, err) {
			return nil, errs.E(op, err)
		}
	}

	for _, ar := range accessRequestSQLs {
		ar.Polly, err = s.addPollyDoc(ctx, &ar)
		if err != nil {
			return nil, errs.E(op, err)
		}
		userData.AccessRequests = append(userData.AccessRequests, ar)
	}

	return userData, nil
}

func (s *userService) addPollyDoc(ctx context.Context, ar *service.AccessRequest) (*service.Polly, error) {
	var polly *service.Polly
	var err error
	if ar.Polly != nil {
		polly, err = s.pollyStorage.GetPollyDocumentation(ctx, ar.Polly.ID)
		if err != nil && !errs.KindIs(errs.NotExist, err) {
			return polly, err
		}
	}
	return polly, nil
}

func teamNamesFromGroups(gcpProjects []service.GCPProject) []string {
	teams := make([]string, len(gcpProjects))

	for i, g := range gcpProjects {
		teams[i] = g.Name
	}

	return teams
}

func NewUserService(
	accessStorage service.AccessStorage,
	pollyStorage service.PollyStorage,
	tokenStorage service.TokenStorage,
	storyStorage service.StoryStorage,
	dataProductStorage service.DataProductsStorage,
	insightProductStorage service.InsightProductStorage,
	naisConsoleStorage service.NaisConsoleStorage,
	log zerolog.Logger,
) *userService {
	return &userService{
		accessStorage:         accessStorage,
		pollyStorage:          pollyStorage,
		tokenStorage:          tokenStorage,
		storyStorage:          storyStorage,
		dataProductStorage:    dataProductStorage,
		insightProductStorage: insightProductStorage,
		naisConsoleStorage:    naisConsoleStorage,
		log:                   log,
	}
}
