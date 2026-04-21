package validation

import "github.com/pupload/pupload/internal/models"

type ValidationSeverity string

const (
	ValidationError   ValidationSeverity = "ValidationError"
	ValidationWarning ValidationSeverity = "ValidationWarning"
)

type ValidationEntry struct {
	Type        ValidationSeverity
	Code        string
	Name        string
	Description string
}

type ValidationResult struct {
	Errors   []ValidationEntry
	Warnings []ValidationEntry
}

func Validate(flow models.Flow, tasks []models.Task) *ValidationResult {
	res := &ValidationResult{}

	// Store errors and warnings
	for _, store := range flow.Stores {
		storeInvalidType(res, store)
	}

	// Step errors and warnings
	stepDuplicateID(res, flow.Steps)
	for _, step := range flow.Steps {
		stepNoTaskFound(res, step, tasks)
		stepMissingInput(res, step, tasks)
		// stepMissingOutput(res, step, tasks)
		stepInvalidTier(res, step, tasks)
		stepMissingFlag(res, step, tasks)
		stepUnknownFlag(res, step, tasks)
		stepMissingID(res, step)
	}

	// Edge errors and warnings
	edgeNoConsumers(res, flow)
	edgeNoProducers(res, flow)
	edgeTypeMismatch(res, flow, tasks)

	// Well errors and warnings
	wellDuplicateEdge(res, flow.DataWells)
	wellEdgeNotFound(res, flow.DataWells, flow.Steps)
	for _, well := range flow.DataWells {
		wellInvalidSource(res, well)
		wellStoreNotFound(res, well, flow.Stores)
		wellInvalidKeyTemplate(res, well)
		wellStaticMissingKey(res, well)
		wellStaticKeyIsDynamic(res, well)
		wellDynamicKeyIsStatic(res, well)
	}

	// Flow errors and warnings
	flowDetectEmpty(res, flow)
	flowDetectCycle(res, flow)
	flowValidateTimeout(res, flow)

	return res
}

func (r *ValidationResult) HasError() bool {
	return len(r.Errors) > 0
}

func (r *ValidationResult) HasWarnings() bool {
	return len(r.Errors)+len(r.Warnings) > 0
}

func (r *ValidationResult) AddError(entry ValidationEntry) {
	r.Errors = append(r.Errors, entry)
}

func (r *ValidationResult) AddWarning(entry ValidationEntry) {
	r.Warnings = append(r.Warnings, entry)
}
