package models

type Step struct {
	ID      string
	Uses    string
	Inputs  []StepEdge
	Outputs []StepEdge
	Flags   []StepFlag
	Command string
}

type StepEdge struct {
	Name string
	Edge string
}

type StepFlag struct {
	Name  string
	Value string
}
