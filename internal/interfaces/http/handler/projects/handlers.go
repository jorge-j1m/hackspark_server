package projects

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jorge-j1m/hackspark_server/ent"
	"github.com/jorge-j1m/hackspark_server/ent/like"
	"github.com/jorge-j1m/hackspark_server/ent/project"
	"github.com/jorge-j1m/hackspark_server/ent/projecttag"
	"github.com/jorge-j1m/hackspark_server/ent/tag"
	"github.com/jorge-j1m/hackspark_server/ent/user"
	log "github.com/jorge-j1m/hackspark_server/internal/infrastructure/logger"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/middleware"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/response"
	"github.com/jorge-j1m/hackspark_server/internal/pkg/common/errors"
)

type CreateProjectRequest struct {
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	IsPublic    *bool    `json:"is_public"`
	Tags        []string `json:"tags"`
}

func (r CreateProjectRequest) Validate() error {
	if r.Name == "" {
		return errors.ErrInvalidRequest
	}
	if len(r.Name) > 255 {
		return errors.ErrInvalidRequest
	}
	if r.Description != nil && len(*r.Description) > 1000 {
		return errors.ErrInvalidRequest
	}
	return nil
}

type UpdateProjectRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	IsPublic    *bool    `json:"is_public"`
	Tags        []string `json:"tags"`
}

func (r UpdateProjectRequest) Validate() error {
	if r.Name != nil && (*r.Name == "" || len(*r.Name) > 255) {
		return errors.ErrInvalidRequest
	}
	if r.Description != nil && len(*r.Description) > 1000 {
		return errors.ErrInvalidRequest
	}
	return nil
}

type ProjectResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	IsPublic    bool     `json:"is_public"`
	LikeCount   int      `json:"like_count"`
	StarCount   int      `json:"star_count"`
	Tags        []string `json:"tags"`
	Owner       *struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"owner"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (h *ProjectsHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get user ID from context")
		response.Error(w, errors.ErrUserNotFound)
		return
	}

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error(ctx).Err(err).Msg("Failed to decode request body")
		response.Error(w, errors.ErrInvalidRequest)
		return
	}

	if err := req.Validate(); err != nil {
		log.Error(ctx).Err(err).Msg("Invalid request data")
		response.Error(w, errors.ErrInvalidRequest)
		return
	}

	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	project, err := h.client.Project.Create().
		SetName(req.Name).
		SetNillableDescription(req.Description).
		SetIsPublic(isPublic).
		SetOwnerID(userID).
		Save(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to create project")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	if len(req.Tags) > 0 {
		if err := h.addTagsToProject(ctx, project.ID, req.Tags); err != nil {
			log.Error(ctx).Err(err).Msg("Failed to add tags to project")
		}
	}

	projectResp, err := h.getProjectResponse(ctx, project.ID)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get project response")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	log.Info(ctx).Msgf("Project created successfully: %s", project.ID)
	response.JSON(w, http.StatusCreated, "Project created successfully", projectResp)
}

func (h *ProjectsHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectID := chi.URLParam(r, "id")

	projectResp, err := h.getProjectResponse(ctx, projectID)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Error(ctx).Err(err).Msg("Project not found")
			response.Error(w, errors.ErrNotFound)
			return
		}
		log.Error(ctx).Err(err).Msg("Failed to get project")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	response.JSON(w, http.StatusOK, "Project retrieved successfully", projectResp)
}

func (h *ProjectsHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectID := chi.URLParam(r, "id")
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get user ID from context")
		response.Error(w, errors.ErrUserNotFound)
		return
	}

	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error(ctx).Err(err).Msg("Failed to decode request body")
		response.Error(w, errors.ErrInvalidRequest)
		return
	}

	if err := req.Validate(); err != nil {
		log.Error(ctx).Err(err).Msg("Invalid request data")
		response.Error(w, errors.ErrInvalidRequest)
		return
	}

	project, err := h.client.Project.Query().
		Where(project.ID(projectID)).
		WithOwner().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Error(ctx).Err(err).Msg("Project not found")
			response.Error(w, errors.ErrNotFound)
			return
		}
		log.Error(ctx).Err(err).Msg("Failed to get project")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	if project.Edges.Owner.ID != userID {
		log.Error(ctx).Msg("User is not the owner of the project")
		response.Error(w, errors.ErrForbidden)
		return
	}

	updateQuery := h.client.Project.UpdateOneID(projectID)
	if req.Name != nil {
		updateQuery = updateQuery.SetName(*req.Name)
	}
	if req.Description != nil {
		updateQuery = updateQuery.SetNillableDescription(req.Description)
	}
	if req.IsPublic != nil {
		updateQuery = updateQuery.SetIsPublic(*req.IsPublic)
	}

	if _, err := updateQuery.Save(ctx); err != nil {
		log.Error(ctx).Err(err).Msg("Failed to update project")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	if req.Tags != nil {
		if err := h.replaceProjectTags(ctx, projectID, req.Tags); err != nil {
			log.Error(ctx).Err(err).Msg("Failed to update project tags")
		}
	}

	projectResp, err := h.getProjectResponse(ctx, projectID)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get project response")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	log.Info(ctx).Msgf("Project updated successfully: %s", projectID)
	response.JSON(w, http.StatusOK, "Project updated successfully", projectResp)
}

