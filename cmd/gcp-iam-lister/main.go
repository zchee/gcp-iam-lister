package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/zchee/gcp-iam-lister/pkg/iam"
)

var (
	gcpProject string
	credFile   string
)

func init() {
	flag.StringVar(&gcpProject, "project", "", "GCP project name")
	flag.StringVar(&credFile, "credential", "", "GCP credentials file")
}

func main() {
	flag.Parse()
	if gcpProject == "" {
		log.Fatal("missing GCP project name")
	}
	if credFile == "" {
		log.Fatal("missing credential file")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := iam.NewClient(ctx, gcpProject, credFile)
	if err != nil {
		log.Fatal(err)
	}

	listReq := c.NewListRequest(100, "")
	results, _, err := c.ListServiceAccounts(ctx, listReq)
	if err != nil {
		log.Fatal(err)
	}

	for _, resp := range results {
		policy, err := c.GetIamPolicy(ctx, "projects/"+gcpProject+"/serviceAccounts/"+resp.GetEmail())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("policy.Roles(): %s\n", spew.Sdump(policy.Roles()))
	}
}
