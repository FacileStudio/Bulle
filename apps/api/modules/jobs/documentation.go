package jobs

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

const sessionAuth = "session cookie or bearer token"

type EmptyRequest struct{}

type EmptyResponse struct{}

var unauthenticated = documentation.Error{Status: 401, Code: "unauthenticated", Description: "No valid session."}
var notFound = documentation.Error{Status: 404, Code: "not_found", Description: "The requested job was not found."}
var csrf = documentation.Error{Status: 403, Code: "permission_denied", Description: "A cookie-authenticated request without the X-Facile-CSRF header."}
var badRequest = documentation.Error{Status: 400, Code: "invalid_argument", Description: "Malformed request payload."}

var Documentation = documentation.Module{
	Name:        "jobs",
	Description: "Scheduled background jobs that execute prompts according to cron schedules.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/jobs",
			Summary:      "List jobs",
			Description:  "Lists all background jobs owned by the authenticated user.",
			Auth:         sessionAuth,
			ResponseBody: []JobResponse{},
			Errors:       []documentation.Error{unauthenticated},
		},
		{
			Method:       "POST",
			Path:         "/jobs",
			Summary:      "Create a job",
			Description:  "Creates a background job with a cron schedule.",
			Auth:         sessionAuth,
			RequestBody:  CreateJobRequest{},
			ResponseBody: JobResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, badRequest},
		},
		{
			Method:       "GET",
			Path:         "/jobs/{id}",
			Summary:      "Get a job",
			Description:  "Reads a single background job by ID.",
			Auth:         sessionAuth,
			ResponseBody: JobResponse{},
			Errors:       []documentation.Error{unauthenticated, notFound},
		},
		{
			Method:       "PUT",
			Path:         "/jobs/{id}",
			Summary:      "Update a job",
			Description:  "Updates an existing background job.",
			Auth:         sessionAuth,
			RequestBody:  UpdateJobRequest{},
			ResponseBody: JobResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, notFound, badRequest},
		},
		{
			Method:       "DELETE",
			Path:         "/jobs/{id}",
			Summary:      "Delete a job",
			Description:  "Deletes a background job and all its runs.",
			Auth:         sessionAuth,
			ResponseBody: EmptyResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, notFound},
		},
		{
			Method:       "POST",
			Path:         "/jobs/{id}/run",
			Summary:      "Trigger a job run",
			Description:  "Manually triggers an immediate run of a background job.",
			Auth:         sessionAuth,
			RequestBody:  EmptyRequest{},
			ResponseBody: RunJobResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, notFound},
		},
		{
			Method:       "GET",
			Path:         "/jobs/{id}/runs",
			Summary:      "List job runs",
			Description:  "Lists all execution runs for a background job.",
			Auth:         sessionAuth,
			ResponseBody: []JobRunResponse{},
			Errors:       []documentation.Error{unauthenticated, notFound},
		},
	},
}
