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
		} else if status := r.Writer.Status(); status != http.StatusOK {
			// handle HTTP status code error
			msg = http.StatusText(status)
			switch status {
			case http.StatusNotFound:
				code = mcode.CodeNotFound
			case http.StatusForbidden:
				code = mcode.CodeForbidden
			case http.StatusUnauthorized:
				code = mcode.CodeNotAuthorized
			default:
				code = mcode.CodeInternalError
			}
			data = nil
			// create error object for other middleware usage
			r.Error(merror.NewCode(code, msg))
		} else {
			msg = code.Message()
		}

		// return standard response
		r.JSON(r.Writer.Status(), DefaultResponse{
			Code:    code.Code(),
			Message: msg,
			Data:    data,
		})
	}
}
