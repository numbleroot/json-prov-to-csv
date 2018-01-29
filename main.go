package main

import (
	"flag"
	"fmt"
	"os"

	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

func main() {

	var provCont ProvFile

	provFileFlag := flag.String("prov", "", "Specify name of JSON file containing provenance graph to convert.")
	outDirFlag := flag.String("out", "csv", "Specify output directory that will hold the generated CSV files.")
	flag.Parse()

	provFile := *provFileFlag
	if provFile == "" {
		fmt.Println("Please supply a provenance JSON file with \"-prov <FILE>.json\"\n")
		os.Exit(1)
	}

	outDir, err := filepath.Abs(*outDirFlag)
	if err != nil {
		fmt.Printf("Failed to call Abs() on supplied output directory: %v\n", err)
		os.Exit(1)
	}

	// Read in JSON provenance graph.
	rawProvCont, err := ioutil.ReadFile(provFile)
	if err != nil {
		fmt.Printf("Error reading '%s': %v\n", provFile, err)
		os.Exit(1)
	}

	// Unmarshal (decode) JSON into defined
	// provenance file structure.
	err = json.Unmarshal(rawProvCont, &provCont)
	if err != nil {
		fmt.Printf("Failed to unmarshal JSON content to provenance file structure: %v\n", err)
		os.Exit(1)
	}

	// Create slice of string slices containing
	// all information for converting all nodes
	// of the provenance graph from JSON to CSV.
	nodes := make([][]string, 1, len(provCont.Nodes))
	nodes[0] = []string{"id", "label", "table"}
	for i := range provCont.Nodes {
		nodes = append(nodes, []string{provCont.Nodes[i].ID, provCont.Nodes[i].Label, provCont.Nodes[i].Table})
	}

	// Create slice of string slices containing
	// all information for converting all edges
	// of the provenance graph from JSON to CSV.
	edges := make([][]string, 1, len(provCont.Edges))
	edges[0] = []string{"from", "to"}
	for i := range provCont.Edges {
		edges = append(edges, []string{provCont.Edges[i].From, provCont.Edges[i].To})
	}

	// If output directory does not exist,
	// create it.
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		fmt.Printf("Could not create specified directory '%s': %v\n", outDir, err)
		os.Exit(1)
	}

	nodesFile, err := os.OpenFile(filepath.Join(outDir, "nodes.csv"), (os.O_CREATE | os.O_TRUNC | os.O_WRONLY), 0644)
	if err != nil {
		fmt.Printf("Could not open 'nodes.csv' file in output directory: %v\n", err)
		os.Exit(1)
	}
	defer nodesFile.Close()

	edgesFile, err := os.OpenFile(filepath.Join(outDir, "edges.csv"), (os.O_CREATE | os.O_TRUNC | os.O_WRONLY), 0644)
	if err != nil {
		fmt.Printf("Could not open 'edges.csv' file in output directory: %v\n", err)
		os.Exit(1)
	}

	nodesWriter := csv.NewWriter(nodesFile)
	nodesWriter.WriteAll(nodes)

	err = nodesWriter.Error()
	if err != nil {
		fmt.Printf("Error while writing back CSV data for nodes: %v\n", err)
		os.Exit(1)
	}

	edgesWriter := csv.NewWriter(edgesFile)
	edgesWriter.WriteAll(edges)

	err = edgesWriter.Error()
	if err != nil {
		fmt.Printf("Error while writing back CSV data for edges: %v\n", err)
		os.Exit(1)
	}
}
