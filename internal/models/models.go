package models

type Policy struct {
	PolicyId int    `json:"policy_id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Effect   string `json:"effect"`
}

type AccessTokenClaims struct {
	Sub string `json:"sub"`
	Aud string `json:"aud"`
	Iss string `json:"iss"`
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
}

type PolicyResponse struct {
	User []Policy            `json:"user"`
	Org  []OrgPolicyResponse `json:"org"`
}

type OrgPolicyResponse struct {
	OrgId  string   `json:"org_id"`
	Policy []Policy `json:"policy"`
}

type AccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Expires      int64  `json:"expires"`
}

type PublicKeyResponse struct {
	PublicKey string `json:"public_key"`
	TTL       int64  `json:"ttl"`
}

type Notifier interface {
	Notify() *Notification
}

type Notification struct {
	Message string `json:"message"`
}

func (self *Notification) Notify() *Notification {
	return self
}

type SessionData struct {
	Payload   string `json:"payload"`
	Type      string `json:"type"`
	CsrfToken string `json:"csrf_token"`
}
