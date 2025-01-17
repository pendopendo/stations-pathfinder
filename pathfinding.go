package main

import (
	"fmt"
	"reflect"
)

// errorfunktsioonid wrappituna annavad parema erorrite jada
// helperfunktsioonid et kergem lugeda oleks
func dynamicDFS(trainName, startStation, endStation string, network *Network, occupiedStations, usedSegments map[string]bool, trains []*Train, visitedHistory map[string]bool) []string {
	// Initialize the stack with the start station
	stack := [][]string{{startStation}}
	// Slice to store all possible paths
	var allPaths [][]string
	// Initialize variables for the active path
	var activePath []string

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

	// If no paths were found, return nil
	if len(allPaths) == 0 {
		return nil
	}

	// 1. Initialize variable to store combinations of non-crossing paths
	var bestPathCombination [][]string
	var allPathCombinations [][][]string

	// 2. Helper function to check if two paths cross each other (share any segments)
	pathsConflict := func(path1, path2 []string) bool {
		// If the paths are too short to have any intermediate stations, return false
		if len(path1) <= 2 || len(path2) <= 2 {
			return false // No intermediate stations to check
		}

		// Create a set (map) for path1's intermediate stations (ignore first and last station)
		stationSet := make(map[string]bool)
		for i := 1; i < len(path1)-1; i++ { // Skip the first and last station
			stationSet[path1[i]] = true
		}

		// Check if any intermediate station from path2 exists in path1's set
		for i := 1; i < len(path2)-1; i++ { // Skip the first and last station
			if stationSet[path2[i]] {
				return true // Conflict found: shared intermediate station
			}
		}

		return false // No conflict found
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
	shortestPath := bestPathCombination[0] // Default to the first found path
	var alternativePath []string

	if len(bestPathCombination) > 1 {
		// If there's more than one path, find the shortest and an alternative
		shortestPath = bestPathCombination[0]
		alternativePath = bestPathCombination[1]

		for _, path := range bestPathCombination {
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

	// Check if the current alternativePath has occupied stations or segments
	alternativePathAvailable := true
	if len(alternativePath) > 1 {
		for i := 0; i < len(alternativePath)-1; i++ {
			segment := fmt.Sprintf("%s-%s", alternativePath[i], alternativePath[i+1])
			if occupiedStations[alternativePath[i+1]] || usedSegments[segment] {
				alternativePathAvailable = false
				break
			}
		}
	}

	// If the current alternativePath is not available, find a new alternative path
	if !alternativePathAvailable {
		for _, path := range bestPathCombination {
			if len(path) > 1 && !reflect.DeepEqual(path, shortestPath) {
				segment := fmt.Sprintf("%s-%s", path[0], path[1])
				if !occupiedStations[path[1]] && !usedSegments[segment] {
					alternativePath = path
					break
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
	if activePath == nil || (len(activePath) > 1 && occupiedStations[activePath[1]] && usedSegments[fmt.Sprintf("%s-%s", activePath[0], activePath[1])]) {
		return nil
	}

	// Return the selected path
	return activePath
}
