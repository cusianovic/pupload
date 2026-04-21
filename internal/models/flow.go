package models

type Flow struct {
	Name string

	Timeout *string

	Stores          []StoreInput
	DefaultDataWell *DataWell
	DataWells       []DataWell
	Steps           []Step
}

func (f *Flow) Normalize() {
	for _, s := range f.Stores {
		s.Normalize()
	}

}
