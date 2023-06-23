package models

import "github.com/tee8z/nullable"

type Thread struct {
	Id      int             `json:"id"`
	Title   string          `json:"title"`
	Author  string          `json:"author"`
	Forum   string          `json:"forum"`
	ForumId int             `json:"-"`
	Message string          `json:"message"`
	Votes   int             `json:"votes"`
	Slug    nullable.String `json:"slug, omitempty"`
	Created string          `json:"created"`
}
