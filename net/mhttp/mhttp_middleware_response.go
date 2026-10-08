package mhttp

import (
	"errors"
	"net/http"

	"github.com/graingo/maltose/net/mhttp/contract"

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

		if r.operation != nil && len(r.Errors) == 0 {
			writeContractResponse(r)
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
		if len(r.Errors) > 0 {
			var decode *contract.DecodeError
			if errors.As(r.Errors.Last().Err, &decode) {
				httpStatus = decode.Status
			}
		}
		r.JSON(httpStatus, DefaultResponse{
			Code:    code.Code(),
			Message: msg,
			Data:    data,
		})
	}
}

func writeContractResponse(r *Request) {
	status := r.SuccessStatus()
	if status == http.StatusNoContent || r.Request.Method == http.MethodHead {
		r.Status(status)
		r.Writer.WriteHeaderNow()
		return
	}
	data := r.GetHandlerResponse()
	if r.operation.Envelope == "maltose" {
		r.JSON(status, DefaultResponse{Code: 0, Message: mcode.CodeOK.Message(), Data: data})
		return
	}
	r.JSON(status, data)
}
