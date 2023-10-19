package models

type User struct {
	UserId string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

type Organization struct {
	OrgId       string `json:"org_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type App struct {
	AppId       string `json:"app_id"`
	ClientId    string `json:"client_id"`
	RedirectUri string `json:"redirect_uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type InviteData struct {
	InvitedBy string `json:"invited_by"`
	Email     string `json:"email"`
}

type EmailWithTokenData struct {
	BaseUrl string `json:"base_url"`
	Token   string `json:"token"`
	Id      string `json:"id"`
}
