package dbaas_test

import (
	"context"
	"fmt"
	"os"

	dbaas "github.com/foyez/dbaas-platform/sdk"
)

func newTestClient() *dbaas.Client {
	return dbaas.New(
		os.Getenv("DBAAS_API_URL"),
		dbaas.WithClientCredentials(
			os.Getenv("DBAAS_TOKEN_URL"),
			os.Getenv("DBAAS_CLIENT_ID"),
			os.Getenv("DBAAS_CLIENT_SECRET"),
			[]string{
				"openid",
				"urn:zitadel:iam:org:project:id:" + os.Getenv("DBAAS_PROJECT_ID") + ":aud",
				os.Getenv("DBAAS_ROLE_SCOPE"),
			},
		),
	)
}

func Example_createAndGetInstance() {
	client := newTestClient()
	ctx := context.Background()

	created, err := client.Instances.Create(ctx, dbaas.CreateInstanceInput{
		Name:           "sdk-test-db",
		Version:        16,
		Storage:        "1Gi",
		Username:       "app",
		IdempotencyKey: "sdk-test-key-1",
	})
	if err != nil {
		panic(err)
	}

	fetched, err := client.Instances.Get(ctx, created.ID)
	if err != nil {
		panic(err)
	}

	fmt.Println(fetched.Name)
	// Output: sdk-test-db
}
