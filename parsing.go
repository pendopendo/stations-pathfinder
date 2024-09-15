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
		return nil, errors.New("map contains more than 10,000 stations")
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
