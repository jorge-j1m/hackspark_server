package auth

import (
	"github.com/jorge-j1m/hackspark_server/ent"
)

type AuthHandler struct {
	client *ent.Client
}

func NewAuthHandler(client *ent.Client) *AuthHandler {
	return &AuthHandler{
		client: client,
	}
}
