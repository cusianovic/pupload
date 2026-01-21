package validation

import (
	"fmt"
	"slices"

	"github.com/pupload/pupload/internal/models"
)

func storeInvalidType(r *ValidationResult, store models.StoreInput) {
	validTypes := []string{"s3"}

	if !slices.Contains(validTypes, store.Type) {
		r.AddError(ValidationEntry{
			ValidationError,
			ErrStoreInvalidType,
			"StoreInvalidType",
			fmt.Sprintf("Store %s has invalid type", store.Name),
		})
	}
}
