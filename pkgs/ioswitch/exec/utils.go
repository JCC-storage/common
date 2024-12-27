package exec

import (
	"github.com/google/uuid"
)

func genRandomPlanID() PlanID {
	return PlanID(uuid.NewString())
}
