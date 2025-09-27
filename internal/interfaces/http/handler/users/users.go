package users

import (
	"github.com/jorge-j1m/hackspark_server/ent"
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

func NewUsersHandler(client *ent.Client) *UsersHandler {
	return &UsersHandler{
		client: client,
	}
}
