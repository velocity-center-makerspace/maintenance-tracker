package assets

import (
	"errors"
	"fmt"
)

const (
	OPERATIONAL       = "operational"
	DOWN              = "down"
	UNAVAILABLE       = "unavailable"
	END_OF_LIFE       = "end_of_life"
	NEEDS_MAINTENANCE = "needs_maintenance"
	NEEDS_REPAIR      = "needs_repair"
	NONE              = "none"
)

func checkAvailability(availability string) error {
	switch availability {
	case OPERATIONAL:
		return nil
	case DOWN:
		return nil
	case UNAVAILABLE:
		return nil
	default:
		return fmt.Errorf(
			"%w: %s",
			errors.New("ineligible availability string"),
			availability,
		)
	}
}

func checkAttentionNeeded(attentionNeeded string) error {
	switch attentionNeeded {
	case NEEDS_MAINTENANCE:
		return nil
	case NEEDS_REPAIR:
		return nil
	case END_OF_LIFE:
		return nil
	case NONE:
		return nil
	default:
		return fmt.Errorf(
			"%w: %s",
			errors.New("ineligible attention_needed string"),
			attentionNeeded,
		)
	}
}

func validateAssetStatus(availability, attentionNeeded string) error {
	if availability == OPERATIONAL && attentionNeeded == NEEDS_REPAIR {
		return errors.New("asset cannot be operational status while in need of repair")
	}

	if availability == UNAVAILABLE && attentionNeeded == NEEDS_REPAIR {
		return errors.New(
			"asset cannot be unavailable status while in need of repair; use down instead",
		)
	}

	return nil
}
