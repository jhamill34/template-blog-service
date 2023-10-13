package dao

import (
	"context"
	"database/sql"

	"github.com/jhamill34/notion-provisioner/internal/database"
)

type OrganizationDao struct {
	databaseProvider database.DatabaseProvider
}

func NewOrganizationDao(databaseProvider database.DatabaseProvider) *OrganizationDao {
	return &OrganizationDao{
		databaseProvider: databaseProvider,
	}
}

func (dao *OrganizationDao) ListUsersOrgs(
	ctx context.Context,
	userId string,
) ([]database.OrganizationEntity, error) {
	db := dao.databaseProvider.Get()

	var orgs []database.OrganizationEntity
	err := db.SelectContext(ctx, &orgs, `
		SELECT 
			organization.id as id,
			organization.name as name,
			organization.description as description
		FROM organization
		INNER JOIN organization_user ON organization.id = organization_user.org_id
		WHERE organization_user.user_id = ?
	`, userId)

	if err != nil {
		return nil, err
	}

	return orgs, nil
}

func (dao *OrganizationDao) CreateOrganization(
	ctx context.Context,
	id, name, description string,
) (*database.OrganizationEntity, error) {
	db := dao.databaseProvider.Get()

	if _, err := dao.FindByName(ctx, name); err == database.NotFound {
		_, err := db.ExecContext(ctx, `
		INSERT INTO organization (id, name, description)
		VALUES (?, ?, ?)
	`, id, name, description)

		if err != nil {
			return nil, err
		}

		return &database.OrganizationEntity{
			Id:   id,
			Name: name,
			Description: description,
		}, nil
	} else {
		return nil, database.Duplicate
	}
}

func (dao *OrganizationDao) FindById(
	ctx context.Context,
	id string,
) (*database.OrganizationEntity, error) {
	db := dao.databaseProvider.Get()

	var org database.OrganizationEntity
	err := db.GetContext(ctx, &org, `
		SELECT id, name, description
		FROM organization 
		WHERE id = ?
	`, id)

	if err == sql.ErrNoRows {
		return nil, database.NotFound
	}

	if err != nil {
		return nil, err
	}

	return &org, nil
}

func (dao *OrganizationDao) FindByName(
	ctx context.Context,
	name string,
) (*database.OrganizationEntity, error) {
	db := dao.databaseProvider.Get()

	var org database.OrganizationEntity
	err := db.GetContext(ctx, &org, `
		SELECT id, name, description
		FROM organization 
		WHERE name = ?
	`, name)

	if err == sql.ErrNoRows {
		return nil, database.NotFound
	}

	if err != nil {
		return nil, err
	}

	return &org, nil
}

func (dao *OrganizationDao) DeleteOrganization(
	ctx context.Context,
	id string,
) error {
	db := dao.databaseProvider.Get()

	_, err := db.ExecContext(ctx, `
		DELETE FROM organization
		WHERE id = ?
	`, id)

	if err != nil {
		return err
	}

	return nil
}

func (dao *OrganizationDao) GetPermissions(
	ctx context.Context,
	id string,
) ([]database.OrganizationPermissionEntity, error) {
	db := dao.databaseProvider.Get()

	var permissions []database.OrganizationPermissionEntity
	err := db.SelectContext(ctx, &permissions, `
		SELECT 
			id, org_id, resource, action, effect
		FROM 
			organization_permission
		WHERE 
			org_id = ?
	`, id)

	if err != nil {
		return nil, err
	}

	return permissions, nil
}

func (dao *OrganizationDao) CreatePermission(
	ctx context.Context,
	orgId, resource, action, effect string,
) error {
	db := dao.databaseProvider.Get()

	_, err := db.ExecContext(ctx, `
		INSERT INTO organization_permission
			(org_id, resource, action, effect)
		VALUES (?, ?, ?, ?)
	`, orgId, resource, action, effect)

	if err != nil {
		return err
	}

	return nil
}

func (dao *OrganizationDao) DeletePermission(
	ctx context.Context,
	orgId string,
	permissionId int,
) error {
	db := dao.databaseProvider.Get()

	_, err := db.ExecContext(ctx, `
		DELETE FROM organization_permission
		WHERE org_id = ? AND id = ?
	`, orgId, permissionId)

	if err != nil {
		return err
	}

	return nil
}

func (dao *OrganizationDao) GetUsers(
	ctx context.Context,
	orgId string,
) ([]database.UserEntity, error) {
	db := dao.databaseProvider.Get()

	var users []database.UserEntity
	err := db.SelectContext(ctx, &users, `
		SELECT 
			user.id as id, 
			user.name as name, 
			user.email as email, 
			user.hashashed_password as hashed_password, 
			user.verified as verified, 
			user.created_at as created_at, 
			user.updated_at as updated_at 
		FROM user 
		INNER JOIN organization_user ON user.id = organization_user.user_id
		WHERE organization_user.org_id = ?
	`, orgId)

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (dao *OrganizationDao) CheckIsMember(ctx context.Context, orgId, userId string) error {
	db := dao.databaseProvider.Get()

	var orgUser database.OrganizationUserEntity
	err := db.GetContext(ctx, &orgUser, `
		SELECT 
			id, org_id, user_id
		FROM 
			organization_user
		WHERE 
			org_id = ? AND user_id = ?
	`, orgId, userId)

	if err == sql.ErrNoRows {
		return database.NotFound
	}

	if err != nil {
		return err
	}

	return nil
}

func (dao *OrganizationDao) AddUser(
	ctx context.Context,
	orgId, userId string,
) error {
	db := dao.databaseProvider.Get()

	if err := dao.CheckIsMember(ctx, orgId, userId); err == database.NotFound {
		_, err := db.ExecContext(ctx, `
			INSERT INTO organization_user
				(org_id, user_id)
			VALUES (?, ?)
		`, orgId, userId)

		if err != nil {
			return err
		}
	} else {
		return database.Duplicate
	}

	return nil
}

func (dao *OrganizationDao) RemoveUser(
	ctx context.Context,
	orgId, userId string,
) error {
	db := dao.databaseProvider.Get()

	_, err := db.ExecContext(ctx, `
		DELETE FROM organization_user
		WHERE org_id = ? AND user_id = ?
	`, orgId, userId)

	if err != nil {
		return err
	}

	return nil
}

func (dao *OrganizationDao) RemoveAllUsers(
	ctx context.Context,
	orgId string,
) error {
	db := dao.databaseProvider.Get()

	_, err := db.ExecContext(ctx, `
		DELETE FROM organization_user
		WHERE org_id = ? 
	`, orgId)

	if err != nil {
		return err
	}

	return nil
}
