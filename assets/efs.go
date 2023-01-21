package assets

import "embed"

//go:embed "swagger-ui" "paydex.swagger.json"
var EmbeddedFiles embed.FS
