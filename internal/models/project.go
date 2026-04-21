package models

import "github.com/google/uuid"

type Project struct {
	ID uuid.UUID // Controller scoped, should not be set

	Flows        []Flow
	Tasks        []Task
	GlobalStores []StoreInput
}
