package main

import (
	"github.com/iarenzana/ft/ft"
	"github.com/codegangsta/cli"
	"os"
	"fmt"
)

/*
Constant definitions
*/

const version string = "0.2"
const author string = "Ismael Arenzana"
const email string = "iarenzana@gmail.com"
const appName string = "ft"
const appDescription string = "Command-line flight tracker"

/*
Main function
*/

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = appDescription
	app.Email = email
	app.Author = author
	app.Version = version
	app.Commands = []cli.Command{
		{
			Name:      "airportinfo",
			ShortName: "a",
			Usage:     "Display Airport Information",
			Action: func(c *cli.Context) {
				airportVal := ft.ValidateAirportCode(c.Args()[0])
				if airportVal == -2 {
					fmt.Println("Error, could not validate airport code.")
					os.Exit(1)
				}

				ft.AirportInfoEval(c.Args()[0])

			},
		},
		{
			Name:      "track",
			ShortName: "t",
			Usage:     "Track a Flight",
			Action: func(c *cli.Context) {
			ft.FlightTrackingEval(c.Args()[0])

			},
		},
		{
			Name:      "airlineinfo",
			ShortName: "l",
			Usage:     "Airline Information",
			Action: func(c *cli.Context) {

			},
		},
	}

	app.Run(os.Args)
}
