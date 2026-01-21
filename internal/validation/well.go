package validation

import (
	"fmt"
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
