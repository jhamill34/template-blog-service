package rbac

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/jhamill34/notion-provisioner/internal/database"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/redis/go-redis/v9"
)

const PREFIX = "policy:"

type PolicyProvider interface {
	GetPolicies(ctx context.Context, id string) (models.PolicyResponse, error)
}

type CasbinAccessControl struct {
	model          string
	keyValueStore  database.KeyValueStoreProvider
	publisher      database.PublisherProvider
	policyProvider PolicyProvider
}

func NewCasbinAccessControl(
	modelDef string,
	keyValueStore database.KeyValueStoreProvider,
	publisher database.PublisherProvider,
	poliPolicyProvider PolicyProvider,
) *CasbinAccessControl {
	return &CasbinAccessControl{
		model:          modelDef,
		keyValueStore:  keyValueStore,
		publisher:      publisher,
		policyProvider: poliPolicyProvider,
	}
}

func (self *CasbinAccessControl) getEnforcer(ctx context.Context, id string) *casbin.Enforcer {
	var p models.PolicyResponse

	result, err := self.keyValueStore.Get().Get(ctx, PREFIX+id)
	if err == redis.Nil {
		p, err = self.policyProvider.GetPolicies(ctx, id)
		if err != nil {
			panic(err)
		}

		valBytes, err := json.Marshal(p)
		if err != nil {
			panic(err)
		}

		value := base64.StdEncoding.EncodeToString(valBytes)
		err = self.keyValueStore.Get().Set(ctx, PREFIX+id, value, 5*time.Minute)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	} else {
		valBytes, err := base64.StdEncoding.DecodeString(result)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(valBytes, &p)
		if err != nil {
			panic(err)
		}
	}

	return self.makeEnforcer(id, p)
}

func (self *CasbinAccessControl) makeEnforcer(
	id string,
	policy models.PolicyResponse,
) *casbin.Enforcer {
	m, err := model.NewModelFromString(self.model)
	if err != nil {
		panic(err)
	}

	e, err := casbin.NewEnforcer(m, false)
	if err != nil {
		panic(err)
	}
	e.EnableLog(true)

	userPrinciple := fmt.Sprintf("u_%s", id)
	for _, permission := range policy.User {
		e.AddPolicy(userPrinciple, permission.Resource, permission.Action, permission.Effect)
	}

	return e
}

// Enforce implements services.AccessControlService.
func (self *CasbinAccessControl) Enforce(
	ctx context.Context,
	resource string,
	action string,
) models.Notifier {
	userId, ok := ctx.Value("user_id").(string)

	if !ok || userId == "" {
		return services.AccessDenied
	}

	principle := fmt.Sprintf("u_%s", userId)
	enforcer := self.getEnforcer(ctx, userId)

	ok, err := enforcer.Enforce(principle, resource, action)
	if err != nil {
		panic(err)
	}

	if !ok {
		return services.AccessDenied
	}

	return nil
}

func (self *CasbinAccessControl) Invalidate(ctx context.Context, id string) {
	self.keyValueStore.Get().Del(ctx, PREFIX+id)

	if self.publisher != nil {
		self.publisher.Get().Publish(ctx, "policy_invalidate", PREFIX+id)
	}
}

// var _ services.AccessControlService = (*CasbinAccessControl)(nil)
