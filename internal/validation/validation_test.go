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
		Steps: []models.Step{
			{
				ID:      "testnode",
				Uses:    "pupload/test",
				Inputs:  []models.StepEdge{{Name: "node_in", Edge: "in"}},
				Outputs: []models.StepEdge{{Name: "node_out", Edge: "out"}},
				Flags:   []models.StepFlag{{Name: "node_flag", Value: "10"}},
			},
		},
	}

	tasks := []models.Task{{
		Publisher:   "pupload",
		Name:        "test",
		Image:       "",
		MaxAttempts: 3,
		Tier:        "c-small",
		Inputs: []models.TaskEdgeDef{{
			Name:     "node_in",
			Required: true,
			Type:     []models.MimeType{"image/*"},
		}},
		Outputs: []models.TaskEdgeDef{{
			Name:     "node_out",
			Required: true,
			Type:     []models.MimeType{"image/*"},
		}},

		Flags: []models.TaskFlagDef{{
			Name:     "node_flag",
			Required: true,
			Type:     "string",
		}},

		Command: models.TaskCommandDef{
			Name: "testCommand",
			Exec: "exit 1",
		},
	}}

	res := Validate(flow, tasks)
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
		Steps: []models.Step{
			{
				ID:      "testnode1",
				Uses:    "pupload/test",
				Inputs:  []models.StepEdge{{Name: "node_in", Edge: "in-1"}},
				Outputs: []models.StepEdge{{Name: "node_out", Edge: "edge1-2"}},
				Flags:   []models.StepFlag{{Name: "node_flag", Value: "10"}},
			},
			{
				ID:      "testnode2",
				Uses:    "pupload/test",
				Inputs:  []models.StepEdge{{Name: "node_in", Edge: "edge1-2"}},
				Outputs: []models.StepEdge{{Name: "node_out", Edge: "out-2"}},
				Flags:   []models.StepFlag{{Name: "node_flag", Value: "10"}},
			},
		},
	}

	tasks := []models.Task{{
		Publisher:   "pupload",
		Name:        "test",
		Image:       "",
		MaxAttempts: 3,
		Tier:        "c-small",
		Inputs: []models.TaskEdgeDef{{
			Name:     "node_in",
			Required: true,
			Type:     []models.MimeType{"image/*"},
		}},
		Outputs: []models.TaskEdgeDef{{
			Name:     "node_out",
			Required: true,
			Type:     []models.MimeType{"image/*"},
		}},

		Flags: []models.TaskFlagDef{{
			Name:     "node_flag",
			Required: true,
			Type:     "string",
		}},

		Command: models.TaskCommandDef{
			Name: "testCommand",
			Exec: "exit 1",
		},
	}}

	res := Validate(flow, tasks)
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
		Steps: []models.Step{
			{
				ID:      "testnode1",
				Uses:    "pupload/test",
				Inputs:  []models.StepEdge{{Name: "node_in1", Edge: "in-1"}, {Name: "node_in1", Edge: "recursive-edge"}},
				Outputs: []models.StepEdge{{Name: "node_out", Edge: "edge1-2"}},
				Flags:   []models.StepFlag{{Name: "node_flag", Value: "10"}},
			},
			{
				ID:      "testnode2",
				Uses:    "pupload/test",
				Inputs:  []models.StepEdge{{Name: "node_in1", Edge: "edge1-2"}},
				Outputs: []models.StepEdge{{Name: "node_out", Edge: "recursive-edge"}},
				Flags:   []models.StepFlag{{Name: "node_flag", Value: "10"}},
			},
		},
	}

	tasks := []models.Task{{
		Publisher:   "pupload",
		Name:        "test",
		Image:       "",
		MaxAttempts: 3,
		Tier:        "c-small",
		Inputs: []models.TaskEdgeDef{
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
		Outputs: []models.TaskEdgeDef{{
			Name:     "node_out",
			Required: true,
			Type:     []models.MimeType{"image/*"},
		}},

		Flags: []models.TaskFlagDef{{
			Name:     "node_flag",
			Required: true,
			Type:     "string",
		}},

		Command: models.TaskCommandDef{
			Name: "testCommand",
			Exec: "exit 1",
		},
	}}

	res := Validate(flow, tasks)
	if res.HasError() {
		if len(res.Errors) != 1 {
			t.Errorf("expected flow to have only one errors: %v", *res)
		}

		if res.Errors[0].Code != ErrFlowCycle {
			t.Errorf("expected error to be ErrFlowCycle: %v", *res)
		}
	}

}
