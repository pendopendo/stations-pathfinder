package main

import (
	"fmt"
	"reflect"
	"strings"
)

func simulateTrains(network *Network, startStation, endStation string, numTrains int) {
	// Create a slice to hold the trains
	trains := make([]*Train, numTrains)
	// Create a slice to track delays for each train
	trainDelays := make([]int, numTrains)
	// Create a map to track visited history for each train
	visitedHistories := make([]map[string]bool, numTrains)

	// Initialize all trains at the start station
	for i := 0; i < numTrains; i++ {
		trains[i] = &Train{Name: fmt.Sprintf("T%d", i+1), Current: startStation}
		visitedHistories[i] = make(map[string]bool) // Properly initialize the map
		visitedHistories[i][startStation] = true
	}

	// Initialize turn counter and consecutive stuck turns counter
	turn := 1
	consecutiveStuckTurns := 0

	// Main simulation loop
	for {
		// Maps to track used segments and occupied stations
		usedSegments := make(map[string]bool)
		occupiedStations := make(map[string]bool)

		// Mark stations occupied by trains (except the end station)
		for _, train := range trains {
			if train.Current != endStation {
				occupiedStations[train.Current] = true
			}
		}

		// Ensure start station is not marked as occupied
		delete(occupiedStations, startStation)
		delete(occupiedStations, endStation)

		// Slice to track movements in the current turn
		movement := []string{}
		// Flag to check if all trains have reached their destinations
		allTrainsAtDestination := true

		fmt.Printf("\nTurn %d:\n", turn)

		// Iterate over each train to determine its movement
		for i, train := range trains {
			// Skip trains that have already reached the end station
			if train.Current == endStation {
				continue
			}

			// Find the path for the current train
			currentPath := dynamicDFS(train.Name, train.Current, endStation, network, occupiedStations, usedSegments, trains, visitedHistories[i])

			// If no valid path is found, mark that not all trains are at their destination
			if currentPath == nil || len(currentPath) < 2 {
				allTrainsAtDestination = false
				continue
			}

			// Determine the next station and segment for the train
			nextStation := currentPath[1]
			segment := fmt.Sprintf("%s-%s", train.Current, nextStation)

			// Ensure the next station and the segment are available
			if !occupiedStations[nextStation] && !usedSegments[segment] {
				fmt.Printf("Train %s moving from %s to %s\n", train.Name, train.Current, nextStation)
				previousStation := train.Current
				train.Current = nextStation
				movement = append(movement, fmt.Sprintf("%s-%s", train.Name, nextStation))

				// Update occupancy
				if previousStation != endStation {
					delete(occupiedStations, previousStation)
				}
				if nextStation != endStation {
					occupiedStations[nextStation] = true
				}
				usedSegments[segment] = true

				// Update visited history
				visitedHistories[i][nextStation] = true

				//fmt.Printf("Occupied Stations after move: %v\n", occupiedStations)
				//fmt.Printf("Used Segments after move: %v\n", usedSegments)

				// Update delay for subsequent trains
				for j := i + 1; j < numTrains; j++ {
					trainDelays[j]++
				}
			} else {
				allTrainsAtDestination = false
			}
		}

		// If no movements occurred, increment the consecutive stuck turns counter
		if len(movement) == 0 {
			consecutiveStuckTurns++
		} else {
			consecutiveStuckTurns = 0
		}

		fmt.Printf("Movements this turn: %s\n", strings.Join(movement, " "))

		// Check if all trains have reached their destinations
		allTrainsAtDestination = true
		for _, train := range trains {
			if train.Current != endStation {
				allTrainsAtDestination = false
				break
			}
		}

		// If all trains have reached their destinations, end the simulation
		if allTrainsAtDestination {
			fmt.Println("All trains have reached their destinations. Simulation ending.")
			break
		}

		// If no trains moved for 2 consecutive turns, end the simulation
		if consecutiveStuckTurns >= 2 {
			fmt.Println("Faulty simulation detected: No trains moved for 2 consecutive turns. Exiting simulation.")
			break
		}

		// Increment the turn counter
		turn++
	}
}

