package assets

import (
	// "net/http/httptest"
	"testing"
)

func TestDeleteAsset(t *testing.T) {
	// assign: need dependencies - *sql.DB, *http.ServeMux, Environment
	// assign: need a dummy router
	// assign: add route to dummy router
	// assign: need a dummy server
	// assign: need to add assets, some with asset_files & files, some without to data/dev.db

	// act: send requests to dummy server for DELETE /assets

	// assert: handle errors
	// assert: ensure correct headers and response
}

func TestDeleteFileFromDisk(t *testing.T) {
	// assign: create temp files
	// assign: create some non-existent file paths

	// act: call deleteFileFromDisk & capture err for each file

	// assert: err response is as expected
}
