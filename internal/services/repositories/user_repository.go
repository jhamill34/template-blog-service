package repositories

import (
	"context"

	"github.com/jhamill34/notion-provisioner/internal/database"
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
func (self *UserRepository) ListUsers(ctx context.Context) ([]models.User, models.Notifier) {
	if acErr := self.accessControlService.Enforce(ctx, "/user", "list"); acErr != nil {
		return nil, acErr
	}

	data, err := self.userDao.ListUsers(ctx)
	if err != nil {
		panic(err)
	}

	users := make([]models.User, len(data))

	i := 0
	for _, user := range data {
		if acErr := self.accessControlService.Enforce(ctx, "/user/"+user.Id, "read"); acErr == nil &&
			user.Id != "ROOT" {
			users[i] = models.User{
				UserId: user.Id,
				Name:   user.Name,
				Email:  user.Email,
			}

			i++
		}
	}

	return users[:i], nil
}

func (self *UserRepository) GetUser(
	ctx context.Context,
	id string,
) (*models.User, models.Notifier) {
	if acErr := self.accessControlService.Enforce(ctx, "/user/"+id, "read"); acErr != nil {
		return nil, acErr
	}

	user, err := self.userDao.FindById(ctx, id)
	if err == database.NotFound {
		return nil, services.UserNotFound
	}

	if err != nil {
		panic(err)
	}

	return &models.User{
		UserId: user.Id,
		Name:   user.Name,
		Email:  user.Email,
	}, nil
}

// ListPolicies implements services.UserService.
func (self *UserRepository) ListPolicies(ctx context.Context, id string) ([]models.Policy, models.Notifier) {
	if acErr := self.accessControlService.Enforce(ctx, "/user/"+id+"/policy", "list"); acErr != nil {
		return nil, acErr
	}

	data, err := self.userDao.GetPermissions(ctx, id)
	if err != nil {
		panic(err)
	}

	permissions := make([]models.Policy, len(data))
	for i, permission := range data {
		permissions[i] = models.Policy{
			PolicyId: permission.Id,
			Resource: permission.Resource,
			Action:   permission.Action,
			Effect:   permission.Effect,
		}
	}

	return permissions, nil
}

func (self *UserRepository) CreatePolicy(
	ctx context.Context,
	id string,
	resource string,
	action string,
	effect string,
) models.Notifier {
	if acErr := self.accessControlService.Enforce(ctx, "/user/"+id+"/policy", "create"); acErr != nil {
		return acErr
	}

	if err := self.userDao.CreatePermission(ctx, id, resource, action, effect); err != nil {
		panic(err)
	}

	self.accessControlService.Invalidate(ctx, id)

	return nil
}

func (self *UserRepository) DeletePolicy(
	ctx context.Context,
	id, policyId string,
) models.Notifier {
	if acErr := self.accessControlService.Enforce(ctx, "/user/"+id+"/policy/"+policyId, "delete"); acErr != nil {
		return acErr
	}

	if err := self.userDao.DeletePermission(ctx, id, policyId); err != nil {
		panic(err)
	}

	self.accessControlService.Invalidate(ctx, id)

	return nil
}

// var _ services.UserService = (*UserRepository)(nil)
