package model2

type PanicMessage struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
