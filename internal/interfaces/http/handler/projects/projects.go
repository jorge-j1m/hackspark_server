package projects

import (
	"github.com/jorge-j1m/hackspark_server/ent"
)

type ProjectsHandler struct {
	client *ent.Client
}

func NewProjectsHandler(client *ent.Client) *ProjectsHandler {
	return &ProjectsHandler{
		client: client,
	}
}