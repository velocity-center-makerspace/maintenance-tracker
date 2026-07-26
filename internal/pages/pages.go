package pages

import (
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/pages/assets"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/pages/calendar"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/pages/tasks"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/router"
)

type Registrar interface {
	RegisterRoutes(r router.Router)
}

type registrar struct {
	pages []Registrar
}

func NewRegistrar() Registrar {
	return &registrar{}
}

func (reg *registrar) RegisterRoutes(r router.Router) {
	reg.pages = append(
		reg.pages,
		assets.AssetsPage{},
		calendar.CalendarPage{},
		tasks.TasksPage{},
	)

	for _, page := range reg.pages {
		page.RegisterRoutes(r)
	}
}
