package resourcemanager

import (
	"context"
	"strings"

	errors "golang.org/x/xerrors"

	cloudresourcemanagerv1 "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"

	"github.com/zchee/gcp-iam-lister/pkg/errcode"
)

// ResourceManager represents a Google Cloud Resource Manager service client.
type ResourceManager struct {
	svc *cloudresourcemanagerv1.Service

	projectID string
}

// NewResourceManager creates a new ResourceManager.
func NewResourceManager(ctx context.Context, projectID, credfile string) (*ResourceManager, error) {
	hc, _, err := transport.NewHTTPClient(ctx)
	if err != nil {
		return nil, errors.Errorf("unable to create the Google API specific http client: %w", errcode.FromError(err))
	}

	opts := []option.ClientOption{
		option.WithCredentialsFile(credfile),
		option.WithHTTPClient(hc),
	}
	svc, err := cloudresourcemanagerv1.NewService(ctx, opts...)
	if err != nil {
		return nil, errors.Errorf("unable to create the Resource Manager service client: %w", errcode.FromError(err))
	}

	return &ResourceManager{
		svc:       svc,
		projectID: projectID,
	}, nil
}

// IAMMemberType defines the IAM member type.
type IAMMemberType string

const (
	User           = IAMMemberType("user")
	ServiceAccount = IAMMemberType("serviceAccount")
	Group          = IAMMemberType("group")
	Domain         = IAMMemberType("domain")

	// AllUsers a special identifier that represents anyone who is on the internet; with or without a Google account.
	AllUsers = IAMMemberType("allUsers")

	// AllAuthenticatedUsers a special identifier that represents.
	AllAuthenticatedUsers = IAMMemberType("allAuthenticatedUsers")
)

func detectIAMMemberType(member string) (string, IAMMemberType) {
	if idx := strings.Index(member, ":"); idx > -1 {
		switch iamType := IAMMemberType(member[:idx]); iamType {
		case User, ServiceAccount, Group, Domain:
			return member[idx+1:], iamType
		default:
			return "", "" // not reached
		}
	}

	switch iamType := IAMMemberType(member); iamType {
	case AllUsers, AllAuthenticatedUsers:
		return member, iamType
	}

	return "", "" // unknown
}

var emptyGetIamPolicyRequest = &cloudresourcemanagerv1.GetIamPolicyRequest{}

// GetIamPolicy returns the IAM access control policy for the specified Google Cloud Platform project.
func (rm *ResourceManager) GetIamPolicy(ctx context.Context) (*cloudresourcemanagerv1.Policy, error) {
	policy, err := rm.svc.Projects.GetIamPolicy(rm.projectID, emptyGetIamPolicyRequest).Context(ctx).Do()
	if err != nil {
		return nil, errors.Errorf("unable to get IAM policy: %w", errcode.FromError(err))
	}

	return policy, nil
}

// Member represents an IAM identity details accessing for a Google Cloud Platform resource.
type Member struct {
	Type  IAMMemberType
	Name  string
	Roles []string
}

// GetMembers parses IAM access control policy and returns the project Members.
func (rm *ResourceManager) GetMembers(ctx context.Context, policy *cloudresourcemanagerv1.Policy) ([]*Member, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	m := make(map[string]*Member) // policy.Binding.Member: []*Member
	for _, binding := range policy.Bindings {
		for _, member := range binding.Members {
			mb, ok := m[member]
			if ok {
				mb.Roles = append(mb.Roles, binding.Role)
				m[member] = mb
				break // break binding.Members loop
			}

			name, typ := detectIAMMemberType(member)
			if name == "" {
				break // break binding.Members loop
			}

			m[member] = &Member{
				Type:  typ,
				Name:  name,
				Roles: []string{binding.Role},
			}
		}
	}

	ms := make([]*Member, len(m))
	var i int
	for _, member := range m {
		ms[i] = member
		i++
	}

	return ms, nil
}
