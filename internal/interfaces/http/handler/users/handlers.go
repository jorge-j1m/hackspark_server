package users

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jorge-j1m/hackspark_server/ent"
	"github.com/jorge-j1m/hackspark_server/ent/tag"
	"github.com/jorge-j1m/hackspark_server/ent/user"
	"github.com/jorge-j1m/hackspark_server/ent/usertechnology"
	log "github.com/jorge-j1m/hackspark_server/internal/infrastructure/logger"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/middleware"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/response"
	"github.com/jorge-j1m/hackspark_server/internal/pkg/common/errors"
)

// UserResponse represents the filtered user data returned in API responses
type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Status    string `json:"status"`
}

// convertUserToResponse converts ent.User to UserResponse, filtering out sensitive data
func convertUserToResponse(user *ent.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		Status:    string(user.AccountStatus),
	}
}

func (u *UsersHandler) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value(log.UserCtxKey).(*ent.User)
	if !ok || user == nil {
		log.Debug(ctx).Msg("Failed to get user from context")
		response.Error(w, errors.ErrUserNotFound)
		return
	}

	response.JSON(w, http.StatusOK, "User fetched successfully", convertUserToResponse(user))
}

func (u *UsersHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := chi.URLParam(r, "username")

	user, err := u.client.User.Query().
		Where(user.Username(username)).
		WithTechnologies().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Error(ctx).Err(err).Msg("User not found")
			response.Error(w, errors.ErrUserNotFound)
			return
		}
		log.Error(ctx).Err(err).Msg("Failed to get user")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	type ProfileResponse struct {
		ID           string   `json:"id"`
		Username     string   `json:"username"`
		Bio          *string  `json:"bio"`
		AvatarURL    *string  `json:"avatar_url"`
		Technologies []string `json:"technologies"`
		ProjectCount int      `json:"project_count"`
	}

	projectCount, _ := user.QueryOwnedProjects().Count(ctx)

	var technologies []string
	if user.Edges.Technologies != nil {
		for _, tech := range user.Edges.Technologies {
			technologies = append(technologies, tech.Slug)
		}
	}

	resp := ProfileResponse{
		ID:           user.ID,
		Username:     user.Username,
		Bio:          user.Bio,
		AvatarURL:    user.AvatarURL,
		Technologies: technologies,
		ProjectCount: projectCount,
	}

	response.JSON(w, http.StatusOK, "User profile retrieved successfully", resp)
}

type AddTechnologyRequest struct {
	TagSlug         string   `json:"tag_slug"`
	SkillLevel      string   `json:"skill_level"`
	YearsExperience *float64 `json:"years_experience"`
	IsPrimary       *bool    `json:"is_primary"`
}

func (r AddTechnologyRequest) Validate() error {
	if r.TagSlug == "" {
		return errors.ErrInvalidRequest
	}
	if r.SkillLevel != "" && r.SkillLevel != "beginner" && r.SkillLevel != "intermediate" && r.SkillLevel != "expert" {
		return errors.ErrInvalidRequest
	}
	return nil
}

func (u *UsersHandler) AddUserTechnology(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get user ID from context")
		response.Error(w, errors.ErrUserNotFound)
		return
	}

	var req AddTechnologyRequest
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

	normalizedSlug := u.normalizeSlug(req.TagSlug)

	tag, err := u.client.Tag.Query().Where(tag.Slug(normalizedSlug)).First(ctx)
	if ent.IsNotFound(err) {
		tag, err = u.client.Tag.Create().
			SetName(req.TagSlug).
			SetSlug(normalizedSlug).
			Save(ctx)
		if err != nil {
			log.Error(ctx).Err(err).Msg("Failed to create tag")
			response.Error(w, errors.ErrInternalServerError)
			return
		}
	} else if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get tag")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	existing, err := u.client.UserTechnology.Query().
		Where(usertechnology.UserID(userID), usertechnology.TechnologyID(tag.ID)).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		log.Error(ctx).Err(err).Msg("Failed to check existing user technology")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	if existing != nil {
		response.Error(w, errors.ErrConflict)
		return
	}

	skillLevel := "beginner"
	if req.SkillLevel != "" {
		skillLevel = req.SkillLevel
	}

	isPrimary := false
	if req.IsPrimary != nil {
		isPrimary = *req.IsPrimary
	}

	_, err = u.client.UserTechnology.Create().
		SetUserID(userID).
		SetTechnologyID(tag.ID).
		SetSkillLevel(usertechnology.SkillLevel(skillLevel)).
		SetNillableYearsExperience(req.YearsExperience).
		SetIsPrimary(isPrimary).
		Save(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to create user technology")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	log.Info(ctx).Msgf("Technology added to user successfully: %s -> %s", userID, tag.Slug)
	response.JSON(w, http.StatusCreated, "Technology added successfully", nil)
}

