# Raw HTTP upload and download

This example belongs to the unreleased `feat/raw-http-stream` branch. It uses the
root module directly; framework and CLI version constants remain v0.5.1 until a
release is authorized. The new feature is intended for the next minor release.

From the Maltose repository root, build the CLI and generate the documents:

```bash
(cd cmd/maltose && go build -o /tmp/maltose-raw-cli .)
/tmp/maltose-raw-cli gen openapi -s examples/raw-stream/api \
  -o examples/raw-stream/openapi.yaml --openapi-version 3.1.0 \
  --extensions github.com/graingo/maltose/examples/raw-stream/extension
/tmp/maltose-raw-cli gen openapi -s examples/raw-stream/api \
  -o examples/raw-stream/openapi.yaml --openapi-version 3.1.0 \
  --extensions github.com/graingo/maltose/examples/raw-stream/extension --check
RAW_STREAM_TOKEN=local-example go run ./examples/raw-stream
```

In another terminal:

```bash
python3 -c 'from pathlib import Path; Path("/tmp/stream-input.bin").write_bytes(bytes([0, 255, 254, 1]) * 1024)'
curl --fail-with-body -X PUT \
  -H 'Authorization: Bearer local-example' \
  -H 'Content-Type: application/octet-stream' \
  --data-binary @/tmp/stream-input.bin \
  http://127.0.0.1:8080/files/ccba9e91-bfa0-468e-91ac-956ee9097271
curl --fail-with-body \
  -H 'Authorization: Bearer local-example' \
  http://127.0.0.1:8080/files/ccba9e91-bfa0-468e-91ac-956ee9097271 \
  -o /tmp/stream-output.bin
cmp /tmp/stream-input.bin /tmp/stream-output.bin
```

The server authenticates before binding or reading, limits uploads to 32 MiB, and
sets 30-second read/write and context deadlines. Files live in a temporary directory
and are removed at shutdown. Uploads become visible after the entire copy succeeds.
The response size and raw download bytes come from the same Req/Res definitions
used to generate OpenAPI 3.1 and the manifest. Startup verifies those artifacts.

`body:"raw"` plus an `io.ReadCloser` field tagged `stream:"body"` declares an upload.
`field:"required"` requires presence and permits zero bytes. `consumes` defaults to
`application/octet-stream`. `response:"raw"` plus the response's stream field
selects a download; `produces` declares its media type. Both modes can be combined.
Path/query/header rules remain available. Streaming conflicts with JSON/form
body fields, success envelopes, HEAD and 204 responses.

Incoming bodies belong to net/http and are consumed synchronously. Returning a
successful output stream transfers ownership to Maltose, which closes it once.
Close must be safe alongside Read and unblock it for cancellation. Errors before
the first byte reach application error middleware; errors after response commit
are recorded as `*mhttp.StreamError` and terminate the transport. Error middleware
checks `Writer.Written()` before writing JSON. Preparation failures restore headers
to their state before the controller ran, preserving middleware metadata.
Read cancellation and deadlines remain application policy; the example shows
how to coordinate deadline callbacks before the request ends.

This is a transport example. Durable storage, upload intents, authorization per
file, concurrency policy and evidence associations belong to the application.

Acceptance tests:

```bash
go test -race ./net/mhttp/... ./examples/raw-stream/...
(cd cmd/maltose && GOWORK=off go test -race ./internal/openapi)
```

The HTTP tests cover binary and empty content, zero-read binding/authentication,
chunked uploads, incremental downloads, duplex transfer, limits/deadlines,
cancellation, pre/post-commit failures and closure. The contract and CLI tests
cover both OpenAPI versions, deterministic output and stale transport detection.
