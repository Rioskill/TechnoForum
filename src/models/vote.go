package models

type Vote struct {
	UserId   int
	ThreadId int
	Value    int
}

type VoteRequest struct {
	Nickname string
	Voice    int
}
