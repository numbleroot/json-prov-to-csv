package main

import (
	"flag"
	"fmt"
	"os"

	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"path/filepath"
)

func main() {

	mollyOutFlag := flag.String("mollyOut", "", "Specify path to output directory of molly execution to convert to CSV.")
	csvOutDirFlag := flag.String("csvOutDir", "csv", "Specify output directory that will hold subfolders containing the generated CSV files.")
	csvFilesUserFlag := flag.String("csvUser", "", "If final CSV files are owned by special user, define it here.")
	csvFilesGroupFlag := flag.String("csvGroup", "", "If final CSV files are owned by special group, define it here.")
	flag.Parse()

	mollyOutRaw := *mollyOutFlag
	if mollyOutRaw == "" {
		fmt.Println("Please supply the path to the molly output directory you want to convert, \"-mollyOut <PATH>\"")
		os.Exit(1)
	}

	mollyOut, err := filepath.Abs(mollyOutRaw)
	if err != nil {
		fmt.Printf("Failed to call Abs() on supplied molly output directory: %v\n", err)
		os.Exit(1)
	}
	mollyRun := filepath.Base(mollyOut)

	csvOutDir, err := filepath.Abs(*csvOutDirFlag)
	if err != nil {
		fmt.Printf("Failed to call Abs() on supplied output directory: %v\n", err)
		os.Exit(1)
	}
	csvOutDir = filepath.Join(csvOutDir, mollyRun)

	csvFilesUser := *csvFilesUserFlag
	csvFilesGroup := *csvFilesGroupFlag

	if ((csvFilesUser != "") && (csvFilesGroup == "")) || ((csvFilesUser == "") && (csvFilesGroup != "")) {
		fmt.Printf("Please either provide both, '-csvUser' and '-csvGroup', or none of them.\n", err)
		os.Exit(1)
	}

	// Find all run_X_provenance.json files in the
	// supplied output directory of a molly execution.
	runFiles, err := filepath.Glob(filepath.Join(mollyOut, "run_*_provenance.json"))
	if err != nil {
		fmt.Printf("Glob() not possible on molly output directory: %v\n", err)
		os.Exit(1)
	}

	tmpOutDir, err := ioutil.TempDir("", "JSONtoCSV")
	if err != nil {
		fmt.Printf("Failed to create temporary directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpOutDir)

	for i := range runFiles {

		var provCont ProvFile

		// Read in JSON provenance graph.
		rawProvCont, err := ioutil.ReadFile(runFiles[i])
		if err != nil {
			fmt.Printf("Error reading '%s': %v\n", runFiles[i], err)
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

		goalsFile, err := os.OpenFile(filepath.Join(tmpOutDir, fmt.Sprintf("%d_goals.csv", i)), (os.O_CREATE | os.O_TRUNC | os.O_WRONLY), 0644)
		if err != nil {
			fmt.Printf("Could not open 'goals.csv' file in output directory: %v\n", err)
			os.Exit(1)
		}
		defer goalsFile.Close()

		rulesFile, err := os.OpenFile(filepath.Join(tmpOutDir, fmt.Sprintf("%d_rules.csv", i)), (os.O_CREATE | os.O_TRUNC | os.O_WRONLY), 0644)
		if err != nil {
			fmt.Printf("Could not open 'rules.csv' file in output directory: %v\n", err)
			os.Exit(1)
		}
		defer rulesFile.Close()

		edgesFile, err := os.OpenFile(filepath.Join(tmpOutDir, fmt.Sprintf("%d_edges.csv", i)), (os.O_CREATE | os.O_TRUNC | os.O_WRONLY), 0644)
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

	// Move created temporary output directory
	// to location specified by command-line flag.
	mvCmd := exec.Command("sudo", "mv", fmt.Sprintf("%s/", tmpOutDir), csvOutDir)
	mvOut, err := mvCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to 'mv' temporary to final output directory: %s, %v\n", mvOut, err)
		os.Exit(1)
	}

	if string(mvOut) != "" {
		fmt.Printf("Unexpected output ('mv'): %s\n", mvOut)
		os.Exit(1)
	}

	if (csvFilesUser != "") && (csvFilesGroup != "") {

		chownCmd := exec.Command("sudo", "chown", "-R", fmt.Sprintf("%s:%s", csvFilesUser, csvFilesGroup), fmt.Sprintf("%s/", csvOutDir))
		chownOut, err := chownCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Failed to 'chown' the final output CSV files: %s, %v\n", chownOut, err)
			os.Exit(1)
		}

		if string(chownOut) != "" {
			fmt.Printf("Unexpected output ('chown'): %s\n", chownOut)
			os.Exit(1)
		}
	}
}
