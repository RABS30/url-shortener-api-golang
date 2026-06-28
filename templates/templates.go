package templates

import "embed"

//go:embed mail/*.html
var EmailTemplatesFS embed.FS