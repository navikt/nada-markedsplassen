package core

import (
	"context"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.ProductAreaService = &productAreaService{}

type productAreaService struct {
	productAreaStorage    service.ProductAreaStorage
	dataProductStorage    service.DataProductsStorage
	insightProductStorage service.InsightProductStorage
	storyStorage          service.StoryStorage
}

// nolint: gocyclo
func (s *productAreaService) GetProductAreaWithAssets(ctx context.Context, id uuid.UUID) (*service.ProductAreaWithAssets, error) {
	const op errs.Op = "productAreaService.GetProductAreaWithAssets"

	rawProductArea, err := s.productAreaStorage.GetProductArea(ctx, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	dash, err := s.productAreaStorage.GetDashboard(ctx, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	productArea := &service.ProductAreaWithAssets{
		ProductAreaBase: &service.ProductAreaBase{
			TeamkatalogenProductArea: rawProductArea.TeamkatalogenProductArea,
			DashboardURL:             dash.Url,
		},
		TeamsWithAssets: make([]*service.PATeamWithAssets, 0),
	}

	teamIDs := make([]uuid.UUID, len(rawProductArea.PATeams))
	for idx, tkTeam := range rawProductArea.PATeams {
		productArea.TeamsWithAssets = append(productArea.TeamsWithAssets, &service.PATeamWithAssets{
			TeamkatalogenTeam: tkTeam.TeamkatalogenTeam,
			Dataproducts:      []*service.Dataproduct{},
			Stories:           []*service.Story{},
			InsightProducts:   []*service.InsightProduct{},
		})
		teamIDs[idx] = tkTeam.ID

		teamDash, err := s.productAreaStorage.GetDashboard(ctx, teamIDs[idx])
		if err != nil {
			return nil, errs.E(op, err)
		}
		productArea.TeamsWithAssets[idx].DashboardURL = teamDash.Url
	}

	dataproducts, err := s.dataProductStorage.GetDataproductsByTeamID(ctx, teamIDs)
	if err != nil {
		return nil, errs.E(op, err)
	}

	for _, dp := range dataproducts {
		for idx, team := range productArea.TeamsWithAssets {
			if dp.Owner.TeamID != nil && team.ID == *dp.Owner.TeamID {
				productArea.TeamsWithAssets[idx].Dataproducts = append(productArea.TeamsWithAssets[idx].Dataproducts, dp)
			}
		}
	}

	stories, err := s.storyStorage.GetStoriesByTeamID(ctx, teamIDs)
	if err != nil {
		return nil, errs.E(op, err)
	}

	for _, s := range stories {
		for idx, team := range productArea.TeamsWithAssets {
			if s.TeamID != nil && team.ID == *s.TeamID {
				productArea.TeamsWithAssets[idx].Stories = append(productArea.TeamsWithAssets[idx].Stories, s)
			}
		}
	}

	insightProducts, err := s.insightProductStorage.GetInsightProductsByTeamID(ctx, teamIDs)
	if err != nil {
		return nil, errs.E(op, err)
	}

	for _, ip := range insightProducts {
		for idx, team := range productArea.TeamsWithAssets {
			if ip.TeamID != nil && team.ID == *ip.TeamID {
				productArea.TeamsWithAssets[idx].InsightProducts = append(productArea.TeamsWithAssets[idx].InsightProducts, ip)
			}
		}
	}

	return productArea, nil
}

func (s *productAreaService) GetProductAreas(ctx context.Context) (*service.ProductAreasDto, error) {
	const op errs.Op = "productAreaService.GetProductAreas"

	pa, err := s.productAreaStorage.GetProductAreas(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	for _, p := range pa {
		dash, err := s.productAreaStorage.GetDashboard(ctx, p.ID)
		if err != nil {
			return nil, errs.E(op, err)
		}

		p.DashboardURL = dash.Url

		for _, team := range p.PATeams {
			numDataProducts, err := s.dataProductStorage.GetDataproductsNumberByTeam(ctx, team.ID)
			if err != nil {
				return nil, errs.E(op, err)
			}

			numStories, err := s.storyStorage.GetStoriesNumberByTeam(ctx, team.ID)
			if err != nil {
				return nil, errs.E(op, err)
			}

			numInsightProducts, err := s.insightProductStorage.GetInsightProductsNumberByTeam(ctx, team.ID)
			if err != nil {
				return nil, errs.E(op, err)
			}

			team.DataproductsNumber = int(numDataProducts)
			team.StoriesNumber = int(numStories)
			team.InsightProductsNumber = int(numInsightProducts)
		}
	}

	return &service.ProductAreasDto{
		ProductAreas: pa,
	}, nil
}

func NewProductAreaService(
	productAreaStorage service.ProductAreaStorage,
	dataProductStorage service.DataProductsStorage,
	insightProductStorage service.InsightProductStorage,
	storyStorage service.StoryStorage,
) *productAreaService {
	return &productAreaService{
		productAreaStorage:    productAreaStorage,
		dataProductStorage:    dataProductStorage,
		insightProductStorage: insightProductStorage,
		storyStorage:          storyStorage,
	}
}