func (h *ProjectsHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectID := chi.URLParam(r, "id")
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get user ID from context")
		response.Error(w, errors.ErrUserNotFound)
		return
	}

	project, err := h.client.Project.Query().
		Where(project.ID(projectID)).
		WithOwner().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Error(ctx).Err(err).Msg("Project not found")
			response.Error(w, errors.ErrNotFound)
			return
		}
		log.Error(ctx).Err(err).Msg("Failed to get project")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	if project.Edges.Owner.ID != userID {
		log.Error(ctx).Msg("User is not the owner of the project")
		response.Error(w, errors.ErrForbidden)
		return
	}

	if err := h.client.Project.DeleteOneID(projectID).Exec(ctx); err != nil {
		log.Error(ctx).Err(err).Msg("Failed to delete project")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	log.Info(ctx).Msgf("Project deleted successfully: %s", projectID)
	response.JSON(w, http.StatusOK, "Project deleted successfully", nil)
}

func (h *ProjectsHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit, offset := h.getPagination(r)

	query := h.client.Project.Query().
		Where(project.IsPublic(true)).
		WithOwner().
		WithTags().
		Limit(limit).
		Offset(offset).
		Order(ent.Desc(project.FieldCreateTime))

	if tagsParam := r.URL.Query().Get("tags"); tagsParam != "" {
		tagSlugs := strings.Split(tagsParam, ",")
		query = query.Where(project.HasTagsWith(tag.SlugIn(tagSlugs...)))
	}

	if ownerParam := r.URL.Query().Get("owner"); ownerParam != "" {
		query = query.Where(project.HasOwnerWith(user.Username(ownerParam)))
	}

	projects, err := query.All(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to list projects")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	var projectResponses []ProjectResponse
	for _, p := range projects {
		resp := h.buildProjectResponse(p)
		projectResponses = append(projectResponses, resp)
	}

	response.JSON(w, http.StatusOK, "Projects retrieved successfully", projectResponses)
}

func (h *ProjectsHandler) getProjectResponse(ctx context.Context, projectID string) (*ProjectResponse, error) {
	project, err := h.client.Project.Query().
		Where(project.ID(projectID)).
		WithOwner().
		WithTags().
		First(ctx)
	if err != nil {
		return nil, err
	}

	resp := h.buildProjectResponse(project)
	return &resp, nil
}

func (h *ProjectsHandler) buildProjectResponse(p *ent.Project) ProjectResponse {
	resp := ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		IsPublic:    p.IsPublic,
		LikeCount:   p.LikeCount,
		StarCount:   p.StarCount,
		CreatedAt:   p.CreateTime.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   p.UpdateTime.Format("2006-01-02T15:04:05Z"),
	}

	if p.Edges.Owner != nil {
		resp.Owner = &struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		}{
			ID:       p.Edges.Owner.ID,
			Username: p.Edges.Owner.Username,
		}
	}

	if p.Edges.Tags != nil {
		for _, t := range p.Edges.Tags {
			resp.Tags = append(resp.Tags, t.Slug)
		}
	}

	return resp
}

func (h *ProjectsHandler) getPagination(r *http.Request) (limit, offset int) {
	limit = 20
	offset = 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	return limit, offset
}

