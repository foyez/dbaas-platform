<template>
  <article class="space-y-6">
    <h1 class="text-3xl font-bold tracking-tight">Getting started</h1>

    <div class="space-y-2">
      <h2 class="text-xl font-semibold tracking-tight">1. Sign in</h2>
      <p class="leading-7 text-muted-foreground">
        Click <strong class="font-medium text-foreground">Get started</strong> on the homepage,
        or <strong class="font-medium text-foreground">Log in</strong> in the header. You'll be
        redirected to sign in, then brought back to your dashboard.
      </p>
    </div>

    <div class="space-y-2">
      <h2 class="text-xl font-semibold tracking-tight">2. Create an instance</h2>
      <p class="leading-7 text-muted-foreground">
        From the dashboard, choose
        <strong class="font-medium text-foreground">New instance</strong> and set:
      </p>
      <ul class="list-disc space-y-1.5 pl-6 text-sm text-muted-foreground">
        <li><strong class="font-medium text-foreground">PostgreSQL version</strong></li>
        <li><strong class="font-medium text-foreground">Instances</strong> - number of replicas</li>
        <li>
          <strong class="font-medium text-foreground">Storage</strong> - e.g.
          <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-sm">2Gi</code>
        </li>
      </ul>
      <p class="leading-7 text-muted-foreground">
        Your instance starts in
        <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-sm">Pending</code>
        and moves to
        <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-sm">Ready</code>
        once provisioning finishes - the page updates automatically, no need to refresh.
      </p>
    </div>

    <div class="space-y-2">
      <h2 class="text-xl font-semibold tracking-tight">3. Connect</h2>
      <p class="leading-7 text-muted-foreground">
        Open the instance and go to
        <strong class="font-medium text-foreground">Credentials</strong> for a ready-to-use
        connection string, or use the built-in
        <strong class="font-medium text-foreground">Query console</strong> to run SQL directly in
        the browser - no local client required.
      </p>
    </div>

    <div class="space-y-2">
      <h2 class="text-xl font-semibold tracking-tight">4. Using the API or SDK instead</h2>
      <p class="leading-7 text-muted-foreground">
        Everything in the UI is also available over the REST API - see
        <RouterLink :to="{ name: 'docs-api' }" class="font-medium text-foreground underline underline-offset-4">
          API reference
        </RouterLink>
        - or the
        <RouterLink :to="{ name: 'docs-sdk' }" class="font-medium text-foreground underline underline-offset-4">
          Go SDK
        </RouterLink>
        for programmatic use.
      </p>
    </div>
  </article>
</template>