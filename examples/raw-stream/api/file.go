// Package api declares the upload/download contract used by the runnable example
// and the framework's contract, HTTP and CLI acceptance tests.
package api

import (
	"io"

	"github.com/graingo/maltose/util/mmeta"
)

type UploadReq struct {
	mmeta.Meta `method:"PUT" path:"/files/:id" operation_id:"file.upload" body:"raw" consumes:"application/octet-stream"`
	ID         string        `path:"id" binding:"uuid"`
	Body       io.ReadCloser `stream:"body" field:"required" dc:"File bytes, including an empty file"`
}
type UploadRes struct {
	Size int64 `json:"size" field:"required"`
}

type DownloadReq struct {
	mmeta.Meta `method:"GET" path:"/files/:id" operation_id:"file.download" response:"raw" produces:"application/octet-stream"`
	ID         string `path:"id" binding:"uuid"`
}
type DownloadRes struct {
	Body io.ReadCloser `stream:"body"`
}
