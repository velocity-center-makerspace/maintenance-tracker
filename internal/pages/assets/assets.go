package assets

import (
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/router"
)

type AssetsPage struct{}

func (a AssetsPage) RegisterRoutes(r router.Router) {
	RegisterCreateAsset(r)
}
