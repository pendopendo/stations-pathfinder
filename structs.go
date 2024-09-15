package main

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
