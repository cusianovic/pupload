package models

type Node struct {
	ID      string
	Uses    string
	Inputs  []NodeEdge
	Outputs []NodeEdge
	Flags   []NodeFlag
	Command string
}

type NodeEdge struct {
	Name string
	Edge string
}

type NodeFlag struct {
	Name  string
	Value string
}
