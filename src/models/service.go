package models

type ServiceInfo struct {
	User   uint64 `json:"user"`
	Forum  uint64 `json:"forum"`
	Thread uint64 `json:"thread"`
	Post   uint64 `json:"post"`
}
