package users

import (
	"github.com/jorge-j1m/hackspark_server/ent"
)

type UsersHandler struct {
	client *ent.Client
}

func NewUsersHandler(client *ent.Client) *UsersHandler {
	return &UsersHandler{
		client: client,
	}
}
