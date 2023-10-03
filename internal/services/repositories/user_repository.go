package repositories

import (
	"context"

	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
)

type UserRepository struct {
	userDao              *dao.UserDao
	accessControlService services.AccessControlService
}

func NewUserRepository(
	userDao *dao.UserDao,
	accessControlService services.AccessControlService,
) *UserRepository {
	return &UserRepository{
		userDao:              userDao,
		accessControlService: accessControlService,
	}
}

// ListUsers implements services.UserService.
func (self *UserRepository) ListUsers(ctx context.Context) []models.User {
	data, err := self.userDao.ListUsers(ctx)
	if err != nil {
		panic(err)
	}

	users := make([]models.User, len(data))

	i := 0
	for _, user := range data {
		if acErr := self.accessControlService.Enforce(ctx, "user:"+user.Id, "read"); acErr == nil {
			users[i] = models.User{
				UserId: user.Id,
				Name:   user.Name,
				Email:  user.Email,
			}
			
			i++
		}
	}

	return users[:i]
}

// var _ services.UserService = (*UserRepository)(nil)
