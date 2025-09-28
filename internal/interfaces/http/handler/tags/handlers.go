package tags

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jorge-j1m/hackspark_server/ent"
	"github.com/jorge-j1m/hackspark_server/ent/project"
	"github.com/jorge-j1m/hackspark_server/ent/tag"
	log "github.com/jorge-j1m/hackspark_server/internal/infrastructure/logger"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/response"
	"github.com/jorge-j1m/hackspark_server/internal/pkg/common/errors"
)

type TagResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Icon        *string `json:"icon"`
	Description *string `json:"description"`
	Category    string  `json:"category"`
	UsageCount  int     `json:"usage_count"`
	CreatedAt   string  `json:"created_at"`
}

func (h *TagsHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit, offset := h.getPagination(r)

	query := h.client.Tag.Query().
		Limit(limit).
		Offset(offset)

	if search := r.URL.Query().Get("search"); search != "" {
		query = query.Where(tag.NameContains(search))
	}

	if category := r.URL.Query().Get("category"); category != "" {
		query = query.Where(tag.CategoryEQ(tag.Category(category)))
	}

	sortBy := r.URL.Query().Get("sort")
	switch sortBy {
	case "popular":
		query = query.Order(ent.Desc(tag.FieldUsageCount))
	case "alphabetical":
		query = query.Order(ent.Asc(tag.FieldName))
	case "recent":
		query = query.Order(ent.Desc(tag.FieldCreateTime))
	default:
		query = query.Order(ent.Desc(tag.FieldUsageCount))
	}

	tags, err := query.All(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to list tags")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	var tagResponses []TagResponse
	for _, t := range tags {
		resp := h.buildTagResponse(t)
		tagResponses = append(tagResponses, resp)
	}

	response.JSON(w, http.StatusOK, "Tags retrieved successfully", tagResponses)
}

func (h *TagsHandler) GetTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	tag, err := h.client.Tag.Query().
		Where(tag.Slug(slug)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Error(ctx).Err(err).Msg("Tag not found")
			response.Error(w, errors.ErrNotFound)
			return
		}
		log.Error(ctx).Err(err).Msg("Failed to get tag")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	resp := h.buildTagResponse(tag)
	response.JSON(w, http.StatusOK, "Tag retrieved successfully", resp)
}

func (h *TagsHandler) GetTagProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	limit, offset := h.getPagination(r)

	tag, err := h.client.Tag.Query().
		Where(tag.Slug(slug)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Error(ctx).Err(err).Msg("Tag not found")
			response.Error(w, errors.ErrNotFound)
			return
		}
		log.Error(ctx).Err(err).Msg("Failed to get tag")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	projects, err := tag.QueryProjects().
		WithOwner().
		WithTags().
		Limit(limit).
		Offset(offset).
		Order(ent.Desc(project.FieldCreateTime)).
		All(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get tag projects")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	type ProjectResponse struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Owner       struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"owner"`
		LikeCount int    `json:"like_count"`
		CreatedAt string `json:"created_at"`
	}

	var projectResponses []ProjectResponse
	for _, p := range projects {
		resp := ProjectResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			LikeCount:   p.LikeCount,
			CreatedAt:   p.CreateTime.Format("2006-01-02T15:04:05Z"),
		}
		if p.Edges.Owner != nil {
			resp.Owner = struct {
				ID       string `json:"id"`
				Username string `json:"username"`
			}{
				ID:       p.Edges.Owner.ID,
				Username: p.Edges.Owner.Username,
			}
		}
		projectResponses = append(projectResponses, resp)
	}

	response.JSON(w, http.StatusOK, "Tag projects retrieved successfully", projectResponses)
}

func (h *TagsHandler) GetTagUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	limit, offset := h.getPagination(r)

	tag, err := h.client.Tag.Query().
		Where(tag.Slug(slug)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Error(ctx).Err(err).Msg("Tag not found")
			response.Error(w, errors.ErrNotFound)
			return
		}
		log.Error(ctx).Err(err).Msg("Failed to get tag")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	users, err := tag.QueryUsers().
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get tag users")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	type UserResponse struct {
		ID       string  `json:"id"`
		Username string  `json:"username"`
		Bio      *string `json:"bio"`
	}

	var userResponses []UserResponse
	for _, u := range users {
		resp := UserResponse{
			ID:       u.ID,
			Username: u.Username,
			Bio:      u.Bio,
		}
		userResponses = append(userResponses, resp)
	}

	response.JSON(w, http.StatusOK, "Tag users retrieved successfully", userResponses)
}

func (h *TagsHandler) GetTrendingTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit, _ := h.getPagination(r)
	if limit > 50 {
		limit = 50
	}

	tags, err := h.client.Tag.Query().
		Order(ent.Desc(tag.FieldUsageCount)).
		Limit(limit).
		All(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get trending tags")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	var tagResponses []TagResponse
	for _, t := range tags {
		resp := h.buildTagResponse(t)
		tagResponses = append(tagResponses, resp)
	}

	response.JSON(w, http.StatusOK, "Trending tags retrieved successfully", tagResponses)
}

func (h *TagsHandler) buildTagResponse(t *ent.Tag) TagResponse {
	return TagResponse{
		ID:          t.ID,
		Name:        t.Name,
		Slug:        t.Slug,
		Icon:        t.Icon,
		Description: t.Description,
		Category:    string(t.Category),
		UsageCount:  t.UsageCount,
		CreatedAt:   t.CreateTime.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *TagsHandler) getPagination(r *http.Request) (limit, offset int) {
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

func (h *TagsHandler) normalizeSlug(input string) string {
	slug := strings.ToLower(input)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, ".", "")
	slug = strings.ReplaceAll(slug, "/", "")
	return slug
}
