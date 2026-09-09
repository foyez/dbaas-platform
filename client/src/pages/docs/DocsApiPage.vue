<template>
  <article class="space-y-6">
    <h1 class="text-3xl font-bold tracking-tight">API reference</h1>

    <p class="leading-7 text-muted-foreground">
      All endpoints are under
      <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-sm">/api/v1</code>
      and require a Bearer token, obtained the same way the web UI signs in (OIDC). This page
      covers the shape of the API - see the OpenAPI spec in the repository for full
      request/response schemas.
    </p>

    <div class="overflow-x-auto rounded-lg border">
      <table class="w-full text-left text-sm">
        <thead class="bg-muted/50">
          <tr>
            <th class="px-4 py-2 font-medium">Method</th>
            <th class="px-4 py-2 font-medium">Path</th>
            <th class="px-4 py-2 font-medium">Description</th>
          </tr>
        </thead>
        <tbody class="divide-y">
          <tr>
            <td class="px-4 py-2 font-mono text-xs">POST</td>
            <td class="px-4 py-2 font-mono text-xs">/instances</td>
            <td class="px-4 py-2 text-muted-foreground">
              Create an instance. Requires an
              <code class="rounded bg-muted px-1 py-0.5 font-mono text-xs">Idempotency-Key</code>
              header.
            </td>
          </tr>
          <tr>
            <td class="px-4 py-2 font-mono text-xs">GET</td>
            <td class="px-4 py-2 font-mono text-xs">/instances</td>
            <td class="px-4 py-2 text-muted-foreground">
              List your instances. Cursor-paginated:
              <code class="rounded bg-muted px-1 py-0.5 font-mono text-xs">?limit=&cursor=</code>,
              returns
              <code class="rounded bg-muted px-1 py-0.5 font-mono text-xs">nextCursor</code>.
            </td>
          </tr>
          <tr>
            <td class="px-4 py-2 font-mono text-xs">GET</td>
            <td class="px-4 py-2 font-mono text-xs">/instances/:id</td>
            <td class="px-4 py-2 text-muted-foreground">Get a single instance.</td>
          </tr>
          <tr>
            <td class="px-4 py-2 font-mono text-xs">PATCH</td>
            <td class="px-4 py-2 font-mono text-xs">/instances/:id</td>
            <td class="px-4 py-2 text-muted-foreground">Update an instance.</td>
          </tr>
          <tr>
            <td class="px-4 py-2 font-mono text-xs">DELETE</td>
            <td class="px-4 py-2 font-mono text-xs">/instances/:id</td>
            <td class="px-4 py-2 text-muted-foreground">Delete an instance.</td>
          </tr>
          <tr>
            <td class="px-4 py-2 font-mono text-xs">GET</td>
            <td class="px-4 py-2 font-mono text-xs">/instances/:id/credentials</td>
            <td class="px-4 py-2 text-muted-foreground">Connection credentials.</td>
          </tr>
          <tr>
            <td class="px-4 py-2 font-mono text-xs">GET</td>
            <td class="px-4 py-2 font-mono text-xs">/instances/:id/logs</td>
            <td class="px-4 py-2 text-muted-foreground">Recent logs for the instance.</td>
          </tr>
          <tr>
            <td class="px-4 py-2 font-mono text-xs">POST</td>
            <td class="px-4 py-2 font-mono text-xs">/instances/:id/query</td>
            <td class="px-4 py-2 text-muted-foreground">
              Run a single SQL statement. Body:
              <code class="rounded bg-muted px-1 py-0.5 font-mono text-xs">{ "sql": "..." }</code>
            </td>
          </tr>
          <tr>
            <td class="px-4 py-2 font-mono text-xs">GET</td>
            <td class="px-4 py-2 font-mono text-xs">/audit-logs</td>
            <td class="px-4 py-2 text-muted-foreground">
              Admin only. Cursor-paginated:
              <code class="rounded bg-muted px-1 py-0.5 font-mono text-xs"
                >?resourceId=&amp;limit=&amp;cursor=</code
              >.
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </article>
</template>