package mhttp

import (
	"net/http"

	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"
)

// DefaultResponse standard response structure
type DefaultResponse struct {
	Code    int    `json:"code"`    // business code
	Message string `json:"message"` // prompt information
	Data    any    `json:"data"`    // business data
}

func codeToHTTPStatus(code mcode.Code) int {
	switch code {
	case mcode.CodeOK:
		return http.StatusOK
	case mcode.CodeValidationFailed:
		return http.StatusBadRequest
	case mcode.CodeNotFound:
		return http.StatusNotFound
	case mcode.CodeNotAuthorized:
		return http.StatusUnauthorized
	case mcode.CodeForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// MiddlewareResponse standard response middleware
func MiddlewareResponse() MiddlewareFunc {
	return func(r *Request) {
		r.Next()

		// if response has been written, skip
		if r.Writer.Written() {
			return
		}

		var (
			msg  string
			code mcode.Code = mcode.CodeOK
			data            = r.GetHandlerResponse()
		)

		// handle error case
		if len(r.Errors) > 0 {
			err := r.Errors.Last().Err
			// get error code
			code = merror.Code(err)
			if code == mcode.CodeNil {
				code = mcode.CodeInternalError
			}
			msg = err.Error()
			data = nil
		} else {
			msg = code.Message()
		}

		// return standard response
		httpStatus := codeToHTTPStatus(code)
		r.JSON(httpStatus, DefaultResponse{
			Code:    code.Code(),
			Message: msg,
			Data:    data,
		})
	}
}
