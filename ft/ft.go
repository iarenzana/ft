package ft

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ttacon/chalk"
)

//FlightAwareBase var with the base URI for the Flight data server
var FlightAwareBase = "https://flightxml.flightaware.com/json/FlightXML2/"

//UserHome get env for user home
var UserHome = os.Getenv("HOME")

//FlightAwareAPIKey key for the flight data service.
var FlightAwareAPIKey = os.Getenv("FLIGHTAWARE_API_KEY")

//FlightAwareAPIUser user for the flight data service.
var FlightAwareAPIUser = os.Getenv("FLIGHTAWARE_API_USER")

//BaseDir base data directory.
var BaseDir = filepath.Dir(UserHome + "/.ft/")

//OutputFileAirports file with airport data.
var OutputFileAirports = BaseDir + "/airports.dat"

//OutputFileAirlines file with airline data.
var OutputFileAirlines = BaseDir + "/airlines.dat"

//OutputFileRoutes file with route data.
var OutputFileRoutes = BaseDir + "/routes.dat"

//AirportInfoEval - Main airportInfo function
func AirportInfoEval(inputAirport string) {

	if _, err := os.Stat(OutputFileAirports); os.IsNotExist(err) {
		getStaticData()
	}

	var ai airportInformation

	airportIndex, err := getAirportIndex(strings.ToUpper(inputAirport))

	if err != nil || airportIndex == -2 {
		fmt.Println(chalk.Red, "Airport Unknown.")
		return
	}
	ai = getAirportData(airportIndex)
	metar := getAirportMETAR(ai.airportICAOCode)

	fmt.Print(chalk.Blue, "Airport Name: ", chalk.Green, ai.airportName, "\n")
	fmt.Print(chalk.Blue, "Location    : ", chalk.Green, ai.airportCity, ", ", chalk.Green, ai.airportCountry, "\n")
	fmt.Print(chalk.Blue, "Altitude    : ", chalk.Green, strings.TrimSpace(strconv.Itoa(ai.airportAltitude)), "ft", "\n")
	fmt.Print(chalk.Blue, "ICAO        : ", chalk.Green, ai.airportICAOCode, chalk.Blue, " IATA: ", chalk.Green, ai.airportIATACode, "\n")
	fmt.Print(chalk.Blue, "METAR       : ", chalk.Green, metar, "\n")

}

//FlightTrackingEval - Flight track evaluation
func FlightTrackingEval(inputFlightToTrack string) {

	if _, err := os.Stat(OutputFileRoutes); os.IsNotExist(err) {
		getStaticData()
	}

	if checkEnvVariables() != 0 {
		fmt.Println(chalk.Red, "Please set the FLIGHTAWARE_API_KEY and FLIGHTAWARE_API_USER variables.")
		os.Exit(-1)
	}

	flightData := getFlightData(strings.ToUpper(inputFlightToTrack))
	for i := range flightData.FlightInfoExResult.Flights {
		fmt.Println(chalk.Blue, "Origin City      : ", chalk.Green, flightData.FlightInfoExResult.Flights[i].OriginCity)
		fmt.Println(chalk.Blue, "Destination City : ", chalk.Green, flightData.FlightInfoExResult.Flights[i].DestinationCity)
		fmt.Println(chalk.Blue, "Aircraft Type    : ", chalk.Green, flightData.FlightInfoExResult.Flights[i].Aircrafttype)
		t2 := int64(flightData.FlightInfoExResult.Flights[i].Estimatedarrivaltime)
		actualETA := time.Unix(t2, 0)

		fmt.Println(chalk.Blue, "Filed Arrival    : ", chalk.Green, flightData.FlightInfoExResult.Flights[i].FiledEte)
		fmt.Println(chalk.Blue, "Scheduled Arrival: ", chalk.Green, actualETA)
		fmt.Println(chalk.Blue, "Route            : ", chalk.Green, flightData.FlightInfoExResult.Flights[i].Route)
	}

}

