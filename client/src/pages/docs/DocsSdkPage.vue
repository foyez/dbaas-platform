<template>
  <article class="space-y-6">
    <h1 class="text-3xl font-bold tracking-tight">Go SDK</h1>

    <p class="leading-7 text-muted-foreground">
      A typed Go client for the REST API, with OAuth2 client-credentials authentication built in.
    </p>

    <div class="space-y-2">
      <h2 class="text-xl font-semibold tracking-tight">Install</h2>
      <pre class="overflow-x-auto rounded-lg bg-muted p-4 font-mono text-sm"><code>go get github.com/foyez/dbaas-platform/sdk</code></pre>
    </div>

    <div class="space-y-2">
      <h2 class="text-xl font-semibold tracking-tight">Usage</h2>
      <pre class="overflow-x-auto rounded-lg bg-muted p-4 font-mono text-sm"><code>c := dbaas.New("https://api.example.com",
    dbaas.WithClientCredentials(tokenURL, clientID, clientSecret, []string{"dbaas"}),
)

inst, _ := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
    Name:           "orders-db",
    Version:        16,
    Storage:        "10Gi",
    IdempotencyKey: "orders-db-create-1",
})

creds, _ := c.Instances.Credentials(ctx, inst.ID)</code></pre>
    </div>

    <p class="text-sm text-muted-foreground">
      Everything the SDK does maps directly to the
      <RouterLink :to="{ name: 'docs-api' }" class="font-medium text-foreground underline underline-offset-4">
        REST API
      </RouterLink>
      - it's a thin, typed wrapper, not a separate code path.
    </p>
  </article>
</template>