package tags

import (
	"github.com/jorge-j1m/hackspark_server/ent"
)

type TagsHandler struct {
	client *ent.Client
}

func NewTagsHandler(client *ent.Client) *TagsHandler {
	return &TagsHandler{
		client: client,
	}
}