# Stations - A pathfinding solution for optimal movement management

This is a Command-Line Tool (CLT) meant to be used for optimising movement for multiple object between set points by predefined connections. In this case, it can move a given number of trains between stations defined by the set connections between stations. 
## Functionality

- Parsing - Upon running the CLT, the first thing to happen is parsing through the provided that contains the set points and the connections betwwen them : in this instance the station names and the connections betwwen them (the "rails"). This will generate the paths through which movement will occur for all stations and all connections.

- Pathfinding - The next step is  generating a path between the provided stations using the generated connections. It uses a depth-first approach to generate all possible connections to travel through from the starting station to the end station, and then chooses a combination of paths to take that would be most optimal.

- Travel - The CLT then uses the chosen paths and assigns them to the trains upon leaving the station, making sure no erroneous movement takes place. 

- Troubleshooting - The CLT also checks that the provided inputs are correct and properly formatted for it to function correctly, and gives appropriate error messages in required cases.

## Running the Application

- To run the CLT, simply type the proper command to the command line, one example being:
  * ```go run . network.map waterloo st_pancras 2``` where:
  * ```network.map```: the file containing the stations and connections
  * ```waterloo```: starting station
  * ```st_pancras```: end station
  * ```2```: number of trains

- There is also the option to run a set of prescripted commands to test the CLT's functionality with the following command:
  * ```go run . test[#]``` where:
  * ```#```: marks numbers from ```0...7```, with "0" running tests 1...7 simultaneously
  * ```go run . 10000``` will run a simulation with map map file that has more than 10000 stations in it, properly displaying that specific error handling function. it can also be run with the standard command, exchanging the ```network.map``` argument with ```10000.map```