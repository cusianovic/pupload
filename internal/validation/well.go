package validation

import (
	"fmt"
	"os"
	"slices"

	"github.com/pupload/pupload/internal/models"
)

func wellInvalidSource(r *ValidationResult, well models.DataWell) {
	if well.Source == nil {
		return
	}

	allowedSources := []string{"upload", "static", "webhook"}
	if !slices.Contains(allowedSources, *well.Source) {
		r.AddError(ValidationEntry{
			ValidationError,
			ErrDatawellInvalidSource,
			"DatawellInvalidSource",
			fmt.Sprintf("Source %s is not a valid datawell source (datawell with edge %s)", *well.Source, well.Edge),
		})
	}
}

func wellEdgeNotFound(r *ValidationResult, wells []models.DataWell, nodes []models.Node) {
	edgeSet := make(map[string]struct{})
	for _, node := range nodes {
		for _, in := range node.Inputs {
			edgeSet[in.Edge] = struct{}{}
		}

		for _, out := range node.Outputs {
			edgeSet[out.Edge] = struct{}{}
		}
	}

	for _, well := range wells {
		if _, ok := edgeSet[well.Edge]; !ok {
			r.AddError(ValidationEntry{
				ValidationError,
				ErrDatawellEdgeNotFound,
				"DatawellEdgeNotFound",
				fmt.Sprintf("datawell has edge %s, but edge is not used in nodes", well.Edge),
			})
		}
	}

}

func wellDuplicateEdge(r *ValidationResult, wells []models.DataWell) {
	edgeCount := make(map[string]int)
	for _, well := range wells {
		edgeCount[well.Edge]++
	}

	for edge, count := range edgeCount {
		if count > 1 {
			r.AddError(ValidationEntry{
				ValidationError,
				ErrDatawellDuplicateEdge,
				"DatawellDuplicateEdge",
				fmt.Sprintf("Datawell edge %s is used more than once", edge),
			})
		}
	}
}

func wellStoreNotFound(r *ValidationResult, well models.DataWell, stores []models.StoreInput) {
	storesSet := make(map[string]struct{})

	for _, store := range stores {
		storesSet[store.Name] = struct{}{}
	}

	if _, ok := storesSet[well.Store]; !ok {
		r.AddError(ValidationEntry{
			ValidationError,
			ErrDatawellStoreNotFound,
			"DatawellStoreNotFound",
			fmt.Sprintf("Datawell references nonexistant store %s", well.Store),
		})
	}

}

var allowedKeyVars = map[string]struct{}{
	"RUN_ID": {}, "FLOW_NAME": {}, "EDGE": {},
	"TIMESTAMP": {}, "DATE": {}, "YEAR": {}, "MONTH": {}, "DAY": {},
	"UUID": {},
}

func wellInvalidKeyTemplate(r *ValidationResult, well models.DataWell) {
	if well.Key == nil {
		return
	}

	os.Expand(*well.Key, func(s string) string {
		if _, ok := allowedKeyVars[s]; !ok {
			r.AddError(ValidationEntry{
				ValidationError,
				ErrDatawellInvalidKeyTemplate,
				"DatawellInvalidKeyTemplate",
				fmt.Sprintf("Unknown variable ${%s} in key template for edge %s", s, well.Edge),
			})
		}
		return ""
	})
}

func wellDynamicKeyIsStatic(r *ValidationResult, well models.DataWell) {
	if well.Key == nil || well.Source == nil {
		return
	}

	// Only check dynamic sources
	if *well.Source == "static" {
		return
	}

	// Check if key contains any variables
	hasVariable := false
	os.Expand(*well.Key, func(s string) string {
		if _, ok := allowedKeyVars[s]; ok {
			hasVariable = true
		}

		return ""
	})

	if !hasVariable {
		r.AddError(ValidationEntry{
			ValidationError,
			ErrDatawellDynamicHasStaticKey,
			"DatawellDynamicHasStaticKey",
			fmt.Sprintf("Datawell with source %q on edge %s has static key %q - all runs will overwrite the same object", *well.Source, well.Edge, *well.Key),
		})
	}
}

func wellStaticKeyIsDynamic(r *ValidationResult, well models.DataWell) {
	if well.Key == nil || well.Source == nil || *well.Source != "static" {
		return
	}

	os.Expand(*well.Key, func(s string) string {
		r.AddError(ValidationEntry{
			ValidationError,
			ErrDatawellStaticHasDynamicKey,
			"DatawellStaticHasDynamicKey",
			fmt.Sprintf("Datawell with source \"static\" on edge %s uses key variable %s", well.Edge, s),
		})
		return ""
	})
}

func wellStaticMissingKey(r *ValidationResult, well models.DataWell) {
	if well.Source == nil || *well.Source != "static" {
		return
	}

	if well.Key == nil {
		r.AddError(ValidationEntry{
			ValidationError,
			ErrDatawellStaticMissingKey,
			"DatawellStaticMissingKey",
			fmt.Sprintf("Datawell on edge %s set to static, but no key is provided", well.Edge),
		})
	}

}
