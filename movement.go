package main

import (
	"fmt"
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

		fmt.Printf("Turn %d:\n", turn)

		// Iterate over each train to determine its movement
		for i, train := range trains {
			// Skip trains that have already reached the end station
			if train.Current == endStation {
				continue
			}

			// Assign path if not already assigned and the train is not at the start station
			if train.AssignedPath == nil || len(train.AssignedPath) == 0 && train.Current != startStation {
				train.AssignedPath = dynamicDFS(train.Name, train.Current, endStation, network, occupiedStations, usedSegments, trains, visitedHistories[i])
				if train.AssignedPath == nil {
					allTrainsAtDestination = false
					continue
				}
			}

			// Determine the next station and segment for the train
			if len(train.AssignedPath) > 1 {
				nextStation := train.AssignedPath[1]
				segment := fmt.Sprintf("%s-%s", train.Current, nextStation)

				// Ensure the next station and the segment are available
				if !occupiedStations[nextStation] && !usedSegments[segment] {
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

					// Remove the first station from the assigned path
					train.AssignedPath = train.AssignedPath[1:]

					// Update delay for subsequent trains
					for j := i + 1; j < numTrains; j++ {
						trainDelays[j]++
					}
				} else {
					allTrainsAtDestination = false
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

		fmt.Printf("%s\n", strings.Join(movement, " "))

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
