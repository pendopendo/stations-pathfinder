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
            fmt.Printf("Train %s: Path found: %v\n", trainName, currentPath)
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
        fmt.Printf("Train %s: No path found.\n", trainName)
        return nil
    }

    // Initialize variables for shortest and alternative paths
    var shortestPath []string
    var shortestPathLength int = 10000 //leave at 10000 since max map 10000 stations
    var alternativePath []string
    var alternativePathLength int = 10000 //leave at 10000 since max map 10000 stations
    fmt.Printf("Initial SPF: %d\n", shortestPathLength)

    numTrains := len(trains)
    currentTrain := 0
    for i, train := range trains {
        if train.Name == trainName {
            currentTrain = i + 1
            break
        }
    }

    if currentTrain == 0 {
        fmt.Printf("Train %s not found in the train list.\n", trainName)
        return nil
    }

    // Find the shortest path and alternative paths
    for _, path := range allPaths {
        available := true
        pathLength := len(path)
        // Check only the segment between the current and next station
        if len(path) > 1 {
            segment := fmt.Sprintf("%s-%s", path[0], path[1])
            if occupiedStations[path[1]] || usedSegments[segment] {
                available = false
            }   
        }

        // Debugging print statements
        fmt.Printf("Train %s: Checking path %v, Length: %d\n", trainName, path, pathLength)

        // Update the shortest path and its length
        if available {
            if pathLength < shortestPathLength {
                fmt.Printf("pathlength %d < shortestpathlength %d\n", pathLength, shortestPathLength)
                shortestPath = path
                shortestPathLength = pathLength
                fmt.Printf("Train %s: New shortest path: %v, Length: %d\n", trainName, shortestPath, shortestPathLength)
            }
        } else {
            if pathLength < alternativePathLength {
                fmt.Printf("pathlength %d < alternativepathlength %d\n", pathLength, alternativePathLength)
                alternativePath = path
                alternativePathLength = pathLength
                fmt.Printf("Train %s: New alternative path: %v, Length: %d\n", trainName, alternativePath, alternativePathLength)
            }
        }
    }
    fmt.Printf("pealeT%dnewshortestpath\n", currentTrain)
    // Calculate threshold using the shortest path length
    threshold := numTrains - currentTrain + shortestPathLength
    fmt.Printf("Train %s: numTrains %d - currentTrain %d + shortestPathLength %d = Calculated threshold: %d\n", trainName, numTrains, currentTrain, shortestPathLength, threshold)

    // If the shortest path is blocked, compare the cumulative lengths
    if shortestPath == nil && alternativePath != nil {
        // Calculate the cumulative length for the alternative path
        cumulativeLength := alternativePathLength * numTrains
        // Calculate the cumulative length for the shortest path if it were to wait
        waitLength := shortestPathLength * numTrains

        fmt.Printf("Train %s: Cumulative length of alternative path: %d\n", trainName, cumulativeLength)
        fmt.Printf("Train %s: Waiting length for shortest path: %d\n", trainName, waitLength)

        // Choose the path with the lesser cumulative length
        if cumulativeLength < waitLength {
            shortestPath = alternativePath
            fmt.Printf("Train %s: Alternative path chosen over waiting: %v\n", trainName, shortestPath)
        }
    }

    // If no available path was found, return nil
    if shortestPath == nil {
        fmt.Printf("Train %s: No available path found.\n", trainName)
        return nil
    }

    // Print all found paths and the selected shortest path
    fmt.Printf("Train %s: Found paths:\n", trainName)
    for _, path := range allPaths {
        fmt.Printf("Path: %v, Length: %d\n", path, len(path))
    }
    fmt.Printf("Train %s: Selected path: %v\n", trainName, shortestPath)

    // Return the shortest available path
    return shortestPath
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