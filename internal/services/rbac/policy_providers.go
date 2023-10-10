package rbac

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/models"
)

type DatabasePolicyProvider struct {
	dao *dao.UserDao
}

func NewDatabasePolicyProvider(dao *dao.UserDao) *DatabasePolicyProvider {
	return &DatabasePolicyProvider{dao}
}

func (self *DatabasePolicyProvider) GetPolicies(
	ctx context.Context,
	id string,
) (models.PolicyResponse, error) {
	data, err := self.dao.GetPermissions(ctx, id)
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

	return models.PolicyResponse{
		User: policies,
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
