package structs

type Message struct {
	MessageType string      `json:"type"`
	MessageBody interface{} `json:"body"`
}