//TODO: Write test Unit
func AirlineInfo(airlineInfo string) {
	if _, err := os.Stat(OutputFileAirlines); os.IsNotExist(err) {
		getStaticData()
	}
	if checkEnvVariables() != 0 {
		fmt.Println(chalk.Red, "Please set the FLIGHTAWARE_API_KEY and FLIGHTAWARE_API_USER variables.")
		os.Exit(-1)
	}

	airlineData, err := getAirlineData(airlineInfo)
	if err != nil {
		fmt.Println(chalk.Red, "Error getting airline data: ", err)
		os.Exit(1)
	}

	if airlineData.AirlineID == 0 {
		fmt.Println(chalk.Red, "Airline not found.")
		return
	}
	fmt.Println(chalk.Blue, "Airline Name	   : ", chalk.Green, airlineData.AirlineName)
	fmt.Println(chalk.Blue, "ICAO              : ", chalk.Green, airlineData.AirlineICAO, " IATA: ", airlineData.AirlineIATA)
	fmt.Println(chalk.Blue, "Airline Callsign  : ", chalk.Green, airlineData.AirlineCallsign)
	fmt.Println(chalk.Blue, "Airline Country   : ", chalk.Green, airlineData.AirlineCountry)
	if airlineData.AirlineActive {
		fmt.Println(chalk.Blue, "Airline Status    : ", chalk.Green, "Active")
	} else {
		fmt.Println(chalk.Blue, "Airline Status    : ", chalk.Green, "Inactive")
	}
}

//TODO: Write test Unit
func getAirlineData(airline string) (AirlineDataStruct, error) {
	var airlineData AirlineDataStruct

	csvfile, err := os.Open(OutputFileAirlines)
	if err != nil {
		fmt.Println(chalk.Red, err)
		return airlineData, err
	}
	defer csvfile.Close()
	reader := csv.NewReader(csvfile)

	reader.FieldsPerRecord = -1

	rawCSVdata, err := reader.ReadAll()

	if err != nil {
		fmt.Println(err)
		return airlineData, err
	}
	for _, each := range rawCSVdata {
		code := each[4]

		if code == airline {
			airlineData.AirlineID, _ = strconv.Atoi(each[0])
			airlineData.AirlineName = each[1]
			airlineData.AirlineAlias = each[2]
			airlineData.AirlineIATA = each[3]
			airlineData.AirlineICAO = each[4]
			airlineData.AirlineCallsign = each[5]
			airlineData.AirlineCountry = each[6]
			if each[7] == "Y" {
				airlineData.AirlineActive = true
			} else {
				airlineData.AirlineActive = false
			}
		}
	}

	return airlineData, nil
}

//ValidateAirportCode - Make sure the airport code is correctly formatted -2 - Invalid code. 1 - ICAO code. 0 - IATA code.
func ValidateAirportCode(airportCode string) int32 {

	var airportIndex int32 = -2
	var airportCodeBytes = len([]rune(airportCode))

	if airportCodeBytes == 3 {
		airportIndex = 0
	} else if airportCodeBytes == 4 {
		airportIndex = 1
	} else {
		airportIndex = -2
	}
	return airportIndex
}

/*
Get basic airport info
*/
func getAirportData(airportIndex int) airportInformation {
	var airportDataResult airportInformation

	csvfile, err := os.Open(OutputFileAirports)
	if err != nil {
		fmt.Println(chalk.Red, err)
		os.Exit(-1)
	}
	defer csvfile.Close()
	reader := csv.NewReader(csvfile)

	reader.FieldsPerRecord = -1

	rawCSVdata, err := reader.ReadAll()

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	for _, each := range rawCSVdata {
		index, _ := strconv.Atoi(each[0])
		if index == airportIndex {
			lat, _ := strconv.ParseFloat(each[6], 32)
			lon, _ := strconv.ParseFloat(each[7], 32)
			airportDataResult.airportName = each[1]
			airportDataResult.airportICAOCode = each[5]
			airportDataResult.airportIATACode = each[4]
			airportDataResult.airportIndex = airportIndex
			airportDataResult.airportCountry = each[3]
			airportDataResult.airportCity = each[2]
			airportDataResult.airportLat = lat
			airportDataResult.airportLong = lon
			airportDataResult.airportAltitude, _ = strconv.Atoi(each[8])
		}
	}

	return airportDataResult
}

