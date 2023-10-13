package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jhamill34/notion-provisioner/internal/database"
	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
)

type OrganizationRepository struct {
	organizationDao      *dao.OrganizationDao
	accessControlService services.AccessControlService
}

func NewOrganizationRepository(
	organizationDao *dao.OrganizationDao,
	accessControlService services.AccessControlService,
) *OrganizationRepository {
	return &OrganizationRepository{
		organizationDao:      organizationDao,
		accessControlService: accessControlService,
	}
}

func (self *OrganizationRepository) ListUsersOrgs(
	ctx context.Context,
	userId string,
) ([]models.Organization, models.Notifier) {
	data, err := self.organizationDao.ListUsersOrgs(ctx, userId)
	if err != nil {
		panic(err)
	}

	orgs := make([]models.Organization, len(data))

	i := 0
	for _, org := range data {
		orgs[i] = models.Organization{
			OrgId:       org.Id,
			Name:        org.Name,
			Description: org.Description,
		}
	}

	return orgs, nil
}

// CreateOrganization implements services.OrganizationService.
func (self *OrganizationRepository) CreateOrganization(
	ctx context.Context,
	userId, name, description string,
) models.Notifier {
	orgId := uuid.New().String()

	_, err := self.organizationDao.CreateOrganization(ctx, orgId, name, description)
	if err != nil {
		panic(err)
	}

	err = self.organizationDao.AddUser(ctx, orgId, userId)
	if err != nil {
		panic(err)
	}

	return nil
}

// GetOrganizationBydId implements services.OrganizationService.
func (self *OrganizationRepository) GetOrganizationBydId(
	ctx context.Context,
	id string,
) (*models.Organization, models.Notifier) {
	org, err := self.organizationDao.FindById(ctx, id)
	if err == database.NotFound {
		return nil, services.OrganizationNotFound
	}

	if err != nil {
		panic(err)
	}

	return &models.Organization{
		OrgId:       org.Id,
		Name:        org.Name,
		Description: org.Description,
	}, nil
}

// DeleteOrganization implements services.OrganizationService.
func (self *OrganizationRepository) DeleteOrganization(ctx context.Context, id string) models.Notifier {
	err := self.organizationDao.RemoveAllUsers(ctx, id)
	if err != nil {
		panic(err)
	}
	
	err = self.organizationDao.DeleteOrganization(ctx, id)
	if err != nil {
		panic(err)
	}

	return nil
}

//==============================================================================
// Policy Management
//==============================================================================

// ListPolicies implements services.OrganizationService.
func (self *OrganizationRepository) ListPolicies(
	ctx context.Context,
	id string,
) ([]models.Policy, models.Notifier) {
	permissions, err := self.organizationDao.GetPermissions(ctx, id)
	if err != nil {
		panic(err)
	}

	policies := make([]models.Policy, len(permissions))
	for i := 0; i < len(policies); i++ {
		policies[i] = models.Policy{
			PolicyId:  permissions[i].Id,
			Resource:  permissions[i].Resource,
			Action:    permissions[i].Action,
			Effect:    permissions[i].Effect,
		}
	}

	return policies, nil
}

// CreatePolicy implements services.OrganizationService.
func (self *OrganizationRepository) CreatePolicy(
	ctx context.Context,
	orgId string,
	resource string,
	action string,
	effect string,
) models.Notifier {
	err := self.organizationDao.CreatePermission(ctx, orgId, resource, action, effect)
	if err != nil {
		panic(err)
	}

	return nil
}

// DeletePolicy implements services.OrganizationService.
func (self *OrganizationRepository) DeletePolicy(
	ctx context.Context,
	orgId string,
	policyId int,
) models.Notifier {
	err := self.organizationDao.DeletePermission(ctx, orgId, policyId)
	if err != nil {
		panic(err)
	}

	return nil
}

//==============================================================================
// User Management
//==============================================================================

// ListUsers implements services.OrganizationService.
func (*OrganizationRepository) ListUsers(
	ctx context.Context,
	orgId string,
) ([]models.User, models.Notifier) {
	panic("unimplemented")
}

// AddUser implements services.OrganizationService.
func (*OrganizationRepository) AddUser(
	ctx context.Context,
	orgId string,
	email string,
) models.Notifier {
	panic("unimplemented")
}

// RemoveUser implements services.OrganizationService.
func (*OrganizationRepository) RemoveUser(
	ctx context.Context,
	orgId string,
	userId string,
) models.Notifier {
	panic("unimplemented")
}
