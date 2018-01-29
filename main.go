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
	// all information for converting all goal nodes
	// of the provenance graph from JSON to CSV.
	goals := make([][]string, 1, len(provCont.Goals))
	goals[0] = []string{"id", "label", "table"}
	for i := range provCont.Goals {
		goals = append(goals, []string{provCont.Goals[i].ID, provCont.Goals[i].Label, provCont.Goals[i].Table})
	}

	// Create slice of string slices containing
	// all information for converting all rule nodes
	// of the provenance graph from JSON to CSV.
	rules := make([][]string, 1, len(provCont.Rules))
	rules[0] = []string{"id", "label", "table"}
	for i := range provCont.Rules {
		rules = append(rules, []string{provCont.Rules[i].ID, provCont.Rules[i].Label, provCont.Rules[i].Table})
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

	goalsFile, err := os.OpenFile(filepath.Join(outDir, "goals.csv"), (os.O_CREATE | os.O_TRUNC | os.O_WRONLY), 0644)
	if err != nil {
		fmt.Printf("Could not open 'goals.csv' file in output directory: %v\n", err)
		os.Exit(1)
	}
	defer goalsFile.Close()

	rulesFile, err := os.OpenFile(filepath.Join(outDir, "rules.csv"), (os.O_CREATE | os.O_TRUNC | os.O_WRONLY), 0644)
	if err != nil {
		fmt.Printf("Could not open 'rules.csv' file in output directory: %v\n", err)
		os.Exit(1)
	}
	defer rulesFile.Close()

	edgesFile, err := os.OpenFile(filepath.Join(outDir, "edges.csv"), (os.O_CREATE | os.O_TRUNC | os.O_WRONLY), 0644)
	if err != nil {
		fmt.Printf("Could not open 'edges.csv' file in output directory: %v\n", err)
		os.Exit(1)
	}
	defer edgesFile.Close()

	goalsWriter := csv.NewWriter(goalsFile)
	goalsWriter.WriteAll(goals)

	err = goalsWriter.Error()
	if err != nil {
		fmt.Printf("Error while writing back CSV data for goal nodes: %v\n", err)
		os.Exit(1)
	}

	rulesWriter := csv.NewWriter(rulesFile)
	rulesWriter.WriteAll(rules)

	err = rulesWriter.Error()
	if err != nil {
		fmt.Printf("Error while writing back CSV data for rule nodes: %v\n", err)
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
