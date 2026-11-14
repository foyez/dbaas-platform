---
hide:
  - toc
---

# API reference

The full REST API, rendered from [`platform/docs/openapi.yaml`](https://github.com/foyez/dbaas-platform/blob/main/platform/docs/openapi.yaml). It's read-only here, because the production API is shut down.

To try requests against a running API, start the platform with `make dev` and open <http://api.dbaas.localhost/docs>. The API serves the same reference there with "Try it out" enabled. Click **Authorize** and paste an access token; see [Go SDK](components/sdk.md#try-it-against-the-local-platform) for how to get one.

<div id="swagger-ui"></div>

<link rel="stylesheet"
      href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.34.0/swagger-ui.css"
      integrity="sha384-4c3vsE5f+trhFyHrFrLM1fFGUn531NIaC7dcqEWIJdoRQJ5CkJbSuSCwXLpaS1KF"
      crossorigin="anonymous">
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.34.0/swagger-ui-bundle.js"
        integrity="sha384-9Td7MKiIr2S83E1kZyyVi8KdJw/S8CpAvu0llLyMeg0xq/hDGiZJfVn0fV4llCmK"
        crossorigin="anonymous"></script>
<script>
  SwaggerUIBundle({
    url: "../openapi.yaml",
    dom_id: "#swagger-ui",
    deepLinking: true,
    // Read-only: there's no public API to send requests to.
    supportedSubmitMethods: [],
  });
</script>