/*
Get index of an airport code so we can gather all the data by index
*/
func getAirportIndex(airportCode string) (int, error) {
	csvfile, err := os.Open(OutputFileAirports)
	if err != nil {
		fmt.Println(chalk.Red, err)
		return -2, err
	}
	defer csvfile.Close()
	reader := csv.NewReader(csvfile)

	reader.FieldsPerRecord = -1

	rawCSVdata, err := reader.ReadAll()

	if err != nil {
		return -2, err
	}
	for _, each := range rawCSVdata {
		if each[4] == airportCode || each[5] == airportCode {
			index, _ := strconv.Atoi(each[0])
			return index, nil
		}
	}
	return -2, nil
}

/*
Function to get an airport METAR
*/
func getAirportMETAR(airportICAOCode string) string {
	//	var METARURL string = "https://aviationweather.gov/adds/dataserver_current/httpparam?dataSource=metars&requestType=retrieve&format=xml&stationString=" + airportICAOCode + "&hoursBeforeNow=1"
	var METARURL = "http://dev.geonames.org/weatherIcaoJSON?ICAO=" + airportICAOCode
	met := Metar{}

	resp, err := http.Get(METARURL)
	if err != nil {
		fmt.Println(chalk.Red, "Error getting METAR!")
		os.Exit(-1)
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	if err != nil {
		os.Exit(-1)
	}

	err2 := json.Unmarshal([]byte(string(data)), &met)
	if err2 != nil {
		fmt.Println(chalk.Red, "Error parsing data! ", err2)
		os.Exit(-1)
	}
	return met.WeatherObservation.Observation
}

func getFlightData(flightNumber string) flightInformation {
	flightInfo := flightInformation{}
	var flightInfoURL = FlightAwareBase + "FlightInfoEx?ident=" + flightNumber + "&howMany=1&offset=2"
	client := &http.Client{}
	req, err := http.NewRequest("GET", flightInfoURL, nil)
	req.SetBasicAuth(FlightAwareAPIUser, FlightAwareAPIKey)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error %v", err)
		os.Exit(-1)
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal([]byte(string(bodyText)), &flightInfo)
	if err != nil {
		fmt.Println(chalk.Red, "Error parsing data! ", err)
		os.Exit(-1)
	}

	for i := range flightInfo.FlightInfoExResult.Flights {
		//		fmt.Println(flightInfo.FlightInfoExResult.Flights[i].Ident)
		_ = i
	}

	return flightInfo
}

/*
================== Some standard trivial functions to avoid repetition. =============================
*/

/*
Download data from openflights.org to not make much use of the FlightAware API ($$)
*/
func getStaticData() int {
	os.MkdirAll(BaseDir, 0777)
	downloadFromURL("https://raw.githubusercontent.com/jpatokal/openflights/master/data/airports.dat", OutputFileAirports)
	downloadFromURL("https://raw.githubusercontent.com/jpatokal/openflights/master/data/airlines.dat", OutputFileAirlines)
	downloadFromURL("https://raw.githubusercontent.com/jpatokal/openflights/master/data/routes.dat", OutputFileRoutes)
	return 0
}

/*
Download a file from the URL
*/
func downloadFromURL(url string, fileName string) {

	// TODO: check file existence first with io.IsExist
	output, err := os.Create(filepath.FromSlash(fileName))

	if err != nil {
		fmt.Println(chalk.Red, "Error while creating", fileName, "-", err)
		return
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Println(chalk.Red, "Error while downloading", url, "-", err)
		return
	}
	defer response.Body.Close()

	bytesRead, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println(chalk.Red, "Error while downloading", url, "-", err)
		return
	}
	_ = bytesRead
}

//Check that all environment variables are correctly set.
func checkEnvVariables() int {
	if FlightAwareAPIKey == "" || FlightAwareAPIUser == "" {
		return 1
	}
	return 0
}
