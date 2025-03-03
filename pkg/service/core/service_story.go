package core

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
)

var _ service.StoryService = &storyService{}

type storyService struct {
	storyStorage            service.StoryStorage
	teamKatalogenAPI        service.TeamKatalogenAPI
	cloudStorageAPI         service.CloudStorageAPI
	bucket                  string
	createIgnoreMissingTeam bool

	log zerolog.Logger
}

func (s *storyService) GetIndexHtmlPath(ctx context.Context, prefix string) (string, error) {
	const op = "storyService.GetIndexHtmlPath"

	objs, err := s.cloudStorageAPI.GetObjectsWithPrefix(ctx, s.bucket, prefix)
	if err != nil {
		return "", errs.E(op, err)
	}

	sort.Slice(objs, func(i, j int) bool {
		return len(objs[i].Name) < len(objs[j].Name)
	})

	var candidates []string
	for _, obj := range objs {
		if strings.HasSuffix(strings.ToLower(obj.Name), "/index.html") {
			return obj.Name, nil
		} else if strings.HasSuffix(strings.ToLower(obj.Name), ".html") {
			candidates = append(candidates, obj.Name)
		}
	}

	if len(candidates) == 0 {
		return "", errs.E(errs.NotExist, service.CodeGCPStorage, op, fmt.Errorf("no index.html found in %v", prefix), service.ParamObject)
	}

	return candidates[0], nil
}

func (s *storyService) AppendStoryFiles(ctx context.Context, id uuid.UUID, creatorEmail string, files []*service.UploadFile) error {
	const op = "storyService.AppendStoryFiles"

	story, err := s.storyStorage.GetStory(ctx, id)
	if err != nil {
		return errs.E(op, err)
	}

	if story.Group != creatorEmail {
		return errs.E(errs.Unauthorized, op, errs.UserName(creatorEmail), fmt.Errorf("user %s not in the group of the data story: %s", creatorEmail, story.Group))
	}

	err = s.writeStoryFilesToBucket(ctx, story.ID.String(), files, false)
	if err != nil {
		return errs.E(op, err)
	}

	err = s.storyStorage.UpdateStoryLastModified(ctx, id)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *storyService) RecreateStoryFiles(ctx context.Context, id uuid.UUID, creatorEmail string, files []*service.UploadFile) error {
	const op = "storyService.RecreateStoryFiles"

	story, err := s.storyStorage.GetStory(ctx, id)
	if err != nil {
		return errs.E(op, err)
	}

	if story.Group != creatorEmail {
		return errs.E(errs.Unauthorized, op, errs.UserName(creatorEmail), fmt.Errorf("user %s not in the group of the data story: %s", creatorEmail, story.Group))
	}

	err = s.cloudStorageAPI.DeleteObjectsWithPrefix(ctx, s.bucket, id.String())
	if err != nil {
		return errs.E(op, err)
	}

	err = s.writeStoryFilesToBucket(ctx, story.ID.String(), files, true)
	if err != nil {
		return errs.E(op, err)
	}

	err = s.storyStorage.UpdateStoryLastModified(ctx, id)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *storyService) CreateStoryWithTeamAndProductArea(ctx context.Context, creatorEmail string, newStory *service.NewStory) (*service.Story, error) {
	const op = "storyService.CreateStoryWithTeamAndProductArea"

	if newStory.TeamID != nil {
		teamCatalogURL := s.teamKatalogenAPI.GetTeamCatalogURL(*newStory.TeamID)
		team, err := s.teamKatalogenAPI.GetTeam(ctx, *newStory.TeamID)
		if err != nil {
			if !s.createIgnoreMissingTeam {
				return nil, errs.E(op, err)
			}
		}

		newStory.TeamkatalogenURL = &teamCatalogURL
		if team != nil {
			newStory.ProductAreaID = &team.ProductAreaID
		}
	}

	story, err := s.CreateStory(ctx, creatorEmail, newStory, nil)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return story, nil
}

