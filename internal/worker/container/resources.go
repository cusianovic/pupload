package container

type ResourceManager struct {
	MaxCPU     float64
	MaxMemory  string
	MaxStorage string
	MaxTimeout string

	GPU *GPUResources
}

type GPUResources struct {
	Vendor      string
	Features    []string
	MaxCount    int
	MaxMemoryGB float64
}
