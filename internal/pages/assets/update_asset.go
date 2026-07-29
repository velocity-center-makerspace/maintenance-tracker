package assets

import (
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/response"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/router"
	"net/http"
)

func RegisterUpdateAsset(r router.Router) {
	r.AddRoute("PATCH /assets/name", UpdateAssetName)
	r.AddRoute("PATCH /assets/warranty-expiry", UpdateAssetWarrantyExpiry)
	r.AddRoute("PATCH /assets/status", UpdateAssetStatus)
	r.AddRoute("PATCH /assets/end-of-life", UpdateAssetEOL)
	r.AddRoute("PUT /assets", UpdateAsset)
}

func UpdateAssetName(deps router.Dependencies, w http.ResponseWriter, r *http.Request) {

}

func UpdateAssetWarrantyExpiry(deps router.Dependencies, w http.ResponseWriter, r *http.Request) {

}

func UpdateAssetStatus(deps router.Dependencies, w http.ResponseWriter, r *http.Request) {

}

func UpdateAssetEOL(deps router.Dependencies, w http.ResponseWriter, r *http.Request) {

}

// TODO: figure out how to update asset files and files
func UpdateAsset(deps router.Dependencies, w http.ResponseWriter, r *http.Request) {

}
