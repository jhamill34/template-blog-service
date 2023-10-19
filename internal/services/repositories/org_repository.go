package repositories

import (
	"bytes"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jhamill34/notion-provisioner/internal/database"
	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
)

type OrganizationRepository struct {
	baseUrl              string
	organizationDao      *dao.OrganizationDao
	userDao              *dao.UserDao
	accessControlService services.AccessControlService
	tokenService         services.TokenClaimsService
	emailService         services.EmailSender
	templateService      services.TemplateService
}

func NewOrganizationRepository(
	baseUrl string,
	organizationDao *dao.OrganizationDao,
	userDao *dao.UserDao,
	accessControlService services.AccessControlService,
	tokenService services.TokenClaimsService,
	emailService services.EmailSender,
	templateservice services.TemplateService,
) *OrganizationRepository {
	return &OrganizationRepository{
		baseUrl:              baseUrl,
		organizationDao:      organizationDao,
		userDao:              userDao,
		accessControlService: accessControlService,
		tokenService:         tokenService,
		emailService:         emailService,
		templateService:      templateservice,
	}
}

func (self *OrganizationRepository) ListUsersOrgs(
	ctx context.Context,
	userId string,
) ([]models.Organization, models.Notifier) {
	if err := self.accessControlService.Enforce(ctx, "/org", "list"); err != nil {
		return nil, err
	}

	data, err := self.organizationDao.ListUsersOrgs(ctx, userId)
	if err != nil {
		panic(err)
	}

	orgs := make([]models.Organization, len(data))

	i := 0
	for _, org := range data {
		if err := self.accessControlService.Enforce(ctx, "/org/"+org.Id, "read"); err == nil {
			orgs[i] = models.Organization{
				OrgId:       org.Id,
				Name:        org.Name,
				Description: org.Description,
			}
			i++
		}
	}

	return orgs[:i], nil
}

// CreateOrganization implements services.OrganizationService.
func (self *OrganizationRepository) CreateOrganization(
	ctx context.Context,
	userId, name, description string,
) models.Notifier {
	if err := self.accessControlService.Enforce(ctx, "/org", "create"); err != nil {
		return err
	}

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
	if err := self.accessControlService.Enforce(ctx, "/org/"+id, "read"); err != nil {
		return nil, err
	}

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
func (self *OrganizationRepository) DeleteOrganization(
	ctx context.Context,
	id string,
) models.Notifier {
	if err := self.accessControlService.Enforce(ctx, "/org/"+id, "delete"); err != nil {
		return err
	}

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
	if err := self.accessControlService.Enforce(ctx, "/org/"+id+"/policy", "list"); err != nil {
		return nil, err
	}

	permissions, err := self.organizationDao.GetPermissions(ctx, id)
	if err != nil {
		panic(err)
	}

	policies := make([]models.Policy, len(permissions))
	for i := 0; i < len(policies); i++ {
		policies[i] = models.Policy{
			PolicyId: permissions[i].Id,
			Resource: permissions[i].Resource,
			Action:   permissions[i].Action,
			Effect:   permissions[i].Effect,
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
	if err := self.accessControlService.Enforce(ctx, "/org/"+orgId+"/policy", "create"); err != nil {
		return err
	}

	err := self.organizationDao.CreatePermission(ctx, orgId, resource, action, effect)
	if err != nil {
		panic(err)
	}

	users, err := self.organizationDao.GetUsers(ctx, orgId)
	if err != nil {
		panic(err)
	}
	for _, user := range users {
		self.accessControlService.Invalidate(ctx, user.Id)
	}

	return nil
}

// DeletePolicy implements services.OrganizationService.
func (self *OrganizationRepository) DeletePolicy(
	ctx context.Context,
	orgId string,
	policyId int,
) models.Notifier {
	if err := self.accessControlService.Enforce(ctx, fmt.Sprintf("/org/%s/policy/%d", orgId, policyId), "delete"); err != nil {
		return err
	}

	err := self.organizationDao.DeletePermission(ctx, orgId, policyId)
	if err != nil {
		panic(err)
	}

	users, err := self.organizationDao.GetUsers(ctx, orgId)
	if err != nil {
		panic(err)
	}
	for _, user := range users {
		self.accessControlService.Invalidate(ctx, user.Id)
	}

	return nil
}

//==============================================================================
// User Management
//==============================================================================

// ListUsers implements services.OrganizationService.
func (self *OrganizationRepository) ListUsers(
	ctx context.Context,
	orgId string,
) ([]models.User, models.Notifier) {
	if err := self.accessControlService.Enforce(ctx, "/org/"+orgId+"/user", "list"); err != nil {
		return nil, err
	}

	data, err := self.organizationDao.GetUsers(ctx, orgId)
	if err != nil {
		panic(err)
	}

	users := make([]models.User, len(data))
	i := 0
	for _, user := range data {
		if err := self.accessControlService.Enforce(ctx, "/user/"+user.Id, "read"); err == nil &&
			user.Name != "ROOT" {
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

// AddUser implements services.OrganizationService.
func (self *OrganizationRepository) InviteUser(
	ctx context.Context,
	orgId string,
	email string,
) models.Notifier {
	if err := self.accessControlService.Enforce(ctx, "/org/"+orgId+"/user", "create"); err != nil {
		return err
	}

	org, err := self.GetOrganizationBydId(ctx, orgId)
	if err != nil {
		return err
	}

	user, dberr := self.userDao.FindByEmail(ctx, email)
	if dberr != nil {
		if dberr == database.NotFound {
			return services.UserNotFound
		}

		panic(dberr)
	}

	newId := uuid.New().String()
	token := self.tokenService.CreateWithClaims(
		ctx,
		newId,
		models.InviteData{
			InvitedBy: orgId,
			Email:     user.Email,
		},
	)

	buffer := bytes.Buffer{}
	data := models.EmailWithTokenData{
		BaseUrl: self.baseUrl,
		Token:   token,
		Id:      newId,
	}
	self.templateService.Render(
		&buffer,
		"org_invite_email.html",
		"layout",
		models.NewTemplateData(map[string]interface{}{
			"Name":      org.Name,
			"TokenData": data,
		}),
	)

	self.emailService.SendEmail(ctx, user.Email, "You have been invited", buffer.String())

	return nil
}

// AddUser implements services.OrganizationService.
func (self *OrganizationRepository) Join(
	ctx context.Context,
	tokenId, token, userId string,
) models.Notifier {
	var inviteData models.InviteData
	claimErr := self.tokenService.VerifyWithClaims(ctx, tokenId, token, &inviteData)
	if claimErr != nil {
		return claimErr
	}

	user, err := self.userDao.FindById(ctx, userId)
	if err != nil {
		panic(err)
	}

	if user.Email != inviteData.Email {
		panic(err)
	}

	err = self.organizationDao.AddUser(ctx, inviteData.InvitedBy, userId)
	if err != nil {
		panic(err)
	}

	self.tokenService.Destroy(ctx, tokenId)

	return nil
}

// RemoveUser implements services.OrganizationService.
func (self *OrganizationRepository) RemoveUser(
	ctx context.Context,
	orgId string,
	userId string,
) models.Notifier {
	if err := self.accessControlService.Enforce(ctx, "/org/"+orgId+"/user/"+userId, "delete"); err != nil {
		return err
	}

	err := self.organizationDao.RemoveUser(ctx, orgId, userId)
	if err != nil {
		panic(err)
	}

	return nil
}
