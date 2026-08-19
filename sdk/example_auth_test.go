package dbaas_test

import (
	"context"
	"fmt"
	"os"

	dbaas "github.com/foyez/dbaas-platform/sdk"
)

func Example_auth() {
	client := dbaas.New(
		os.Getenv("DBAAS_API_URL"),
		dbaas.WithClientCredentials(
			os.Getenv("DBAAS_TOKEN_URL"),
			os.Getenv("DBAAS_CLIENT_ID"),
			os.Getenv("DBAAS_CLIENT_SECRET"),
			[]string{"openid", "urn:zitadel:iam:org:project:id:" + os.Getenv("DBAAS_PROJECT_ID") + ":aud"},
		),
	)

	status, err := client.Health(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println(status.Status)
	// Output: ok
}
