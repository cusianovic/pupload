package models

// Defines where data should be sourced from, stored, and outputted to on a given edge
type DataWell struct {
	Edge  string // Edge name
	Store string // Store name

	Source   *string           // Where we're sourcing data from on input: upload, webhook, static, etc.
	Key      *string           // defaults to artifact id. ${RUN_ID}, ${ARTIFACT_ID}, ${NODE_ID}
	Lifetime *DataWellLifetime // how long the data on that edge should live for
}

type DataWellLifetime struct {
}