func (u *UsersHandler) GetUserTechnologies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := chi.URLParam(r, "username")

	user, err := u.client.User.Query().
		Where(user.Username(username)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Error(ctx).Err(err).Msg("User not found")
			response.Error(w, errors.ErrUserNotFound)
			return
		}
		log.Error(ctx).Err(err).Msg("Failed to get user")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	userTechs, err := u.client.UserTechnology.Query().
		Where(usertechnology.UserID(user.ID)).
		WithTechnology().
		Order(ent.Desc(usertechnology.FieldIsPrimary), ent.Asc(usertechnology.FieldSkillLevel)).
		All(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get user technologies")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	type TechnologyResponse struct {
		ID              string   `json:"id"`
		Name            string   `json:"name"`
		Slug            string   `json:"slug"`
		SkillLevel      string   `json:"skill_level"`
		YearsExperience *float64 `json:"years_experience"`
		IsPrimary       bool     `json:"is_primary"`
		AddedAt         string   `json:"added_at"`
	}

	var techResponses []TechnologyResponse
	for _, ut := range userTechs {
		resp := TechnologyResponse{
			ID:              ut.TechnologyID,
			SkillLevel:      string(ut.SkillLevel),
			YearsExperience: ut.YearsExperience,
			IsPrimary:       ut.IsPrimary,
			AddedAt:         ut.CreateTime.Format("2006-01-02T15:04:05Z"),
		}
		if ut.Edges.Technology != nil {
			resp.Name = ut.Edges.Technology.Name
			resp.Slug = ut.Edges.Technology.Slug
		}
		techResponses = append(techResponses, resp)
	}

	response.JSON(w, http.StatusOK, "User technologies retrieved successfully", techResponses)
}

type UpdateTechnologyRequest struct {
	SkillLevel      *string  `json:"skill_level"`
	YearsExperience *float64 `json:"years_experience"`
	IsPrimary       *bool    `json:"is_primary"`
}

func (r UpdateTechnologyRequest) Validate() error {
	if r.SkillLevel != nil && *r.SkillLevel != "beginner" && *r.SkillLevel != "intermediate" && *r.SkillLevel != "expert" {
		return errors.ErrInvalidRequest
	}
	return nil
}

func (u *UsersHandler) UpdateUserTechnology(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get user ID from context")
		response.Error(w, errors.ErrUserNotFound)
		return
	}
	tagSlug := chi.URLParam(r, "slug")

	var req UpdateTechnologyRequest
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

	tag, err := u.client.Tag.Query().Where(tag.Slug(tagSlug)).First(ctx)
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

	userTech, err := u.client.UserTechnology.Query().
		Where(usertechnology.UserID(userID), usertechnology.TechnologyID(tag.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Error(ctx).Err(err).Msg("User technology not found")
			response.Error(w, errors.ErrNotFound)
			return
		}
		log.Error(ctx).Err(err).Msg("Failed to get user technology")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	updateQuery := u.client.UserTechnology.UpdateOneID(userTech.ID)
	if req.SkillLevel != nil {
		updateQuery = updateQuery.SetSkillLevel(usertechnology.SkillLevel(*req.SkillLevel))
	}
	if req.YearsExperience != nil {
		updateQuery = updateQuery.SetNillableYearsExperience(req.YearsExperience)
	}
	if req.IsPrimary != nil {
		updateQuery = updateQuery.SetIsPrimary(*req.IsPrimary)
	}

	if _, err := updateQuery.Save(ctx); err != nil {
		log.Error(ctx).Err(err).Msg("Failed to update user technology")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	log.Info(ctx).Msgf("User technology updated successfully: %s -> %s", userID, tagSlug)
	response.JSON(w, http.StatusOK, "Technology updated successfully", nil)
}

func (u *UsersHandler) RemoveUserTechnology(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get user ID from context")
		response.Error(w, errors.ErrUserNotFound)
		return
	}
	tagSlug := chi.URLParam(r, "slug")

	tag, err := u.client.Tag.Query().Where(tag.Slug(tagSlug)).First(ctx)
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

	userTech, err := u.client.UserTechnology.Query().
		Where(usertechnology.UserID(userID), usertechnology.TechnologyID(tag.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Error(ctx).Err(err).Msg("User technology not found")
			response.Error(w, errors.ErrNotFound)
			return
		}
		log.Error(ctx).Err(err).Msg("Failed to get user technology")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	if err := u.client.UserTechnology.DeleteOneID(userTech.ID).Exec(ctx); err != nil {
		log.Error(ctx).Err(err).Msg("Failed to delete user technology")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	log.Info(ctx).Msgf("User technology removed successfully: %s -> %s", userID, tagSlug)
	response.JSON(w, http.StatusOK, "Technology removed successfully", nil)
}

func (u *UsersHandler) normalizeSlug(input string) string {
	slug := strings.ToLower(input)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, ".", "")
	slug = strings.ReplaceAll(slug, "/", "")
	return slug
}

func (u *UsersHandler) getPagination(r *http.Request) (limit, offset int) {
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
