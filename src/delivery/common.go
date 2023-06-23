package delivery

import "fmt"

type ErrorMsg struct {
	Message string `json:"message"`
}

func MakeErrorMsg(msg string) string {
	return fmt.Sprintf("{\"message\": \"%v\"}", msg)
}
