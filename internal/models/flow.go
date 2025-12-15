package models

type Flow struct {
	Name string

	Stores []StoreInput

	DefaultStore *string

	DefaultDataWell *DataWell

	DataWells      []DataWell
	AvailableNodes []string
	Nodes          []Node
}

func (f *Flow) Normalize() {
	for _, s := range f.Stores {
		s.Normalize()
	}

}
