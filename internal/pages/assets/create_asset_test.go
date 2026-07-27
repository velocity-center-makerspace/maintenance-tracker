package assets

import (
	"net/http/httptest"
	"testing"
)

// integration test
func TestCreateAsset(t *testing.T) {
	// assign: need dependencies - *sql.DB, *http.ServeMux, Environment
	// assign: need a dummy router
	// assign: add route to dummy router
	// assign: need a dummy server
	// assign: need some multipart/form-data type requests
	// assign: create expected response for each case

	// act: send requests to dummy server for POST /assets

	// assert: handle errors
	// assert: correct headers and response
}

func TestParseParts(t *testing.T) {
	// assign: create some multipart/form-data request bodies
	// assign: instantiate parsePartsParams w/ *multipart.Reader
	// *db.CreateAssetParams, and dstRoot
	// assign: create expected upload objects for each case

	// act: pass params to parseParts

	// assert: return is as expected for []upload and err
}

func TestProcessFile(t *testing.T) {
	// assign: create temporary files
	// assign: create upload objects
	// assign: create error channels
	// assign: create processFileParams objects
	// assign: create expected fileMetadata objects for each case

	// act: send the processFileParams objects for processing

	// assert: fileMetadata returned matches expected
}
