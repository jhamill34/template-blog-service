package rbac

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
)

// TODO: How should we invalidate caches? Should this be an LRU cache?
type CasbinAccessControl struct {
	model     string
	enforcers map[string]*casbin.Enforcer
	userDao   *dao.UserDao
}

func NewCasbinAccessControl(
	modelDef string,
	userDao *dao.UserDao,
) *CasbinAccessControl {
	enforcers := make(map[string]*casbin.Enforcer)
	return &CasbinAccessControl{
		model:     modelDef,
		enforcers: enforcers,
		userDao:   userDao,
	}
}

func (self *CasbinAccessControl) getEnforcer(ctx context.Context, id string) *casbin.Enforcer {
	e, ok := self.enforcers[id]
	if !ok {
		val := self.makeEnforcer(ctx, id)
		self.enforcers[id] = val
		e = val
	}

	return e
}

func (self *CasbinAccessControl) makeEnforcer(ctx context.Context, id string) *casbin.Enforcer {
	m, err := model.NewModelFromString(self.model)
	if err != nil {
		panic(err)
	}

	e, err := casbin.NewEnforcer(m, false)
	if err != nil {
		panic(err)
	}
	e.EnableLog(true)

	permissions, err := self.userDao.GetPermissions(ctx, id)
	if err != nil {
		panic(err)
	}

	userPrinciple := fmt.Sprintf("u_%s", id)
	for _, permission := range permissions {
		e.AddPolicy(userPrinciple, permission.Resource, permission.Action, permission.Effect)
	}

	return e
}

// Enforce implements services.AccessControlService.
func (self *CasbinAccessControl) Enforce(
	ctx context.Context,
	resource string,
	action string,
) *services.AccessControlError {
	user := ctx.Value("user").(*models.SessionData)

	principle := fmt.Sprintf("u_%s", user.UserId)

	enforcer := self.getEnforcer(ctx, user.UserId)

	if user == nil {
		return services.AccessDenied
	}

	ok, err := enforcer.Enforce(principle, resource, action)
	if err != nil {
		panic(err)
	}

	if !ok {
		return services.AccessDenied
	}

	return nil
}

// var _ services.AccessControlService = (*CasbinAccessControl)(nil)
