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
	coordinates := make(map[string]string)

	//flags
	stationformatflag := false
	xcoordinateflag := false
	ycoordinateflag := false
	duplicatestationflag := false
	samecoordinatesflag := false
	invalidconnectionflag := false
	sameconnectionflag := false
	nonexistingstation1flag := false
	nonexistingstation2flag := false
	duplicateconnectionflag := false

	//variables for error handling
	var (
		stationformatline             string
		xcoordinatename               string
		ycoordinatename               string
		duplicatestationname          string
		samecoordinatename            string
		samecoordinateexistingStation string
		samecoordinatecoordKey        string
		invalidconnectionline         string
		sameconnectionstation1        string
		nonexistingstation1           string
		nonexistingstation2           string
		duplicatestation1             string
		duplicatestation2             string
	)

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
			connectionSection = true
			stationSection = false
			connectionSectionEncountered = true
			continue
		}

		if stationSection {
			match := stationRegex.FindStringSubmatch(line)
			if match == nil {
				stationformatflag = true
				stationformatline = line
				continue
			}
			name, xStr, yStr := match[1], match[2], match[3]
			x, err := strconv.Atoi(xStr)
			if err != nil || x < 0 {
				xcoordinateflag = true
				xcoordinatename = name
			}
			y, err := strconv.Atoi(yStr)
			if err != nil || y < 0 {
				ycoordinateflag = true
				ycoordinatename = name
			}
			if _, exists := network.Stations[name]; exists {
				duplicatestationflag = true
				duplicatestationname = name
			}
			coordKey := fmt.Sprintf("%d,%d", x, y)
			if existingStation, exists := coordinates[coordKey]; exists {
				samecoordinatesflag = true
				samecoordinatename = name
				samecoordinateexistingStation = existingStation
				samecoordinatecoordKey = coordKey
			}
			coordinates[coordKey] = name
			network.Stations[name] = &Station{Name: name, X: x, Y: y}
		} else if connectionSection {
			match := connectionRegex.FindStringSubmatch(line)
			if match == nil {
				invalidconnectionflag = true
				invalidconnectionline = line
				continue
			}
			station1, station2 := match[1], match[2]
			if station1 == station2 {
				sameconnectionflag = true
				sameconnectionstation1 = station1
			}
			if _, exists := network.Stations[station1]; !exists {
				nonexistingstation1flag = true
				nonexistingstation1 = station1
			}
			if _, exists := network.Stations[station2]; !exists {
				nonexistingstation2flag = true
				nonexistingstation2 = station2
			}
			if contains(network.Connections[station1], station2) || contains(network.Connections[station2], station1) {
				duplicateconnectionflag = true
				duplicatestation1 = station1
				duplicatestation2 = station2
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

	if duplicatestationflag {
		return nil, errors.New("Duplicate station name: " + duplicatestationname)
	}

	if xcoordinateflag {
		return nil, errors.New("Invalid X coordinate for station: " + xcoordinatename)
	}

	if ycoordinateflag {
		return nil, errors.New("Invalid Y coordinate for station: " + ycoordinatename)
	}

	if stationformatflag {
		return nil, errors.New("Invalid station format: " + stationformatline + ".")
	}

	if samecoordinatesflag {
		return nil, errors.New("Stations '" + samecoordinatename + "' and '" + samecoordinateexistingStation + "' share the same coordinates (" + samecoordinatecoordKey + ")")
	}

	if invalidconnectionflag {
		return nil, errors.New("Invalid connection format: " + invalidconnectionline)
	}

	if sameconnectionflag {
		return nil, errors.New("Connection between the same station: " + sameconnectionstation1)
	}

	if nonexistingstation1flag {
		return nil, errors.New("Connection with non-existing station: " + nonexistingstation1)
	}

	if nonexistingstation2flag {
		return nil, errors.New("Connection with non-existing station: " + nonexistingstation2)
	}

	if duplicateconnectionflag {
		return nil, errors.New("Duplicate connection between " + duplicatestation1 + " and " + duplicatestation2)
	}

	// break
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
