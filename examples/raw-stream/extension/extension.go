package extension

import "github.com/graingo/maltose/net/mhttp/contract"

type ErrorBody struct {
	Error string `json:"error"`
}

func Configure(e *contract.Extensions) error {
	if err := e.SecurityScheme("bearer", map[string]any{"type": "http", "scheme": "bearer"}); err != nil {
		return err
	}
	for _, id := range []string{"file.upload", "file.download"} {
		if err := e.Security(id, []map[string][]string{{"bearer": {}}}); err != nil {
			return err
		}
		for _, status := range []int{400, 401, 404, 408, 413, 415, 500} {
			if err := e.ErrorResponse(id, status, "Request failed", "application/json", contract.TypeOf[ErrorBody]()); err != nil {
				return err
			}
		}
	}
	for _, name := range []string{"Content-Disposition", "Content-Length"} {
		if err := e.ResponseHeader("file.download", name, "Download metadata", map[string]any{"type": "string"}); err != nil {
			return err
		}
	}
	return nil
}
