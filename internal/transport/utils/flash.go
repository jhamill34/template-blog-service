package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jhamill34/notion-provisioner/internal/models"
)

const NOTIFICATION_COOKIE_NAME = "notifications"

type GenericNotifier struct {
	Message string
}

func NewGenericMessage(message string) *GenericNotifier {
	return &GenericNotifier{Message: message}
}

func (self *GenericNotifier) Notify() *models.Notification {
	return &models.Notification{Message: self.Message}
}

var NotificationNotFound = errors.New("Notification cookie not found")

func GetNotifications(r *http.Request) *models.Notification {
	cookie, err := r.Cookie(NOTIFICATION_COOKIE_NAME)
	if err == http.ErrNoCookie {
		return nil
	}

	if err != nil {
		return nil
	}

	raw, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil
	}

	var notification models.Notification
	err = json.Unmarshal(raw, &notification)
	if err != nil {
		return nil
	}

	return &notification
}

func SetNotifications(
	w http.ResponseWriter,
	notifier models.Notifier,
	scope string,
	dur time.Duration,
) error {
	data, err := json.Marshal(notifier.Notify())
	if err != nil {
		return err
	}

	value := base64.URLEncoding.EncodeToString(data)

	cookie := &http.Cookie{
		Name:     NOTIFICATION_COOKIE_NAME,
		Value:    value,
		Path:     scope,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(dur),
	}

	http.SetCookie(w, cookie)
	return nil
}
