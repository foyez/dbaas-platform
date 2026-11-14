// Package docs embeds the API's OpenAPI description and a Swagger UI page,
// so every running API serves its own reference at /docs.
package docs

import _ "embed"

// OpenAPI is the OpenAPI 3 description of the REST API.
//
//go:embed openapi.yaml
var OpenAPI []byte

// SwaggerUI is an HTML page that renders OpenAPI with Swagger UI.
//
//go:embed swagger.html
var SwaggerUI []byte
