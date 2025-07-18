package handlers

import (
	"github.com/navikt/nada-backend/pkg/config/v2"
	"github.com/navikt/nada-backend/pkg/service/core"
	"github.com/rs/zerolog"
	"riverqueue.com/riverui"
)

type Handlers struct {
	StoryHandler          *StoryHandler
	TokenHandler          *TokenHandler
	DataProductsHandler   *DataProductsHandler
	MetabaseHandler       *MetabaseHandler
	AccessHandler         *AccessHandler
	ProductAreasHandler   *ProductAreasHandler
	BigQueryHandler       *BigQueryHandler
	SearchHandler         *SearchHandler
	UserHandler           *UserHandler
	SlackHandler          *SlackHandler
	JoinableViewsHandler  *JoinableViewsHandler
	InsightProductHandler *InsightProductHandler
	TeamKatalogenHandler  *TeamkatalogenHandler
	PollyHandler          *PollyHandler
	KeywordsHandler       *KeywordsHandler
	WorkstationsHandler   *WorkstationsHandler
	OnpremMappingHandler  *OnpremMappingHandler
	RiverHandler          *RiverHandler
}

func NewHandlers(
	s *core.Services,
	cfg config.Config,
	server *riverui.Server,
	log zerolog.Logger,
) *Handlers {
	return &Handlers{
		StoryHandler:          NewStoryHandler(cfg.EmailSuffix, s.StoryService, s.TokenService, log),
		TokenHandler:          NewTokenHandler(s.TokenService, cfg.API.AuthToken, log),
		DataProductsHandler:   NewDataProductsHandler(s.DataProductService),
		MetabaseHandler:       NewMetabaseHandler(s.MetaBaseService),
		AccessHandler:         NewAccessHandler(s.AccessService, s.MetaBaseService, cfg.Metabase.GCPProject),
		ProductAreasHandler:   NewProductAreasHandler(s.ProductAreaService),
		BigQueryHandler:       NewBigQueryHandler(s.BigQueryService),
		SearchHandler:         NewSearchHandler(s.SearchService),
		UserHandler:           NewUserHandler(s.UserService),
		SlackHandler:          NewSlackHandler(s.SlackService),
		JoinableViewsHandler:  NewJoinableViewsHandler(s.JoinableViewService),
		InsightProductHandler: NewInsightProductHandler(s.InsightProductService),
		TeamKatalogenHandler:  NewTeamKatalogenHandler(s.TeamKatalogenService),
		PollyHandler:          NewPollyHandler(s.PollyService),
		KeywordsHandler:       NewKeywordsHandler(s.KeyWordService),
		WorkstationsHandler:   NewWorkstationsHandler(s.WorkstationService),
		OnpremMappingHandler:  NewOnpremMappingHandler(s.OnpremMappingService),
		RiverHandler:          NewRiverHandler(server, log),
	}
}
