package users

import (
	"context"

	"github.com/jorge-j1m/hackspark_server/ent"
	user_ent "github.com/jorge-j1m/hackspark_server/ent/user"
	"github.com/jorge-j1m/hackspark_server/ent/usertechnology"
)

type UsersHandler struct {
	client *ent.Client
}

type UserData struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
	Email     string `json:"email"`
}

type CreatedUser struct {
	UserData
	Id string `json:"id"`
}

type TechnologyResponse struct {
	Name            string   `json:"name"`
	Slug            string   `json:"slug"`
	SkillLevel      string   `json:"skill_level"`
	YearsExperience *float64 `json:"years_experience"`
}

type ProjectResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	LikeCount   int     `json:"like_count"`
	StarCount   int     `json:"star_count"`
	AddedAt     string  `json:"added_at"`
}

func NewUsersHandler(client *ent.Client) *UsersHandler {
	return &UsersHandler{
		client: client,
	}
}

func (u *UsersHandler) getUserByUsername(ctx context.Context, username string) (*ent.User, error) {
	return u.client.User.Query().
		Where(user_ent.Username(username)).
		First(ctx)
}

func (u *UsersHandler) getUserTechnologies(ctx context.Context, userID string) ([]*ent.UserTechnology, error) {
	return u.client.UserTechnology.
		Query().
		Where(usertechnology.UserID(userID)).
		WithTechnology().
		All(ctx)
}

func (u *UsersHandler) getUserProjects(ctx context.Context, username string) ([]*ent.Project, error) {
	return u.client.User.Query().
		Where(user_ent.Username(username)).
		QueryOwnedProjects().
		All(ctx)
}

func (u *UsersHandler) getUserLikedProjects(ctx context.Context, username string) ([]*ent.Project, error) {
	return u.client.User.Query().
		Where(user_ent.Username(username)).
		QueryLikedProjects().
		All(ctx)
}

func convertUserTechnologiesToResponse(userTechs []*ent.UserTechnology) []TechnologyResponse {
	techResponses := make([]TechnologyResponse, len(userTechs))
	for i, tech := range userTechs {
		techResponses[i] = TechnologyResponse{
			Name:            tech.Edges.Technology.Name,
			Slug:            tech.Edges.Technology.Slug,
			SkillLevel:      string(tech.SkillLevel),
			YearsExperience: tech.YearsExperience,
		}
	}
	return techResponses
}

func convertProjectsToResponse(projects []*ent.Project) []ProjectResponse {
	projectResponses := make([]ProjectResponse, len(projects))
	for i, project := range projects {
		projectResponses[i] = ProjectResponse{
			ID:          project.ID,
			Name:        project.Name,
			Description: project.Description,
			LikeCount:   project.LikeCount,
			StarCount:   project.StarCount,
			AddedAt:     project.CreateTime.Format("2006-01-02T15:04:05Z"),
		}
	}
	return projectResponses
}

func convertProjectsToExtendedResponse(projects []*ent.Project) []ProjectResponse {
	projectResponses := make([]ProjectResponse, len(projects))
	for i, project := range projects {
		projectResponses[i] = ProjectResponse{
			ID:          project.ID,
			Name:        project.Name,
			Description: project.Description,
			LikeCount:   project.LikeCount,
			StarCount:   project.StarCount,
			AddedAt:     project.CreateTime.Format("2006-01-02T15:04:05Z"),
		}
	}
	return projectResponses
}

func convertUserToUserData(user *ent.User) UserData {
	return UserData{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
	}
}
