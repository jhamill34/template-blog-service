package models

type User struct {
	UserId string `json:"user_id" redis:"user_id"`
	Name   string `json:"name"    redis:"name"`
	Email  string `json:"email"   redis:"email"`
}

type InviteData struct {
	InvitedBy string `json:"invited_by"`
	Email     string `json:"email"`
}
