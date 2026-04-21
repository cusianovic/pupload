package validation

import (
	"fmt"

	"github.com/pupload/pupload/internal/models"
	"github.com/pupload/pupload/internal/resources"
)

func getTaskDef(step models.Step, tasks []models.Task) *models.Task {
	var def *models.Task

	for _, d := range tasks {
		if fmt.Sprintf("%s/%s", d.Publisher, d.Name) == step.Uses {
			def = &d
		}
	}

	return def
}

func stepNoTaskFound(r *ValidationResult, step models.Step, tasks []models.Task) {
	def := getTaskDef(step, tasks)
	if def == nil {
		r.AddError(ValidationEntry{
			ValidationError,
			ErrTaskNotFound,
			"TaskNotFound",
			fmt.Sprintf("Definition %s not found (step %s)", step.Uses, step.ID),
		})
	}
}

func stepMissingInput(r *ValidationResult, step models.Step, tasks []models.Task) {
	def := getTaskDef(step, tasks)
	if def == nil {
		return
	}

	for _, inDef := range def.Inputs {
		found := false
		for _, inStep := range step.Inputs {
			if !inDef.Required || inStep.Name == inDef.Name {
				found = true
				break
			}
		}

		if found {
			continue
		}

		r.AddError(ValidationEntry{
			ValidationError,
			ErrStepMissingInput,
			"StepMissingInput",
			fmt.Sprintf("Step %s missing required input %s", step.ID, inDef.Name),
		})

	}
}

// should be a runtime error, the required flag should mean that the worker doesn't generate an output
func stepMissingOutput(r *ValidationResult, step models.Step, tasks []models.Task) {
	def := getTaskDef(step, tasks)
	if def == nil {
		return
	}

	for _, outDef := range def.Outputs {
		found := false
		for _, outStep := range step.Outputs {
			if !outDef.Required || outStep.Name == outDef.Name {
				found = true
				break
			}
		}

		if found {
			continue
		}

		r.AddError(ValidationEntry{
			ValidationError,
			ErrStepMissingInput,
			"StepMissingOutput",
			fmt.Sprintf("Step %s missing required output %s", step.ID, outDef.Name),
		})
	}
}

func stepInvalidTier(r *ValidationResult, step models.Step, tasks []models.Task) {
	def := getTaskDef(step, tasks)
	if def == nil {
		return
	}

	_, tierValid := resources.StandardTierMap[def.Tier]
	if !tierValid {
		r.AddError(ValidationEntry{
			ValidationError,
			ErrStepMissingInput,
			"StepInvalidTier",
			fmt.Sprintf("Step %s uses invalid tier %s", step.ID, def.Tier),
		})
	}
}

func stepMissingFlag(r *ValidationResult, step models.Step, tasks []models.Task) {
	def := getTaskDef(step, tasks)
	if def == nil {
		return
	}

	for _, flagDef := range def.Flags {
		found := false
		for _, flagStep := range step.Flags {
			if !flagDef.Required || flagStep.Name == flagDef.Name {
				found = true
				break
			}
		}

		if found {
			continue
		}

		r.AddError(ValidationEntry{
			ValidationError,
			ErrStepMissingFlag,
			"StepMissingFlag",
			fmt.Sprintf("Step %s missing required flag %s", step.ID, flagDef.Name),
		})
	}
}

func stepUnknownFlag(r *ValidationResult, step models.Step, tasks []models.Task) {
	def := getTaskDef(step, tasks)
	if def == nil {
		return
	}

	for _, flagStep := range step.Flags {
		found := false
		for _, flagDef := range def.Flags {
			if flagStep.Name == flagDef.Name {
				found = true
				break
			}
		}

		if found {
			continue
		}

		r.AddError(ValidationEntry{
			ValidationError,
			ErrStepMissingFlag,
			"StepUnknownFlag",
			fmt.Sprintf("Step %s has unknown flag %s", step.ID, flagStep.Name),
		})
	}
}

func stepMissingID(r *ValidationResult, step models.Step) {
	if step.ID == "" {
		r.AddError(ValidationEntry{
			ValidationError,
			ErrStepMissingID,
			"StepMissingID",
			"Step missing ID",
		})
	}
}

func stepDuplicateID(r *ValidationResult, steps []models.Step) {
	idCount := make(map[string]int)
	for _, step := range steps {
		idCount[step.ID]++
	}

	for id, val := range idCount {
		if val > 1 {
			r.AddError(ValidationEntry{
				ValidationError,
				ErrStepDuplicateID,
				"StepDuplicateID",
				fmt.Sprintf("Step ID %s is used more than once", id),
			})
		}
	}

}
