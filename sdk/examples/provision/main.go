// Command provision creates a PostgreSQL instance, waits until it's running
// and prints its connection string: what a CI job that needs a throwaway
// database would do.
//
// It reads the DBAAS_* variables from .env.example; from sdk/:
//
//	set -a; . ./.env; set +a
//	go run ./examples/provision -name orders -key orders-ci-42
//	go run ./examples/provision -key orders-ci-42 -delete
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/foyez/dbaas-platform/sdk/dbaas"
)

func main() {
	name := flag.String("name", "app", "database name")
	key := flag.String("key", "", "idempotency key: rerunning with the same key reuses the instance (required)")
	del := flag.Bool("delete", false, "delete the instance created with -key instead")
	timeout := flag.Duration("timeout", 10*time.Minute, "how long to wait for the instance")
	flag.Parse()

	if *key == "" {
		flag.Usage()
		os.Exit(2)
	}

	// Ctrl-C cancels the context, which stops the wait cleanly.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()

	c := dbaas.New(mustEnv("DBAAS_API_URL"),
		dbaas.WithClientCredentials(mustEnv("DBAAS_TOKEN_URL"), mustEnv("DBAAS_CLIENT_ID"), mustEnv("DBAAS_CLIENT_SECRET"), []string{
			"openid",
			"urn:zitadel:iam:org:project:id:" + mustEnv("DBAAS_PROJECT_ID") + ":aud",
		}),
	)

	// Create is idempotent, so this both creates the instance on the first
	// run and finds it again (for -delete or a rerun) afterwards.
	inst, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
		Name:           *name,
		Version:        16,
		Storage:        "1Gi",
		IdempotencyKey: *key,
	})
	if err != nil {
		log.Fatal(err)
	}

	if *del {
		if err := c.Instances.Delete(ctx, inst.ID); err != nil {
			log.Fatal(err)
		}
		log.Printf("deleting %s", inst.ID)
		return
	}

	log.Printf("waiting for %s to be running...", inst.ID)
	if _, err := c.Instances.WaitUntilRunning(ctx, inst.ID); err != nil {
		log.Fatal(err)
	}

	creds, err := c.Instances.Credentials(ctx, inst.ID)
	if err != nil {
		log.Fatal(err)
	}

	// Only the URI goes to stdout, so it can be captured:
	// DATABASE_URL=$(go run ./examples/provision -key ...)
	fmt.Println(creds.URI)
}

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("%s is not set (see .env.example)", name)
	}
	return v
}
