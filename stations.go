package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Station struct to store station data
type Station struct {
	Name string
	X    int
	Y    int
}

// Train struct to store train data
type Train struct {
	Name    string
	Current string
}

// Network struct to store the whole network graph
type Network struct {
	Stations    map[string]*Station
	Connections map[string][]string
	Paths       map[string]map[string][]string
}

// Error handling
func handleError(msg string) {
	fmt.Fprintln(os.Stderr, "Error:", msg)
	os.Exit(1)
}

// Read and parse the network map file
func parseNetworkMap(filePath string) (*Network, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	network := &Network{
		Stations:    make(map[string]*Station),
		Connections: make(map[string][]string),
		Paths:       make(map[string]map[string][]string),
	}

	scanner := bufio.NewScanner(file)
	stationSection := false
	connectionSection := false

	// Regex to allow flexible whitespace and comments
	stationRegex := regexp.MustCompile(`^\s*([a-zA-Z0-9_]+)\s*,\s*([0-9]+)\s*,\s*([0-9]+)\s*(?:#.*)?$`)
	connectionRegex := regexp.MustCompile(`^\s*([a-zA-Z0-9_]+)\s*-\s*([a-zA-Z0-9_]+)\s*(?:#.*)?$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Ignore blank lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if line == "stations:" {
			stationSection = true
			connectionSection = false
			continue
		}

		if line == "connections:" {
			stationSection = false
			connectionSection = true
			continue
		}

		if stationSection {
			match := stationRegex.FindStringSubmatch(line)
			if match == nil {
				return nil, errors.New("Invalid station format: " + line)
			}
			name, xStr, yStr := match[1], match[2], match[3]
			x, _ := strconv.Atoi(xStr)
			y, _ := strconv.Atoi(yStr)
			if _, exists := network.Stations[name]; exists {
				return nil, errors.New("Duplicate station name: " + name)
			}
			network.Stations[name] = &Station{Name: name, X: x, Y: y}
		} else if connectionSection {
			match := connectionRegex.FindStringSubmatch(line)
			if match == nil {
				return nil, errors.New("Invalid connection format: " + line)
			}
			station1, station2 := match[1], match[2]
			if station1 == station2 {
				return nil, errors.New("Connection between the same station: " + station1)
			}
			if _, exists := network.Stations[station1]; !exists {
				return nil, errors.New("Connection with non-existing station: " + station1)
			}
			if _, exists := network.Stations[station2]; !exists {
				return nil, errors.New("Connection with non-existing station: " + station2)
			}
			if contains(network.Connections[station1], station2) || contains(network.Connections[station2], station1) {
				return nil, errors.New("Duplicate connection between " + station1 + " and " + station2)
			}
			network.Connections[station1] = append(network.Connections[station1], station2)
			network.Connections[station2] = append(network.Connections[station2], station1)
		}
	}

	if len(network.Stations) > 10000 {
		return nil, errors.New("Map contains more than 10,000 stations")
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Precompute all possible paths between stations
	for start := range network.Stations {
		network.Paths[start] = make(map[string][]string)
		for end := range network.Stations {
			if start != end {
				network.Paths[start][end] = findAllPaths(network, start, end)
			}
		}
	}

	return network, nil
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

// Function to find all possible paths between two stations
func findAllPaths(network *Network, startStation, endStation string) []string {
	var paths []string
	queue := []string{startStation}
	predecessors := make(map[string]string)
	predecessors[startStation] = ""

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == endStation {
			// Reconstruct the path
			path := []string{}
			for at := endStation; at != ""; at = predecessors[at] {
				path = append([]string{at}, path...)
			}
			paths = append(paths, fmt.Sprintf("%v", path))
			continue
		}

		for _, neighbor := range network.Connections[current] {
			if _, visited := predecessors[neighbor]; !visited {
				queue = append(queue, neighbor)
				predecessors[neighbor] = current
			}
		}
	}

	return paths
}

// BFS with debugging and finding the shortest path
func dynamicBFS(network *Network, startStation, endStation string, occupiedStations, usedSegments map[string]bool) []string {
	if startStation == endStation {
		return []string{startStation}
	}

	queue := [][]string{{startStation}} // Queue of paths
	visited := make(map[string]bool)
	visited[startStation] = true

	fmt.Printf("Starting BFS from %s to %s\n", startStation, endStation)

	for len(queue) > 0 {
		currentPath := queue[0]
		queue = queue[1:]
		currentStation := currentPath[len(currentPath)-1]

		fmt.Printf("Current station: %s\n", currentStation)
		fmt.Printf("Current path: %v\n", currentPath)
		fmt.Printf("Queue: %v\n", queue)

		if currentStation == endStation {
			fmt.Printf("Path found: %v\n", currentPath)
			return currentPath
		}

		for _, neighbor := range network.Connections[currentStation] {
			segment := fmt.Sprintf("%s-%s", currentStation, neighbor)
			if !visited[neighbor] && !occupiedStations[neighbor] && !usedSegments[segment] {
				visited[neighbor] = true
				newPath := append([]string{}, currentPath...) // Copy current path
				newPath = append(newPath, neighbor)
				queue = append(queue, newPath)
				fmt.Printf("Adding neighbor %s to queue\n", neighbor)
			}
		}
	}

	fmt.Println("No path found.")
	return nil // No path found
}

func printBFSDebugInfo(queue [][]string, visited map[string]bool, predecessors map[string]string) {
	fmt.Println("Queue Debug Info:")
	for i, path := range queue {
		fmt.Printf("  %d: %v\n", i, path)
	}
	fmt.Println("Visited Debug Info:")
	for station := range visited {
		fmt.Printf("  %s\n", station)
	}
	fmt.Println("Predecessors Debug Info:")
	for station, pred := range predecessors {
		fmt.Printf("  %s: %s\n", station, pred)
	}
}

func simulateTrains(network *Network, startStation, endStation string, numTrains int) {
    trains := make([]*Train, numTrains)

    // Initialize all trains at the start station
    for i := 0; i < numTrains; i++ {
        trains[i] = &Train{Name: fmt.Sprintf("T%d", i+1), Current: startStation}
    }

    turn := 1
    consecutiveStuckTurns := 0

    for {
        usedSegments := make(map[string]bool)
        occupiedStations := make(map[string]bool)

        for _, train := range trains {
            if train.Current != endStation {
                occupiedStations[train.Current] = true
            }
        }

        if _, exists := occupiedStations[startStation]; exists {
            delete(occupiedStations, startStation)
        }
        if _, exists := occupiedStations[endStation]; exists {
            delete(occupiedStations, endStation)
        }

        movement := []string{}
        allTrainsAtDestination := true

        fmt.Printf("\nTurn %d:\n", turn)

        for _, train := range trains {
            if train.Current == endStation {
                continue
            }

            currentPath := dynamicBFS(network, train.Current, endStation, occupiedStations, usedSegments)

            if currentPath == nil || len(currentPath) < 2 {
                allTrainsAtDestination = false
                continue
            }

            nextStation := currentPath[1]
            segment := fmt.Sprintf("%s-%s", train.Current, nextStation)

            if !occupiedStations[nextStation] && !usedSegments[segment] {
                train.Current = nextStation
                movement = append(movement, fmt.Sprintf("%s-%s", train.Name, nextStation))

                occupiedStations[nextStation] = true
                usedSegments[segment] = true
            }
        }

        if len(movement) == 0 {
            consecutiveStuckTurns++
        } else {
            consecutiveStuckTurns = 0
        }

        fmt.Printf("Movements this turn: %s\n", strings.Join(movement, " "))

        if allTrainsAtDestination {
            fmt.Println("All trains have reached their destinations. Simulation ending.")
            break
        }

        if consecutiveStuckTurns >= 2 {
            fmt.Println("Faulty simulation detected: No trains moved for 2 consecutive turns. Exiting simulation.")
            break
        }

        turn++
    }
}



func main() {
    if len(os.Args) != 5 {
        handleError("Incorrect number of command line arguments")
    }

    mapFile := os.Args[1]
    startStation := os.Args[2]
    endStation := os.Args[3]
    numTrains, err := strconv.Atoi(os.Args[4])
    if err != nil || numTrains <= 0 {
        handleError("Number of trains is not a valid positive integer")
    }

    network, err := parseNetworkMap(mapFile)
    if err != nil {
        handleError(err.Error())
    }

    if _, exists := network.Stations[startStation]; !exists {
        handleError("Start station does not exist")
    }

    if _, exists := network.Stations[endStation]; !exists {
        handleError("End station does not exist")
    }

    if startStation == endStation {
        handleError("Start and end station are the same")
    }

    // Simulate trains on the dynamic path
    simulateTrains(network, startStation, endStation, numTrains)
}