func (s *storyService) GetObject(ctx context.Context, path string) (*service.ObjectWithData, error) {
	const op = "storyService.GetObject"

	obj, err := s.cloudStorageAPI.GetObject(ctx, s.bucket, path)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return obj, nil
}

func (s *storyService) CreateStory(ctx context.Context, creatorEmail string, newStory *service.NewStory, files []*service.UploadFile) (*service.Story, error) {
	const op = "storyService.CreateStory"

	story, err := s.storyStorage.CreateStory(ctx, creatorEmail, newStory)
	if err != nil {
		return nil, errs.E(op, err)
	}

	err = s.writeStoryFilesToBucket(ctx, story.ID.String(), files, true)
	if err != nil {
		return nil, errs.E(op, err)
	}

	st, err := s.storyStorage.GetStory(ctx, story.ID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return st, nil
}

func (s *storyService) DeleteStory(ctx context.Context, user *service.User, storyID uuid.UUID) (*service.Story, error) {
	const op = "storyService.DeleteStory"

	story, err := s.storyStorage.GetStory(ctx, storyID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	if !user.GoogleGroups.Contains(story.Group) {
		return nil, errs.E(errs.Unauthorized, op, errs.UserName(user.Email), fmt.Errorf("user not in the group of the data story: %s", story.Group))
	}

	err = s.storyStorage.DeleteStory(ctx, storyID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	err = s.cloudStorageAPI.DeleteObjectsWithPrefix(ctx, s.bucket, storyID.String())
	if err != nil {
		return nil, errs.E(op, err)
	}

	return story, nil
}

func (s *storyService) UpdateStory(ctx context.Context, user *service.User, storyID uuid.UUID, input service.UpdateStoryDto) (*service.Story, error) {
	const op = "storyService.UpdateStory"

	existing, err := s.storyStorage.GetStory(ctx, storyID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	if !user.GoogleGroups.Contains(existing.Group) {
		return nil, errs.E(errs.Unauthorized, op, errs.UserName(user.Email), fmt.Errorf("user not in the group of the data story: %s", existing.Group))
	}

	story, err := s.storyStorage.UpdateStory(ctx, storyID, input)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return story, nil
}

func (s *storyService) GetStory(ctx context.Context, storyID uuid.UUID) (*service.Story, error) {
	const op = "storyService.GetStory"

	story, err := s.storyStorage.GetStory(ctx, storyID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return story, nil
}

func (s *storyService) writeStoryFilesToBucket(ctx context.Context, storyID string, files []*service.UploadFile, cleanupOnFailure bool) error {
	const op = "storyService.WriteStoryFilesToBucket"

	var err error
	for _, file := range files {
		err = s.cloudStorageAPI.WriteFileToBucket(ctx, s.bucket, storyID, file)
		if err != nil {
			s.log.Error().Err(err).Msg("writing story file: " + path.Join(storyID, file.Path))
			break
		}
	}
	if err != nil && cleanupOnFailure {
		ed := s.cloudStorageAPI.DeleteObjectsWithPrefix(ctx, s.bucket, storyID)
		if ed != nil {
			s.log.Error().Err(ed).Msg("deleting story folder on cleanup: " + storyID)
		}
	}

	if err != nil {
		return errs.E(errs.IO, service.CodeGCPStorage, op, err)
	}

	return nil
}

func NewStoryService(
	storyStorage service.StoryStorage,
	teamKatalogenAPI service.TeamKatalogenAPI,
	cloudStorageAPI service.CloudStorageAPI,
	createIgnoreMissingTeam bool,
	bucket string,
	log zerolog.Logger,
) *storyService {
	return &storyService{
		storyStorage:            storyStorage,
		teamKatalogenAPI:        teamKatalogenAPI,
		cloudStorageAPI:         cloudStorageAPI,
		createIgnoreMissingTeam: createIgnoreMissingTeam,
		bucket:                  bucket,
		log:                     log,
	}
}
