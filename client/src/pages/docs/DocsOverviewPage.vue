<template>
  <article class="space-y-6">
    <h1 class="text-3xl font-bold tracking-tight">Overview</h1>

    <p class="leading-7 text-muted-foreground">
      This platform lets you provision and manage PostgreSQL databases on Kubernetes without
      touching Kubernetes yourself. Under the hood, every instance you create is a
      <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-sm">PostgreSQL</code>
      custom resource, reconciled by a purpose-built operator into a
      <a
        href="https://cloudnative-pg.io"
        target="_blank"
        rel="noopener"
        class="font-medium text-foreground underline underline-offset-4"
        >CloudNativePG</a
      >
      cluster.
    </p>

    <div class="space-y-3">
      <h2 class="text-xl font-semibold tracking-tight">What you get</h2>
      <ul class="list-disc space-y-1.5 pl-6 text-sm text-muted-foreground">
        <li>A managed PostgreSQL instance, reachable from outside the cluster</li>
        <li>Connection credentials and a browser-based SQL query console</li>
        <li>Per-instance logs and an admin audit trail</li>
        <li>A REST API and a Go SDK for programmatic access</li>
      </ul>
    </div>

    <div class="space-y-3">
      <h2 class="text-xl font-semibold tracking-tight">How it fits together</h2>
      <p class="leading-7 text-muted-foreground">
        The web UI and the API both authenticate through OIDC. Creating an instance from the UI
        (or the API, or the SDK) submits a
        <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-sm">PostgreSQL</code>
        custom resource; the operator takes it from there - provisioning, health checks, and
        external connectivity are all handled automatically.
      </p>
    </div>

    <p class="text-sm text-muted-foreground">
      See
      <RouterLink :to="{ name: 'docs-architecture' }" class="font-medium text-foreground underline underline-offset-4">
        Architecture
      </RouterLink>
      for the full picture, or jump straight to
      <RouterLink :to="{ name: 'docs-getting-started' }" class="font-medium text-foreground underline underline-offset-4">
        Getting started
      </RouterLink>
      to create your first instance.
    </p>
  </article>
</template>