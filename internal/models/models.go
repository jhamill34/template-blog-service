package models

type User struct {
	UserId string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

type InviteData struct {
	InvitedBy string `json:"invited_by"`
	Email     string `json:"email"`
}

type Notifier interface {
	Notify() *Notification
}

type Notification struct {
	Message string `json:"message"`
}

type SessionData struct {
	SessionId string `json:"session_id"`
	UserId    string `json:"user_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CsrfToken string `json:"csrf_token"`
}
