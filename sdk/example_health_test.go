package dbaas_test

import (
	"context"
	"fmt"

	dbaas "github.com/foyez/dbaas-platform/sdk"
)

func Example_health() {
	client := dbaas.New("http://localhost:8080")

	status, err := client.Health(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println(status.Status)
	// Output: ok
}
