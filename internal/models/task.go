package models

const (
	DefaultTier     = "c-small"
	DefaultAttempts = 3
)

type Task struct {
	Publisher   string
	Name        string
	Image       string
	Inputs      []TaskEdgeDef
	Outputs     []TaskEdgeDef
	Flags       []TaskFlagDef
	Command     TaskCommandDef
	Tier        string
	MaxAttempts int
}

type TaskFlagDef struct {
	Name        string
	Description string
	Required    bool
	Type        string
}

type TaskEdgeDef struct {
	Name        string
	Description string
	Required    bool
	Type        []MimeType
}

type TaskCommandDef struct {
	Name        string
	Description string
	Exec        string
}

func (t *Task) Normalize() {
	if t.Tier == "" {
		t.Tier = DefaultTier
	}

	if t.MaxAttempts <= 0 {
		t.MaxAttempts = DefaultAttempts
	}
}
