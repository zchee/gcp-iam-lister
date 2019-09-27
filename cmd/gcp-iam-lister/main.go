// Command gcp-iam-lister for detecting IAM roles of the members for the specified Google Cloud Platform project.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/zchee/gcp-iam-lister/pkg/resourcemanager"
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

	rmgr, err := resourcemanager.NewResourceManager(ctx, projectID, credFile)
	if err != nil {
		log.Fatal(err)
	}

	policy, err := rmgr.GetIamPolicy(ctx)
	if err != nil {
		log.Fatal(err)
	}

	members, err := rmgr.GetMembers(ctx, policy)
	if err != nil {
		log.Fatal(err)
	}

	for _, member := range members {
		fmt.Printf("%s: %v\n", member.Name, member.Roles)
	}
}
