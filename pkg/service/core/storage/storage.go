package storage

import (
	"github.com/navikt/nada-backend/pkg/config/v2"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service"
	riverstore "github.com/navikt/nada-backend/pkg/service/core/queue/river"
	"github.com/navikt/nada-backend/pkg/service/core/storage/postgres"
	"github.com/rs/zerolog"
	"riverqueue.com/riverpro"
)

type Stores struct {
	AccessStorage             service.AccessStorage
	BigQueryStorage           service.BigQueryStorage
	DataProductsStorage       service.DataProductsStorage
	InsightProductStorage     service.InsightProductStorage
	JoinableViewsStorage      service.JoinableViewsStorage
	KeyWordStorage            service.KeywordsStorage
	RestrictedMetaBaseStorage service.RestrictedMetabaseStorage
	PollyStorage              service.PollyStorage
	ProductAreaStorage        service.ProductAreaStorage
	SearchStorage             service.SearchStorage
	StoryStorage              service.StoryStorage
	TokenStorage              service.TokenStorage
	NaisConsoleStorage        service.NaisConsoleStorage
	WorkstationsQueue         service.WorkstationsQueue
	WorkstationsStorage       service.WorkstationsStorage
	MetabaseDashboardStorage  service.MetabaseDashboardStorage
}

func NewStores(
	riverConfig *riverpro.Config,
	db *database.Repo,
	cfg config.Config,
	log zerolog.Logger,
) *Stores {
	return &Stores{
		AccessStorage:             postgres.NewAccessStorage(db.Querier, database.WithTx[postgres.AccessQueries](db)),
		BigQueryStorage:           postgres.NewBigQueryStorage(db),
		DataProductsStorage:       postgres.NewDataProductStorage(cfg.Metabase.DatabasesBaseURL, db, log),
		InsightProductStorage:     postgres.NewInsightProductStorage(db),
		JoinableViewsStorage:      postgres.NewJoinableViewStorage(db),
		KeyWordStorage:            postgres.NewKeywordsStorage(db),
		RestrictedMetaBaseStorage: postgres.NewMetabaseStorage(db),
		MetabaseDashboardStorage:  postgres.NewMetabaseDashboardStorage(db),
		PollyStorage:              postgres.NewPollyStorage(db),
		ProductAreaStorage:        postgres.NewProductAreaStorage(db),
		SearchStorage:             postgres.NewSearchStorage(db),
		StoryStorage:              postgres.NewStoryStorage(db),
		TokenStorage:              postgres.NewTokenStorage(db),
		NaisConsoleStorage:        postgres.NewNaisConsoleStorage(db),
		WorkstationsQueue:         riverstore.NewWorkstationsQueue(riverConfig, db),
		WorkstationsStorage:       postgres.NewWorkstationsStorage(db),
	}
}