func dynamicDFS(trainName, startStation, endStation string, network *Network, occupiedStations, usedSegments map[string]bool, trains []*Train, visitedHistory map[string]bool) []string {
	// Initialize the stack with the start station
	stack := [][]string{{startStation}}
	// Slice to store all possible paths
	var allPaths [][]string
	// Initialize variables for the active path
	var activePath []string

	fmt.Printf("\nTrain %s: Starting DFS from %s to %s\n", trainName, startStation, endStation)

	// DFS loop
	for len(stack) > 0 {
		// Pop the last path from the stack
		currentPath := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		// Get the current station from the path
		currentStation := currentPath[len(currentPath)-1]

		// If the current station is the end station, add the path to allPaths
		if currentStation == endStation {
			allPaths = append(allPaths, currentPath)
			continue
		}

		// Iterate over all neighbors of the current station
		for _, neighbor := range network.Connections[currentStation] {
			// Avoid loops and backtracking
			if contains(currentPath, neighbor) || visitedHistory[neighbor] {
				continue
			}
			// Create a new path by copying the current path and adding the neighbor
			newPath := append([]string{}, currentPath...)
			newPath = append(newPath, neighbor)
			// Push the new path onto the stack
			stack = append(stack, newPath)
		}
	}
	//fmt.Printf("Train %s: All paths found: %v\n", trainName, allPaths)

	// If no paths were found, return nil
	if len(allPaths) == 0 {
		fmt.Printf("Train %s: No path found.\n", trainName)
		return nil
	}

	// Siia uus funktsioon et filtreerida leitud teid
	// 1. Initialize variable to store combinations of non-crossing paths
	var bestPathCombination [][]string
	var allPathCombinations [][][]string

	// 2. Helper function to check if two paths cross each other (share any segments)
	pathsConflict := func(path1, path2 []string) bool {
		for i := 0; i < len(path1)-1; i++ {
			for j := 0; j < len(path2)-1; j++ {
				if path1[i] == path2[j] && path1[i+1] == path2[j+1] {
					// If they share the same segment between two stations
					return true
				}
			}
		}
		return false
	}

	// 3. Function to calculate total length of a set of paths
	totalLength := func(paths [][]string) int {
		total := 0
		for _, path := range paths {
			total += len(path)
		}
		return total
	}

	// 4. Iterate through all combinations of paths to find non-crossing ones
	for i := 0; i < len(allPaths); i++ {
		currentCombination := [][]string{allPaths[i]}

		// For each path, try to combine with other non-conflicting paths
		for j := 0; j < len(allPaths); j++ {
			if i != j {
				conflict := false
				// Check if the current path combination has a conflict with the new path
				for _, existingPath := range currentCombination {
					if pathsConflict(existingPath, allPaths[j]) {
						conflict = true
						break
					}
				}
				// If no conflict, add this path to the current combination
				if !conflict {
					currentCombination = append(currentCombination, allPaths[j])
				}
			}
		}
		// Add the current non-conflicting combination to allPathCombinations
		allPathCombinations = append(allPathCombinations, currentCombination)

		// 5. Select the combination that has the most paths with the least total length
		if len(currentCombination) > len(bestPathCombination) ||
			(len(currentCombination) == len(bestPathCombination) && totalLength(currentCombination) < totalLength(bestPathCombination)) {
			bestPathCombination = currentCombination
		}
	}

	// 6. Select the path with the shortest length from the best combination
	activePath = bestPathCombination[0]
	for _, path := range bestPathCombination {
		// Check if the next station and the connection are not occupied
		if len(path) > 1 {
			segment := fmt.Sprintf("%s-%s", path[0], path[1])
			if !occupiedStations[path[1]] && !usedSegments[segment] {
				if len(path) < len(activePath) {
					activePath = path
				}
			}
		}
	}

	// Ensure there's always a valid path to choose
	shortestPath := allPaths[0] // Default to the first found path
	var alternativePath []string

	if len(allPaths) > 1 {
		// If there's more than one path, find the shortest and an alternative
		shortestPath = allPaths[0]
		alternativePath = allPaths[1]

		for _, path := range allPaths {
			if len(path) > 1 {
				segment := fmt.Sprintf("%s-%s", path[0], path[1])

				// Check if the next station and the segment are not occupied
				if !occupiedStations[path[1]] && !usedSegments[segment] {
					if len(path) < len(shortestPath) {
						alternativePath = shortestPath // Keep track of previous shortest as alternative
						shortestPath = path
					} else if len(alternativePath) == 0 || (len(path) < len(alternativePath) && !reflect.DeepEqual(path, shortestPath)) {
						alternativePath = path
					}
				}
			}
		}
	}

	// Get the train's index in the train list for threshold calculation
	numTrains := len(trains)
	currentTrain := 0
	for i, train := range trains {
		if train.Name == trainName {
			currentTrain = i + 1
			break
		}
	}

	// Check availability of the shortest path
	available := true
	if len(shortestPath) > 1 {
		segment := fmt.Sprintf("%s-%s", shortestPath[0], shortestPath[1])
		if occupiedStations[shortestPath[1]] || usedSegments[segment] {
			available = false
		}
	}

	// Decide on the path: If the shortest path is blocked, consider the alternative
	if !available && alternativePath != nil && len(alternativePath) > 1 {
		threshold := numTrains - currentTrain + len(shortestPath)
		if threshold >= len(alternativePath) {
			// Choose the alternative path if it's better than waiting
			activePath = alternativePath
		}
	} else if available {
		// Choose the shortest path if it's available
		activePath = shortestPath
	}

	// If no available path was found, return nil
	if activePath == nil {
		fmt.Printf("Train %s: No available path found.\n", trainName)
		return nil
	}

	// Print the selected path
	fmt.Printf("Train %s: Selected path: %v\n", trainName, activePath)

	// Return the selected path
	return activePath
}
