package models

import "errors"

var (
	ErrAlreadyExists   = errors.New("such object already exists")
	ErrNotFound        = errors.New("no such object")
	ErrNoParent        = errors.New("no parent found")
	ErrInvalidParent   = errors.New("invalid parent")
	ErrInvalidArgument = errors.New("invalid argument")
)
