package validation

import (
	"testing"

	"github.com/pupload/pupload/internal/models"
)

func ptr(s string) *string {
	return &s
}

func TestValidation_NormalFlow(t *testing.T) {
	flow := models.Flow{
		Name:   "testflow",
		Stores: []models.StoreInput{{Name: "teststore", Type: "s3"}},
		DataWells: []models.DataWell{
			{Edge: "in", Store: "teststore", Source: ptr("upload")},
			{Edge: "out", Store: "teststore"},
		},
		Nodes: []models.Node{
			{
				ID:      "testnode",
				Uses:    "pupload/test",
				Inputs:  []models.NodeEdge{{Name: "node_in", Edge: "in"}},
				Outputs: []models.NodeEdge{{Name: "node_out", Edge: "out"}},
				Flags:   []models.NodeFlag{{Name: "node_flag", Value: "10"}},
			},
		},
	}

	defs := []models.NodeDef{{
		ID:          0,
		Publisher:   "pupload",
		Name:        "test",
		Image:       "",
		MaxAttempts: 3,
		Tier:        "c-small",
		Inputs: []models.NodeEdgeDef{{
			Name:     "node_in",
			Required: true,
			Type:     []models.MimeType{"image/*"},
		}},
		Outputs: []models.NodeEdgeDef{{
			Name:     "node_out",
			Required: true,
			Type:     []models.MimeType{"image/*"},
		}},

		Flags: []models.NodeFlagDef{{
			Name:     "node_flag",
			Required: true,
			Type:     "string",
		}},

		Command: models.NodeCommandDef{
			Name: "testCommand",
			Exec: "exit 1",
		},
	}}

	res := Validate(flow, defs)
	if res.HasError() {
		t.Errorf("expected flow to have no errors: %v", *res)
	}
}

func TestValidation_MultiNodeFlow(t *testing.T) {
	flow := models.Flow{
		Name:   "testflow",
		Stores: []models.StoreInput{{Name: "teststore", Type: "s3"}},
		DataWells: []models.DataWell{
			{Edge: "in-1", Store: "teststore", Source: ptr("upload")},
			{Edge: "edge1-2", Store: "teststore"},
			{Edge: "out-2", Store: "teststore"},
		},
		Nodes: []models.Node{
			{
				ID:      "testnode1",
				Uses:    "pupload/test",
				Inputs:  []models.NodeEdge{{Name: "node_in", Edge: "in-1"}},
				Outputs: []models.NodeEdge{{Name: "node_out", Edge: "edge1-2"}},
				Flags:   []models.NodeFlag{{Name: "node_flag", Value: "10"}},
			},
			{
				ID:      "testnode2",
				Uses:    "pupload/test",
				Inputs:  []models.NodeEdge{{Name: "node_in", Edge: "edge1-2"}},
				Outputs: []models.NodeEdge{{Name: "node_out", Edge: "out-2"}},
				Flags:   []models.NodeFlag{{Name: "node_flag", Value: "10"}},
			},
		},
	}

	defs := []models.NodeDef{{
		ID:          0,
		Publisher:   "pupload",
		Name:        "test",
		Image:       "",
		MaxAttempts: 3,
		Tier:        "c-small",
		Inputs: []models.NodeEdgeDef{{
			Name:     "node_in",
			Required: true,
			Type:     []models.MimeType{"image/*"},
		}},
		Outputs: []models.NodeEdgeDef{{
			Name:     "node_out",
			Required: true,
			Type:     []models.MimeType{"image/*"},
		}},

		Flags: []models.NodeFlagDef{{
			Name:     "node_flag",
			Required: true,
			Type:     "string",
		}},

		Command: models.NodeCommandDef{
			Name: "testCommand",
			Exec: "exit 1",
		},
	}}

	res := Validate(flow, defs)
	if res.HasError() {
		t.Errorf("expected flow to have no errors: %v", *res)
	}
}

func TestValidation_CycleFlow(t *testing.T) {
	flow := models.Flow{
		Name:   "testflow",
		Stores: []models.StoreInput{{Name: "teststore", Type: "s3"}},
		DataWells: []models.DataWell{
			{Edge: "in-1", Store: "teststore", Source: ptr("upload")},
			{Edge: "edge1-2", Store: "teststore"},
			{Edge: "recursive-edge", Store: "teststore"},
		},
		Nodes: []models.Node{
			{
				ID:      "testnode1",
				Uses:    "pupload/test",
				Inputs:  []models.NodeEdge{{Name: "node_in1", Edge: "in-1"}, {Name: "node_in1", Edge: "recursive-edge"}},
				Outputs: []models.NodeEdge{{Name: "node_out", Edge: "edge1-2"}},
				Flags:   []models.NodeFlag{{Name: "node_flag", Value: "10"}},
			},
			{
				ID:      "testnode2",
				Uses:    "pupload/test",
				Inputs:  []models.NodeEdge{{Name: "node_in1", Edge: "edge1-2"}},
				Outputs: []models.NodeEdge{{Name: "node_out", Edge: "recursive-edge"}},
				Flags:   []models.NodeFlag{{Name: "node_flag", Value: "10"}},
			},
		},
	}

	defs := []models.NodeDef{{
		ID:          0,
		Publisher:   "pupload",
		Name:        "test",
		Image:       "",
		MaxAttempts: 3,
		Tier:        "c-small",
		Inputs: []models.NodeEdgeDef{
			{
				Name:     "node_in1",
				Required: true,
				Type:     []models.MimeType{"image/*"},
			},
			{
				Name:     "node_in2",
				Required: false,
				Type:     []models.MimeType{"image/*"},
			},
		},
		Outputs: []models.NodeEdgeDef{{
			Name:     "node_out",
			Required: true,
			Type:     []models.MimeType{"image/*"},
		}},

		Flags: []models.NodeFlagDef{{
			Name:     "node_flag",
			Required: true,
			Type:     "string",
		}},

		Command: models.NodeCommandDef{
			Name: "testCommand",
			Exec: "exit 1",
		},
	}}

	res := Validate(flow, defs)
	if res.HasError() {
		t.Errorf("expected flow to have no errors: %v", *res)
	}

}
