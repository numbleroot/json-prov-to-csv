package main

type ProvFile struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type Node struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Table string `json:"table"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}
