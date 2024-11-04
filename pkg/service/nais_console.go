package service

import (
	"context"
)

type NaisConsoleStorage interface {
	GetAllTeamProjects(ctx context.Context) (map[string]NaisTeamMapping, error)
	UpdateAllTeamProjects(ctx context.Context, teams map[string]NaisTeamMapping) error
	GetTeamProject(ctx context.Context, googleEmail string) (*NaisTeamMapping, error)
}

type NaisConsoleAPI interface {
	GetGoogleProjectsForAllTeams(ctx context.Context) (map[string]NaisTeamMapping, error)
}

type NaisConsoleService interface {
	UpdateAllTeamProjects(ctx context.Context) error
}

type NaisTeamMapping struct {
	Slug       string
	GroupEmail string
	ProjectID  string
}
