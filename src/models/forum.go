package models

type Forum struct {
	Id      int    `json:"-"`
	Title   string `json:"title"`
	Slug    string `json:"slug"`
	Author  string `json:"user"`
	Posts   int    `json:"posts"`
	Threads int    `json:"threads"`
}
