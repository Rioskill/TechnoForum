package models

import "github.com/tee8z/nullable"

type Post struct {
	Id       int64          `json:"id"`
	Parent   nullable.Int64 `json:"parent,omitempty"`
	Author   string         `json:"author"`
	Message  string         `json:"message"`
	IsEdited bool           `json:"isEdited,omitempty"`
	Forum    string         `json:"forum"`
	Thread   int            `json:"thread"`
	Created  string         `json:"created"`
}

const (
	SortFlat = iota
	SortTree
	SortParent
)

type PostListParams struct {
	ThreadId int
	Limit    int
	Since    int
	Sort     int
	Desc     bool
}

type PostFull struct {
	Post   *Post   `json:"post"`
	Author *User   `json:"author,omitempty"`
	Thread *Thread `json:"thread,omitempty"`
	Forum  *Forum  `json:"forum,omitempty"`
}
