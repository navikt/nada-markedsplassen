package queue

import (
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core/queue/river"
	"riverqueue.com/riverpro"
)

type Queues struct {
	WorkstationsQueue service.WorkstationsQueue
	MetabaseQueue     service.MetabaseQueue
}

func NewQueues(
	riverConfig *riverpro.Config,
	repo *database.Repo,
) *Queues {
	return &Queues{
		WorkstationsQueue: river.NewWorkstationsQueue(riverConfig, repo),
		MetabaseQueue:     river.NewMetabaseQueue(repo, riverConfig),
	}
}
