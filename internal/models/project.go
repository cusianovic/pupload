package models

import "github.com/google/uuid"

type Project struct {
	ID uuid.UUID // Controller scoped, should not be set

	Flows        []Flow
	NodeDefs     []NodeDef
	GlobalStores []StoreInput
}
