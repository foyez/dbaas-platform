<template>
  <article class="space-y-6">
    <h1 class="text-3xl font-bold tracking-tight">Architecture</h1>

    <p class="leading-7 text-muted-foreground">At a high level, a request flows like this:</p>

    <ol class="list-decimal space-y-2 pl-6 text-sm text-muted-foreground">
      <li>
        You interact with the
        <strong class="font-medium text-foreground">Web UI</strong> (Vue), which authenticates
        you against <strong class="font-medium text-foreground">ZITADEL</strong> (OIDC) directly.
      </li>
      <li>
        The UI calls the <strong class="font-medium text-foreground">API</strong> (Go), sending
        your access token on every request. The API validates it against ZITADEL via token
        introspection.
      </li>
      <li>
        Creating, updating, or deleting an instance writes a
        <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">PostgreSQL</code>
        custom resource to Kubernetes.
      </li>
      <li>
        A custom <strong class="font-medium text-foreground">operator</strong> reconciles that
        resource into a <strong class="font-medium text-foreground">CloudNativePG</strong>
        cluster, and allocates an external port so the instance is reachable outside the cluster.
      </li>
      <li>
        Logs and metrics from the API and the cluster feed into
        <strong class="font-medium text-foreground">Prometheus, Loki, and Grafana</strong> - the
        same pipeline powers both the admin dashboards and the in-app audit log.
      </li>
    </ol>

    <div class="space-y-2">
      <h2 class="text-xl font-semibold tracking-tight">Why a custom operator?</h2>
      <p class="leading-7 text-muted-foreground">
        The API only ever needs to know a simplified shape - version, instance count, storage
        size. It never depends on CloudNativePG's own resource schema directly; the operator is
        the translation layer between the two.
      </p>
    </div>

    <div class="space-y-2">
      <h2 class="text-xl font-semibold tracking-tight">Delivery</h2>
      <p class="leading-7 text-muted-foreground">
        Every component - the operator, the API, and the UI - is built and deployed independently
        through GitOps: a push builds an image, updates a manifest, and ArgoCD reconciles the
        cluster to match.
      </p>
    </div>
  </article>
</template>