func (h *ProjectsHandler) addTagsToProject(ctx context.Context, projectID string, tagSlugs []string) error {
	for _, slug := range tagSlugs {
		normalizedSlug := h.normalizeSlug(slug)

		tag, err := h.client.Tag.Query().Where(tag.Slug(normalizedSlug)).First(ctx)
		if ent.IsNotFound(err) {
			tag, err = h.client.Tag.Create().
				SetName(slug).
				SetSlug(normalizedSlug).
				Save(ctx)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		_, err = h.client.ProjectTag.Create().
			SetProjectID(projectID).
			SetTagID(tag.ID).
			Save(ctx)
		if err != nil {
			log.Error(ctx).Err(err).Msgf("Failed to add tag %s to project %s", slug, projectID)
		}
	}
	return nil
}

func (h *ProjectsHandler) replaceProjectTags(ctx context.Context, projectID string, tagSlugs []string) error {
	if _, err := h.client.ProjectTag.Delete().Where(
		projecttag.ProjectID(projectID),
	).Exec(ctx); err != nil {
		return err
	}

	return h.addTagsToProject(ctx, projectID, tagSlugs)
}

func (h *ProjectsHandler) normalizeSlug(input string) string {
	slug := strings.ToLower(input)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, ".", "")
	slug = strings.ReplaceAll(slug, "/", "")
	return slug
}

func (h *ProjectsHandler) LikeProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectID := chi.URLParam(r, "id")
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get user ID from context")
		response.Error(w, errors.ErrUserNotFound)
		return
	}

	_, err = h.client.Project.Query().
		Where(project.ID(projectID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Error(ctx).Err(err).Msg("Project not found")
			response.Error(w, errors.ErrNotFound)
			return
		}
		log.Error(ctx).Err(err).Msg("Failed to get project")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	existingLike, err := h.client.Like.Query().
		Where(like.UserID(userID), like.ProjectID(projectID)).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		log.Error(ctx).Err(err).Msg("Failed to check existing like")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	if existingLike != nil {
		response.JSON(w, http.StatusOK, "Project already liked", nil)
		return
	}

	_, err = h.client.Like.Create().
		SetUserID(userID).
		SetProjectID(projectID).
		Save(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to create like")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	if _, err := h.client.Project.UpdateOneID(projectID).
		AddLikeCount(1).
		Save(ctx); err != nil {
		log.Error(ctx).Err(err).Msg("Failed to increment like count")
	}

	log.Info(ctx).Msgf("Project liked successfully: %s by user %s", projectID, userID)
	response.JSON(w, http.StatusOK, "Project liked successfully", nil)
}

func (h *ProjectsHandler) UnlikeProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectID := chi.URLParam(r, "id")
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get user ID from context")
		response.Error(w, errors.ErrUserNotFound)
		return
	}

	existingLike, err := h.client.Like.Query().
		Where(like.UserID(userID), like.ProjectID(projectID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Error(ctx).Err(err).Msg("Like not found")
			response.Error(w, errors.ErrNotFound)
			return
		}
		log.Error(ctx).Err(err).Msg("Failed to check existing like")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	if err := h.client.Like.DeleteOneID(existingLike.ID).Exec(ctx); err != nil {
		log.Error(ctx).Err(err).Msg("Failed to delete like")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	if _, err := h.client.Project.UpdateOneID(projectID).
		AddLikeCount(-1).
		Save(ctx); err != nil {
		log.Error(ctx).Err(err).Msg("Failed to decrement like count")
	}

	log.Info(ctx).Msgf("Project unliked successfully: %s by user %s", projectID, userID)
	response.JSON(w, http.StatusOK, "Project unliked successfully", nil)
}

func (h *ProjectsHandler) GetProjectLikes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectID := chi.URLParam(r, "id")

	limit, offset := h.getPagination(r)

	likes, err := h.client.Like.Query().
		Where(like.ProjectID(projectID)).
		WithUser().
		Limit(limit).
		Offset(offset).
		Order(ent.Desc(like.FieldCreateTime)).
		All(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get project likes")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	type LikeResponse struct {
		UserID    string `json:"user_id"`
		Username  string `json:"username"`
		CreatedAt string `json:"created_at"`
	}

	var likeResponses []LikeResponse
	for _, l := range likes {
		resp := LikeResponse{
			UserID:    l.UserID,
			CreatedAt: l.CreateTime.Format("2006-01-02T15:04:05Z"),
		}
		if l.Edges.User != nil {
			resp.Username = l.Edges.User.Username
		}
		likeResponses = append(likeResponses, resp)
	}

	response.JSON(w, http.StatusOK, "Project likes retrieved successfully", likeResponses)
}

func (h *ProjectsHandler) CheckProjectLiked(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectID := chi.URLParam(r, "id")
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get user ID from context")
		response.Error(w, errors.ErrUserNotFound)
		return
	}

	existingLike, err := h.client.Like.Query().
		Where(like.UserID(userID), like.ProjectID(projectID)).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		log.Error(ctx).Err(err).Msg("Failed to check like status")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	type LikedResponse struct {
		Liked bool `json:"liked"`
	}

	response.JSON(w, http.StatusOK, "Like status retrieved successfully", LikedResponse{
		Liked: existingLike != nil,
	})
}