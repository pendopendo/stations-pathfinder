package main

import (
	"fmt"
	"os"
	"strconv"
)

// Error handling
func handleError(msg string) {
	fmt.Fprintln(os.Stderr, "Error:", msg)
	os.Exit(1)
}

// Check if an element exists in a slice
func contains(slice []string, element string) bool {
	for _, e := range slice {
		if e == element {
			return true
		}
	}
	return false
}

func main() {
	if len(os.Args) != 2 && len(os.Args) != 5 {
		handleError("Incorrect number of command line arguments")
	}

	var mapFile, startStation, endStation string
	var numTrains int
	var err error

	// Predefined test cases
	tests := map[string][]string{
		"test1": {"waterloo", "st_pancras", "2"},
		"test2": {"bond_square", "space_port", "4"},
		"test3": {"beethoven", "part", "9"},
		"test4": {"beginning", "terminus", "20"},
		"test5": {"two", "four", "4"},
		"test6": {"jungle", "desert", "10"},
		"test7": {"small", "large", "9"},
	}

	if len(os.Args) == 2 {
		testName := os.Args[1]
		if testName == "test0" {
			fmt.Println("Running all tests")
			for name, args := range tests {
				fmt.Printf("\nRunning %s: %s %s %s\n", name, args[0], args[1], args[2])
				numTrains, err = strconv.Atoi(args[2])
				if err != nil || numTrains <= 0 {
					handleError("Number of trains is not a valid positive integer")
				}
				network, err := parseNetworkMap("network.map")
				if err != nil {
					handleError(err.Error())
				}
				simulateTrains(network, args[0], args[1], numTrains)
			}
			return
		} else if args, exists := tests[testName]; exists {
			fmt.Printf("Running %s: %s %s %s\n", testName, args[0], args[1], args[2])
			mapFile = "network.map"
			startStation = args[0]
			endStation = args[1]
			numTrains, err = strconv.Atoi(args[2])
			if err != nil || numTrains <= 0 {
				handleError("Number of trains is not a valid positive integer")
			}
		} else if testName == "10000" {
			fmt.Println("Running test for large map")
			mapFile = "10000.map"
			startStation = "000"
			endStation = "001"
			numTrains = 2
		} else {
			handleError("Unknown test name")
		}
	} else {
		mapFile = os.Args[1]
		startStation = os.Args[2]
		endStation = os.Args[3]
		numTrains, err = strconv.Atoi(os.Args[4])
		if err != nil || numTrains <= 0 {
			handleError("Number of trains is not a valid positive integer")
		}
	}

	network, err := parseNetworkMap(mapFile)
	if err != nil {
		handleError(err.Error())
	}

	if _, exists := network.Stations[startStation]; !exists {
		handleError("Start station does not exist: " + startStation)
	}

	if _, exists := network.Stations[endStation]; !exists {
		handleError("End station does not exist: " + endStation)
	}

	if startStation == endStation {
		handleError("Start station: '" + startStation + "' and end station: '" + endStation + "' are the same")
	}

	if !pathExists(startStation, endStation, network) {
		handleError("No path exists between the start station: '" + startStation + "' and end station: '" + endStation + "'")
	}

	// Simulate trains on the dynamic path
	simulateTrains(network, startStation, endStation, numTrains)
}
