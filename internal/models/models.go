package models

type User struct {
	UserId   string `json:"user_id"  redis:"user_id"`
	Name     string `json:"name"     redis:"name"`
	Email    string `json:"email"    redis:"email"`
}
