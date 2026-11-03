package attachments

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

const sessionAuth = "session cookie or bearer token"

var unauthenticated = documentation.Error{Status: 401, Code: "unauthenticated", Description: "No valid session."}
var csrf = documentation.Error{Status: 403, Code: "permission_denied", Description: "A cookie-authenticated request without the X-Facile-CSRF header."}
var malformed = documentation.Error{Status: 400, Code: "invalid_argument", Description: "Not a multipart body, or no file field."}
var tooLarge = documentation.Error{Status: 413, Code: "resource_exhausted", Description: "The file is larger than 20 MB."}
var unsupported = documentation.Error{Status: 415, Code: "unsupported_media_type", Description: "Not an image (png, jpeg, gif, webp), a PDF or UTF-8 text. SVG is refused."}
var notFound = documentation.Error{Status: 404, Code: "not_found", Description: "No attachment with this id belongs to the user."}

var Documentation = documentation.Module{
	Name:        "attachments",
	Description: "Files the user attaches to a message. The model sees images, PDFs and text files; the transcript keeps only a reference.",
	Routes: []documentation.Route{
		{
			Method:       "POST",
			Path:         "/attachments",
			Summary:      "Upload an attachment",
			Description:  "Stores one file sent as the multipart field file. The server detects the type from the bytes and the file name and ignores the browser's label. Pass the returned id in a message's attachments.",
			Auth:         sessionAuth,
			RequestBody:  UploadRequest{},
			ResponseBody: AttachmentResponse{},
			Status:       201,
			Errors:       []documentation.Error{unauthenticated, csrf, malformed, tooLarge, unsupported},
		},
		{
			Method:       "GET",
			Path:         "/attachments/{id}",
			Summary:      "Download an attachment",
			Description:  "Answers the file's bytes to its owner. Text files are served as text/plain, so they never render as a page. Works with the session cookie, so an <img src> can load it.",
			Auth:         sessionAuth,
			PathParams:   []documentation.Field{{Name: "id", Type: "string", Description: "Attachment id"}},
			ResponseBody: "the file's bytes",
			Errors:       []documentation.Error{unauthenticated, notFound},
		},
	},
}
