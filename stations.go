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

// BFS to find the shortest paths
func bfs(network *Network, start, end string) [][]string {
    queue := [][]string{{start}}
    visited := make(map[string]bool)
    visited[start] = true
    var paths [][]string

    for len(queue) > 0 {
        path := queue[0]
        queue = queue[1:]
        current := path[len(path)-1]

        if current == end {
            paths = append(paths, path)
            continue
        }

        for _, neighbor := range network.Connections[current] {
            if !visited[neighbor] {
                newPath := append([]string{}, path...)
                newPath = append(newPath, neighbor)
                queue = append(queue, newPath)
                visited[neighbor] = true
            }
        }
    }

    return paths
}

// Function to simulate the train movements with staggered starts
func simulateTrains(paths [][]string, numTrains int) {
    trains := make([]*Train, numTrains)
    for i := 0; i < numTrains; i++ {
        trains[i] = &Train{Name: fmt.Sprintf("T%d", i+1), Current: paths[0][0]} // All start at the initial station
    }

    for turn := 0; ; turn++ {
        var movement []string
        for i, train := range trains {
            path := paths[0] // All trains follow the first found path
            if turn >= i && train.Current != path[len(path)-1] { // Only move the train if it's their turn
                for j, station := range path {
                    if station == train.Current && j < len(path)-1 {
                        nextStation := path[j+1]
                        train.Current = nextStation
                        movement = append(movement, fmt.Sprintf("%s-%s", train.Name, nextStation))
                        break
                    }
                }
            }
        }
        if len(movement) == 0 {
            break // If no train moved, end the simulation
        }
        fmt.Println(strings.Join(movement, " "))
    }
}

// Function to get the index of a train
func trainIndex(trainName string) int {
    index, _ := strconv.Atoi(trainName[1:])
    return index
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

    simulateTrains(paths, numTrains)
}