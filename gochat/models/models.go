package models

type Message struct {
	Username  string `json:"username"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

type User struct {
	Username string
}
