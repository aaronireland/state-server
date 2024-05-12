package api

import (
	"net/http"

	"github.com/go-chi/render"
)

// Schema for any non-200 HTTP response
type ErrorResponse struct {
	Err            error  `json:"-"`
	StatusText     string `json:"status"`
	ErrorText      string `json:"error,omitempty"`
	HTTPStatusCode int    `json:"-"`
}

// render method which hooks into the go-chi renderer and ensures the correct
// HTTP response status code is writtern to the response
func (e *ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// creates the go-chi renderer for HTTP 404 responses
func NotFoundError(err error) render.Renderer {
	return &ErrorResponse{
		Err:            err,
		HTTPStatusCode: http.StatusNotFound,
		StatusText:     "Not Found",
		ErrorText:      err.Error(),
	}
}

// creates the go-chi renderer for HTTP 500 responses
func InternalServerError(err error) render.Renderer {
	return &ErrorResponse{
		Err:            err,
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     "Internal Server Error",
		ErrorText:      err.Error(),
	}
}

// creates the go-chi renderer for HTTP 400 responses
func BadRequestError(err error) render.Renderer {
	return &ErrorResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "Bad Request",
		ErrorText:      err.Error(),
	}
}
