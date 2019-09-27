package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/zchee/gcp-iam-lister/pkg/iam"
	cloudresourcemanagerv1 "google.golang.org/api/cloudresourcemanager/v1"
)

var (
	projectID string
	credFile  string
)

func init() {
	log.SetFlags(log.Lshortfile)

	flag.StringVar(&projectID, "project", "", "GCP project name")
	flag.StringVar(&credFile, "credential", "", "GCP credentials file")
}

func main() {
	flag.Parse()
	if projectID == "" {
		log.Fatal("missing GCP project name")
	}
	if credFile == "" {
		log.Fatal("missing credential file")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := iam.NewClient(ctx, projectID, credFile)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	listReq := c.NewListRequest(100, "")
	results, _, err := c.ListServiceAccounts(ctx, listReq)
	if err != nil {
		log.Fatal(err)
	}

	sort.Slice(results, func(i, j int) bool { return results[i].GetEmail() < results[j].GetEmail() })
	fmt.Printf("\n\n==================== serviceAccounts ====================\n\n")
	for _, result := range results {
		fmt.Printf("%s\n", result.GetEmail())
	}

	crmgr, err := cloudresourcemanagerv1.NewService(ctx)
	if err != nil {
		log.Fatal(err)
	}

	policy, err := crmgr.Projects.GetIamPolicy(projectID, &cloudresourcemanagerv1.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		log.Fatal(err)
	}

	saRoleMap := make(map[string][]string)
	for _, result := range results {
		saEmail := result.GetEmail()
		roles := make([]string, 0, len(policy.Bindings))

		for _, binding := range policy.Bindings {
			exist, err := existsIamMember(binding.Members, saEmail)
			if err != nil {
				log.Fatal(err)
			}

			if exist {
				roles = append(roles, binding.Role)
			}
		}

		saRoleMap[saEmail] = roles
	}

	fmt.Printf("\n\n==================== serviceAccounts and Roles ====================\n\n")
	for sa, roles := range saRoleMap {
		fmt.Printf("serviceAccount: %s, roles: %v\n", sa, roles)
	}
}

func existsIamMember(members []string, email string) (bool, error) {
	for _, member := range members {
		iamMember, err := ConvertIamMember(member)
		if err != nil {
			return false, err
		}
		if email == iamMember.Email {
			return true, nil
		}
	}

	return false, nil
}

type IamMember struct {
	Type  string
	Email string
}

// ConvertIamMember is IAM RoleのAPIで取得できるMember文字列をIamMember structに変換して返す
func ConvertIamMember(member string) (*IamMember, error) {
	l := strings.Split(member, ":")
	if len(l) != 2 {
		return nil, errors.New("Invalid Iam member text")
	}

	switch l[0] {
	case "user", "serviceAccount", "group", "domain":
		return &IamMember{
			Type:  l[0],
			Email: l[1],
		}, nil
	default:
		return nil, errors.New("Invalid Iam member text")
	}
}
