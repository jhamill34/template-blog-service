package rbac

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/models"
)

type DatabasePolicyProvider struct {
	userDao *dao.UserDao
	orgDao  *dao.OrganizationDao
}

func NewDatabasePolicyProvider(
	userDao *dao.UserDao,
	orgDao *dao.OrganizationDao,
) *DatabasePolicyProvider {
	return &DatabasePolicyProvider{userDao, orgDao}
}

func (self *DatabasePolicyProvider) GetPolicies(
	ctx context.Context,
	id string,
) (models.PolicyResponse, error) {
	data, err := self.userDao.GetPermissions(ctx, id)
	if err != nil {
		return models.PolicyResponse{}, err
	}

	policies := make([]models.Policy, len(data))
	for i, policy := range data {
		policies[i] = models.Policy{
			PolicyId: policy.Id,
			Resource: policy.Resource,
			Action:   policy.Action,
			Effect:   policy.Effect,
		}
	}

	orgData, err := self.orgDao.ListUsersOrgs(ctx, id)
	if err != nil {
		return models.PolicyResponse{}, err
	}

	orgs := make([]models.OrgPolicyResponse, len(orgData))
	for i, org := range orgData {
		orgPolicyData, err := self.orgDao.GetPermissions(ctx, org.Id)
		if err != nil {
			return models.PolicyResponse{}, err
		}

		orgPolicies := make([]models.Policy, len(orgPolicyData))
		for i, policy := range orgPolicyData {
			orgPolicies[i] = models.Policy{
				PolicyId: policy.Id,
				Resource: policy.Resource,
				Action:   policy.Action,
				Effect:   policy.Effect,
			}
		}

		orgs[i] = models.OrgPolicyResponse{
			OrgId:  org.Id,
			Policy: orgPolicies,
		}
	}

	return models.PolicyResponse{
		User: policies,
		Org: orgs,
	}, nil
}

//==============================================================================

type RemotePolicyProvider struct {
	remotePolicyProviderUrl string
	httpClient              *http.Client
}

func NewRemotePolicyProvider(url string, httpClient *http.Client) *RemotePolicyProvider {
	return &RemotePolicyProvider{url, httpClient}
}

func (self *RemotePolicyProvider) GetPolicies(
	ctx context.Context,
	id string,
) (models.PolicyResponse, error) {
	tokenData := ctx.Value("raw_token").(string)

	req, err := http.NewRequestWithContext(ctx, "GET", self.remotePolicyProviderUrl, nil)
	if err != nil {
		return models.PolicyResponse{}, err
	}

	req.Header.Add("Authorization", "Bearer "+tokenData)

	res, err := self.httpClient.Do(req)
	if err != nil {
		return models.PolicyResponse{}, err
	}

	var policyResponse models.PolicyResponse
	json.NewDecoder(res.Body).Decode(&policyResponse)

	return policyResponse, nil
}
