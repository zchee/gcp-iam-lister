package iam

import (
	"context"
	"fmt"
	"path"

	"cloud.google.com/go/iam"
	iamadminv1 "cloud.google.com/go/iam/admin/apiv1"
	apioption "google.golang.org/api/option"
	adminpb "google.golang.org/genproto/googleapis/iam/admin/v1"
	iampb "google.golang.org/genproto/googleapis/iam/v1"
)

// IAM represents a Iam client.
type IAM struct {
	client    *iamadminv1.IamClient
	projectID string
}

// NewClient returns the new iamadminv1.IamClient.
func NewClient(ctx context.Context, projectID, credfile string) (*IAM, error) {
	opts := []apioption.ClientOption{
		apioption.WithCredentialsFile(credfile),
	}
	c, err := iamadminv1.NewIamClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create IamClient: %w", err)
	}

	return &IAM{
		client:    c,
		projectID: projectID,
	}, nil
}

// Close closes the iam client.
func (i *IAM) Close() error {
	return i.client.Close()
}

func (i *IAM) GetIamPolicy(ctx context.Context, resource string) (*iam.Policy, error) {
	return i.client.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: resource,
	})
}

// GetRequest provides the service account list request.
type GetRequest struct {
	req *adminpb.GetServiceAccountRequest
}

// NewGetRequest returns the new GetRequest.
func (iam *IAM) NewGetRequest(saEmail string) *GetRequest {
	return &GetRequest{
		req: &adminpb.GetServiceAccountRequest{
			Name: path.Join("projects", iam.projectID, "serviceAccounts", saEmail),
		},
	}
}

// GetServiceAccounts lists ServiceAccounts for a project.
func (i *IAM) GetServiceAccounts(ctx context.Context, req *GetRequest) (*adminpb.ServiceAccount, error) {
	return i.client.GetServiceAccount(ctx, req.req)
}

// ListRequest provides the service account list request.
type ListRequest struct {
	req *adminpb.ListServiceAccountsRequest
}

// NewListRequest returns the new ListRequest.
func (i *IAM) NewListRequest(pageSize int32, pageToken string) *ListRequest {
	return &ListRequest{
		req: &adminpb.ListServiceAccountsRequest{
			Name:      path.Join("projects", i.projectID),
			PageSize:  pageSize,
			PageToken: pageToken,
		},
	}
}

// ListServiceAccounts lists ServiceAccounts for a project.
func (i *IAM) ListServiceAccounts(ctx context.Context, req *ListRequest) (results []*adminpb.ServiceAccount, nextPageToken string, err error) {
	resp := i.client.ListServiceAccounts(ctx, req.req)
	return resp.InternalFetch(int(req.req.PageSize), req.req.PageToken)
}
