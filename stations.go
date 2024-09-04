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

// BFS to find all paths from start to end
func bfs(network *Network, start, end string) [][]string {
    var paths [][]string
    queue := []struct {
        path    []string
        visited map[string]bool
    }{
        {path: []string{start}, visited: map[string]bool{start: true}},
    }

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]
        currentPath := current.path
        visited := current.visited
        currentStation := currentPath[len(currentPath)-1]

        if currentStation == end {
            paths = append(paths, currentPath)
            continue
        }

        for _, neighbor := range network.Connections[currentStation] {
            if !visited[neighbor] {
                newVisited := make(map[string]bool)
                for k, v := range visited {
                    newVisited[k] = v
                }
                newVisited[neighbor] = true
                newPath := append([]string{}, currentPath...)
                newPath = append(newPath, neighbor)
                queue = append(queue, struct {
                    path    []string
                    visited map[string]bool
                }{
                    path:    newPath,
                    visited: newVisited,
                })
            }
        }
    }

    return paths
}

// simulateTrains simulates train movements with proper conditions, preventing simultaneous use of the same track
func simulateTrains(paths [][]string, numTrains int) {
    // Create an array of Train pointers to hold the trains
    trains := make([]*Train, numTrains)

    // Assign each train to a path; if there are more trains than paths, cycle through paths
    pathAssignments := make([]int, numTrains) // Array to keep track of which path each train is assigned to
    for i := 0; i < numTrains; i++ {
        // Create a train with a unique name and set its current station to the start of its assigned path
        trains[i] = &Train{Name: fmt.Sprintf("T%d", i+1), Current: paths[i%len(paths)][0]}
        pathAssignments[i] = i % len(paths) // Assign the train to a path index (cycling through if more trains than paths)
    }

    turn := 1 // Initialize the turn counter
    for {     // Main simulation loop
        var movement []string                     // List to keep track of movements made in the current turn
        allTrainsAtDestination := true            // Flag to check if all trains have reached their destinations
        occupiedStations := make(map[string]bool) // Track stations occupied by trains during the turn
        usedSegments := make(map[string]bool)     // Track segments used by trains during the turn

        // Print current status of each train and mark their current positions as occupied
        for i, train := range trains {
            assignedPath := paths[pathAssignments[i]] // Get the assigned path for this train
            fmt.Printf("Train %s on path %v, at station %s\n", train.Name, assignedPath, train.Current)
            occupiedStations[train.Current] = true // Mark current station as occupied

            // Check if the train has reached its destination
            if train.Current != assignedPath[len(assignedPath)-1] {
                allTrainsAtDestination = false
            }
        }

        // Exit early if all trains have reached their destinations
        if allTrainsAtDestination {
            fmt.Println("All trains have reached their destinations. Simulation ending.")
            break // Stop the loop if all trains have reached their destinations
        }

        fmt.Printf("\nTurn %d:\n", turn) // Print the current turn

        // Iterate through each train to check and perform movement
        for i, train := range trains {
            assignedPath := paths[pathAssignments[i]] // Get the assigned path for this train

            // Check if the train has not yet reached the final station of its path
            if train.Current != assignedPath[len(assignedPath)-1] {
                fmt.Printf("  Checking movement for Train %s...\n", train.Name)

                // Loop through the path to find the current station and determine the next station
                for j, station := range assignedPath {
                    if station == train.Current && j < len(assignedPath)-1 {
                        nextStation := assignedPath[j+1] // Get the next station on the path
                        segment := fmt.Sprintf("%s-%s", station, nextStation)

                        // Allow trains to enter the end station without restrictions
                        if (nextStation == assignedPath[len(assignedPath)-1] || !occupiedStations[nextStation]) && !usedSegments[segment] {
                            // Move the train to the next station
                            train.Current = nextStation
                            movement = append(movement, fmt.Sprintf("%s-%s", train.Name, nextStation)) // Record the movement
                            fmt.Printf("  Train %s moved from %s to %s\n", train.Name, station, nextStation)
                            occupiedStations[nextStation] = true  // Mark the next station as occupied
                            occupiedStations[station] = false     // Free up the previous station
                            usedSegments[segment] = true          // Mark the segment as used for this turn
                        } else {
                            fmt.Printf("  Train %s cannot move to %s; station is occupied or track is already used.\n", train.Name, nextStation)
                        }
                        break // Exit the loop after checking the movement for the train
                    }
                }
            } else {
                // This else block triggers when the train has reached its destination.
                fmt.Printf("  Train %s has reached its destination.\n", train.Name)
            }
        }

        // If no train moved during this turn and some trains are not at their destination, continue the loop
        if len(movement) == 0 {
            fmt.Println("All trains are stuck this turn but may move in subsequent turns.")
        }

        // Print the movements made during the turn
        fmt.Printf("Movements this turn: %s\n", strings.Join(movement, " "))
        turn++ // Increment the turn counter
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

    paths := bfs(network, startStation, endStation)
    if len(paths) == 0 {
        handleError("No path between the start and end stations")
    }

    fmt.Println("Possible Paths:")
    for _, path := range paths {
        fmt.Println(path)
    }

    // Simulate trains on all found paths
    simulateTrains(paths, numTrains)
}