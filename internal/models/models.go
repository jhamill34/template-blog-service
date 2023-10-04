package models

type User struct {
	UserId string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

type App struct {
	AppId         string `json:"app_id"`
	ClientId      string `json:"client_id"`
	RedirectUri   string `json:"redirect_uri"`
	Name          string `json:"name"`
	Description   string `json:"description"`
}

type Policy struct {
	PolicyId int    `json:"policy_id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Effect   string `json:"effect"`
}

type InviteData struct {
	InvitedBy string `json:"invited_by"`
	Email     string `json:"email"`
}

type SessionData struct {
	SessionId string `json:"session_id"`
	UserId    string `json:"user_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CsrfToken string `json:"csrf_token"`
}

type Notifier interface {
	Notify() *Notification
}

type Notification struct {
	Message string `json:"message"`
}
