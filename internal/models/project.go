package models

type Project struct {
	TenantID    string
	ProjectName string
	Version     int

	Flows        []Flow
	NodeDefs     []NodeDef
	GlobalStores []StoreInput
}
