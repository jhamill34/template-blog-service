package services

import (
	"io"

	"github.com/jhamill34/notion-provisioner/internal/models"
)

type AuthService interface {
	LoginUser(email string, password string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	CreateUser(username, email, password string) error
	CreateRootUser(email, password string) error
}

type SessionService interface {
	Create(data interface{}) string
	Destroy(id string)
	Find(id string) (interface{}, error)
}

type TemplateService interface {
	Render(w io.Writer, template string, layout string, data interface{}) error
}
