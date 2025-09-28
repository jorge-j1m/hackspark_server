package users

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jorge-j1m/hackspark_server/ent"
	"github.com/jorge-j1m/hackspark_server/ent/tag"
	"github.com/jorge-j1m/hackspark_server/ent/usertechnology"
	log "github.com/jorge-j1m/hackspark_server/internal/infrastructure/logger"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/middleware"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/response"
	"github.com/jorge-j1m/hackspark_server/internal/pkg/common/errors"
)

func (u *UsersHandler) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value(log.UserCtxKey).(*ent.User)
	if !ok || user == nil {
		log.Debug(ctx).Msg("Failed to get user from context")
		response.Error(w, errors.ErrUserNotFound)
		return
	}

	userTechs, err := u.getUserTechnologies(ctx, user.ID)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get user technologies")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	userProjects, err := u.getUserProjects(ctx, user.Username)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get user projects")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	type MeResponse struct {
		UserData
		Technologies []TechnologyResponse `json:"technologies"`
		Projects     []ProjectResponse    `json:"projects"`
	}

	response.JSON(w, http.StatusOK, "User fetched successfully", MeResponse{
		UserData:     convertUserToUserData(user),
		Technologies: convertUserTechnologiesToResponse(userTechs),
		Projects:     convertProjectsToResponse(userProjects),
	})
}

// General use handler for getting someone else's profile by username
func (u *UsersHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := chi.URLParam(r, "username")

	user, err := u.getUserByUsername(ctx, username)
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

	userTechs, err := u.getUserTechnologies(ctx, user.ID)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get user technologies")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	type ProfileResponse struct {
		UserData
		Bio          *string  `json:"bio"`
		AvatarURL    *string  `json:"avatar_url"`
		Technologies []string `json:"technologies"`
		ProjectCount int      `json:"project_count"`
		TotalLikes   int      `json:"total_likes"`
	}

	projectCount, _ := user.QueryOwnedProjects().Count(ctx)
	totalLikes, _ := user.QueryOwnedProjects().QueryLikes().Count(ctx)

	var technologies []string
	for _, tech := range userTechs {
		technologies = append(technologies, tech.Edges.Technology.Slug)
	}

	resp := ProfileResponse{
		UserData:     convertUserToUserData(user),
		Bio:          user.Bio,
		AvatarURL:    user.AvatarURL,
		Technologies: technologies,
		ProjectCount: projectCount,
		TotalLikes:   totalLikes,
	}

	response.JSON(w, http.StatusOK, "User profile retrieved successfully", resp)
}

type AddTechnologyRequest struct {
	TagSlug         string   `json:"tag_slug"`
	SkillLevel      string   `json:"skill_level"`
	YearsExperience *float64 `json:"years_experience"`
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

	_, err = u.client.UserTechnology.Create().
		SetUserID(userID).
		SetTechnologyID(tag.ID).
		SetSkillLevel(usertechnology.SkillLevel(skillLevel)).
		SetNillableYearsExperience(req.YearsExperience).
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

	user, err := u.getUserByUsername(ctx, username)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get user")
		response.Error(w, errors.ErrUserNotFound)
		return
	}

	userTechs, err := u.getUserTechnologies(ctx, user.ID)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get user technologies")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	techResponses := convertUserTechnologiesToResponse(userTechs)
	response.JSON(w, http.StatusOK, "User technologies retrieved successfully", techResponses)
}

func (u *UsersHandler) GetUserProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := chi.URLParam(r, "username")

	userProjects, err := u.getUserProjects(ctx, username)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to get user projects")
		response.Error(w, errors.ErrInternalServerError)
		return
	}

	projectResponses := convertProjectsToResponse(userProjects)
	response.JSON(w, http.StatusOK, "User projects retrieved successfully", projectResponses)
}

func (u *UsersHandler) GetUserLikes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := chi.URLParam(r, "username")

	userLikes, err := u.getUserLikedProjects(ctx, username)
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

	projectResponses := convertProjectsToExtendedResponse(userLikes)
	response.JSON(w, http.StatusOK, "User likes retrieved successfully", projectResponses)
}

type UpdateTechnologyRequest struct {
	SkillLevel      *string  `json:"skill_level"`
	YearsExperience *float64 `json:"years_experience"`
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
