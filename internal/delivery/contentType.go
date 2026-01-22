package delivery

type ContentType string

var (
	ApplictionJSON        ContentType = "application/json"
	ApplcationOctetStream ContentType = "application/octet-stream"
	TextHTML              ContentType = "text/html"
	Default                           = TextHTML
)
