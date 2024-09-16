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
	stationSectionEncountered := false
	connectionSection := false
	connectionSectionEncountered := false
	coordinates := make(map[string]string) // Key: "x,y", Value: station name

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
			stationSectionEncountered = true
			continue
		}

		if line == "connections:" {
			stationSection = false
			connectionSection = true
			connectionSectionEncountered = true
			continue
		}

		if stationSection {
			match := stationRegex.FindStringSubmatch(line)
			if match == nil {
				return nil, errors.New("Invalid station format: " + line + ". Please remove any invalid characters from station name")
			}
			name, xStr, yStr := match[1], match[2], match[3]
			x, err := strconv.Atoi(xStr)
			if err != nil || x < 0 {
				return nil, errors.New("Invalid X coordinate for station: " + name)
			}
			y, err := strconv.Atoi(yStr)
			if err != nil || y < 0 {
				return nil, errors.New("Invalid Y coordinate for station: " + name)
			}
			if _, exists := network.Stations[name]; exists {
				return nil, errors.New("Duplicate station name: " + name)
			}
			coordKey := fmt.Sprintf("%d,%d", x, y)
			if existingStation, exists := coordinates[coordKey]; exists {
				return nil, errors.New("Stations '" + name + "' and '" + existingStation + "' share the same coordinates (" + coordKey + ")")
			}
			coordinates[coordKey] = name
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

	if !stationSectionEncountered {
		return nil, errors.New("Map does not contain a 'stations:' section")
	}
	if !connectionSectionEncountered {
		return nil, errors.New("Map does not contain a 'connections:' section")
	}

	if len(network.Stations) > 10000 {
		return nil, errors.New("map contains more than 10,000 stations")
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return network, nil
}

func pathExists(start, end string, network *Network) bool {
	visited := make(map[string]bool)
	var dfs func(station string) bool
	dfs = func(station string) bool {
		if station == end {
			return true
		}
		visited[station] = true
		for _, neighbor := range network.Connections[station] {
			if !visited[neighbor] && dfs(neighbor) {
				return true
			}
		}
		return false
	}
	return dfs(start)
}
