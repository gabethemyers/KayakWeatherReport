package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type WeatherResponse struct {
	Wind struct {
		Gust      int    `json:"gust"`
		Speed     int    `json:"speed"`
		Direction string `json:"direction"`
		Arrow     string `json:"arrow"`
		IsLive    bool   `json:"is_live"`
	} `json:"wind"`

	Current struct {
		Speed string `json:"speed"`
		State string `json:"state"`
		Arrow string `json:"arrow"`
	} `json:"current"`

	Tide struct {
		NextType   string  `json:"next_type"`
		NextTime   string  `json:"next_time"`
		NextHeight float64 `json:"next_height"`
	} `json:"tide"`

	Swell struct {
		WaveHeight float64 `json:"wave_height"`
		WavePeriod float64 `json:"wave_period"`
		Direction  string  `json:"direction"`
		Arrow      string  `json:"arrow"`
	} `json:"swell"`
}

// also time/last updated could be added too.

type WindData struct {
	Gust      int
	Speed     int
	Direction string
	Arrow     string
}
type SwellData struct {
	WaveHeight float64
	WavePeriod float64
	Direction  int
}

type NwsHourlyResponse struct {
	Properties struct {
		Periods []struct {
			WindSpeed     string `json:"windSpeed"`     // E.g., "6 mph"
			WindDirection string `json:"windDirection"` // E.g., "W"
			WindGust      string `json:"windGust"`      // E.g., "12 mph" or empty
		} `json:"periods"`
	} `json:"properties"`
}

type NoaaTideResponse struct {
	Predictions []struct {
		Time  string  `json:"t"`        // "2026-07-07 04:00"
		Value float64 `json:"v,string"` // "3.204"
		Type  string  `json:"type"`     // "H" or "L"
	} `json:"predictions"`
}

type TideAndCurrentData struct {
	CurrentState string // "Flooding", "Ebbing", "Slack"
	CurrentSpeed string // "Fast", "Slow", "Not Moving"
	CurrentArrow string // "→", "←", "●"

	// Tide pieces parsed from the response
	NextType   string
	NextTime   string
	NextHeight float64
}

type TidePeak struct {
	Time   string  `json:"time"`
	Height float64 `json:"height"`
	Type   string  `json:"type"`
}

type OpenMeteoResponse struct {
	Current struct {
		Time          string  `json:"time"`
		WaveHeight    float64 `json:"wave_height"`    // feet
		WavePeriod    float64 `json:"wave_period"`    // seconds
		WaveDirection int     `json:"wave_direction"` // degrees
	} `json:"current"`
}

// TODO: concurently call each fetch function
// Example:
//     var wg sync.WaitGroup
//     var forecastData WindData
//     var windErr error
//     wg.Add(1)
//     go func() {
//         defer wg.Done()
//         forecastData, windErr = fetchWind()
//     }()
// do it for all fetch functions
// wg.Wait()
// handle errors

func conditionsHandler(w http.ResponseWriter, r *http.Request) {
	forcastData, err := fetchWind()
	if err != nil {
		http.Error(w, "Fetching Wind failed", http.StatusInternalServerError)
		return
	}

	liveWindSpeed, liveErr := fetchLiveWindSpeed()

	w.Header().Set("Content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	data := WeatherResponse{}

	data.Wind.Gust = forcastData.Gust
	data.Wind.IsLive = false
	data.Wind.Speed = forcastData.Speed
	data.Wind.Direction = forcastData.Direction
	data.Wind.Arrow = forcastData.Arrow

	if liveErr == nil {
		data.Wind.Speed = liveWindSpeed
		data.Wind.IsLive = true
		if liveWindSpeed > data.Wind.Gust {
			data.Wind.Gust = liveWindSpeed // makes sure that wind gust won't ever be less than wind speed.
		}
	}

	tideAndCurrentData, err := fetchCurrent()
	if err != nil {
		http.Error(w, fmt.Sprintf("Fetching Tide and Current failed: %v", err), http.StatusInternalServerError)
		return
	}

	data.Current.Speed = tideAndCurrentData.CurrentSpeed
	data.Current.Arrow = tideAndCurrentData.CurrentArrow
	data.Current.State = tideAndCurrentData.CurrentState

	data.Tide.NextHeight = tideAndCurrentData.NextHeight
	data.Tide.NextTime = tideAndCurrentData.NextTime
	data.Tide.NextType = tideAndCurrentData.NextType

	swellData, err := fetchSwell()
	if err != nil {
		http.Error(w, fmt.Sprintf("Fetching Swell failed: %v", err), http.StatusInternalServerError)
	}

	data.Swell.WaveHeight = swellData.WaveHeight
	data.Swell.WavePeriod = swellData.WavePeriod
	data.Swell.Direction = getCardinalFromDegree(swellData.Direction)
	_, data.Swell.Arrow = getCompassFromCardinal(data.Swell.Direction, true)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode payload: %v", err)
	}
}

// Arrow direction definitions based on the vector's destination (where it is going)
var destinationArrows = map[string]string{
	"N": "↓", "NE": "↙", "E": "←", "SE": "↖",
	"S": "↑", "SW": "↗", "W": "→", "NW": "↘",
	"NNE": "↙", "ENE": "←", "ESE": "↖", "SSE": "↖",
	"SSW": "↗", "WSW": "→", "WNW": "↘", "NNW": "↓",
}

// Opposites map to flip the vector 180 degrees
var opposites = map[string]string{
	"N": "S", "NE": "SW", "E": "W", "SE": "NW",
	"S": "N", "SW": "NE", "W": "E", "NW": "SE",
	"NNE": "SSW", "ENE": "WSW", "ESE": "WNW", "SSE": "NNW",
	"SSW": "NNE", "WSW": "ENE", "WNW": "ESE", "NNW": "SSE",
}

// getCompassFromCardinal returns the cardinal string and the correct arrow.
// Set 'towards' to true for Swell (where it is going).
// Set 'towards' to false for Wind (arrow representing where it originates/blows from).
func getCompassFromCardinal(cardinal string, towards bool) (string, string) {
	if towards {
		return cardinal, destinationArrows[cardinal]
	}

	// If we want the source direction (wind origin), we flip the destination arrow
	oppositeCardinal := opposites[cardinal]
	return cardinal, destinationArrows[oppositeCardinal]
}

func getCardinalFromDegree(degree int) string {
	mapping := map[int]string{
		0: "N", 1: "NE", 2: "E", 3: "SE", 4: "S",
		5: "SW", 6: "W", 7: "NW",
	}

	index := ((degree + 22) / 45) % 8
	return mapping[index]
}

func fetchWind() (WindData, error) {
	var wind WindData
	url := "https://api.weather.gov/gridpoints/MTR/98,58/forecast/hourly"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return wind, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "MyKayakWeatherApp/1.0")

	client := &http.Client{
		Timeout: 6 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return wind, fmt.Errorf("network error hitting NWS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return wind, fmt.Errorf("upstream NWS server returned status code %d", resp.StatusCode)
	}

	var data NwsHourlyResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return wind, fmt.Errorf("failed to decode NWS JSON payload: %w", err)
	}

	if len(data.Properties.Periods) == 0 {
		return wind, fmt.Errorf("no hourly periods received from NWS")
	}

	// Index [0] is managed dynamically by the weather server to represent the current hour
	currentPeriod := data.Properties.Periods[0]

	// Extract the sustained wind integer out of strings like "6 mph"
	var speedMph int
	if currentPeriod.WindSpeed != "" {
		fmt.Sscanf(currentPeriod.WindSpeed, "%d mph", &speedMph)
	}

	// Handle the vanishing windGust property safely
	var gustMph int
	if currentPeriod.WindGust != "" {
		fmt.Sscanf(currentPeriod.WindGust, "%d mph", &gustMph)
	} else {
		// Fallback to base speed if the wind gust key is missing from the timeline frame
		gustMph = speedMph
	}

	// Unpack directional vectors
	_, arrowStr := getCompassFromCardinal(currentPeriod.WindDirection, true)

	// Consolidate properties directly into the return container structure
	wind.Speed = speedMph
	wind.Gust = gustMph
	wind.Direction = currentPeriod.WindDirection
	wind.Arrow = arrowStr

	return wind, nil
}

// returns live wind speed in mph if it exits, otherwise error
func fetchLiveWindSpeed() (int, error) {
	liveWindSpeed := -1
	url := "https://www.ndbc.noaa.gov/data/realtime2/MLSC1.txt"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return liveWindSpeed, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 8 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return liveWindSpeed, fmt.Errorf("network error hitting MLSC1 Station: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return liveWindSpeed, fmt.Errorf("upstream MLSC1 Station returned status code %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)

	if err := scanner.Err(); err != nil {
		return liveWindSpeed, fmt.Errorf("error reading station stream: %w", err)
	}

	for i := 0; i < 3; i++ {
		scanner.Scan()
	}

	dataLine := scanner.Text()

	fields := strings.Fields(dataLine)

	if len(fields) < 7 {
		return liveWindSpeed, errors.New("malformed data line recived from station")
	}

	windspeed := fields[6]

	if windspeed == "MM" {
		return liveWindSpeed, errors.New("live wind speed is missing")
	}

	windSpeedMeters, convErr := strconv.ParseFloat(windspeed, 64)

	if convErr != nil {
		return liveWindSpeed, fmt.Errorf("error converting wind speed to a float")
	}

	liveWindSpeed = int(math.Round(windSpeedMeters * 2.23693629))

	return liveWindSpeed, nil
}

// current is going to be inferred from tide hi and lo spots.
// current is only going to be flooding, ebbing, or slack. The arrow will be also inferred from these.
// current speed is also going to have to be a status instead of a number

// TODO: Also figure out if this current architeture is right, where i have wind data current data and then combine into the final response.
func fetchCurrent() (TideAndCurrentData, error) {
	var data NoaaTideResponse
	// TODO: fetch data using range and a begin date in this format: begin_date=20120415&range=48 	Retrieves data for 48 hours beginning on April 15, 2012
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)
	url := fmt.Sprintf("https://api.tidesandcurrents.noaa.gov/api/prod/datagetter?station=9413623&product=predictions&datum=MLLW&interval=hilo&time_zone=lst_ldt&units=english&format=json&begin_date=%s&range=72", yesterday.Format("20060102"))

	resp, err := http.Get(url)
	if err != nil {
		return TideAndCurrentData{}, fmt.Errorf("network error: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return TideAndCurrentData{}, fmt.Errorf("upstream server returned status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return TideAndCurrentData{}, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	// first turn all properties into TidePeak and into a slice

	if len(data.Predictions) < 1 {
		return TideAndCurrentData{}, fmt.Errorf("upstream server returned no tide predictions")
	}

	var timeline []TidePeak

	for _, property := range data.Predictions {
		timeline = append(timeline, TidePeak{
			Time:   property.Time,
			Height: property.Value,
			Type:   property.Type,
		})
	}

	// Then find the next tide based on current time
	var nextTide TidePeak
	var previousTide TidePeak

	var nextTideTime time.Time
	var previousTideTime time.Time

	for i, tide := range timeline {
		tideTime, err := time.ParseInLocation("2006-01-02 15:04", tide.Time, time.Local)
		if err != nil {
			continue
		}

		if tideTime.After(today) {
			nextTide = tide
			nextTideTime = tideTime
			previousTide = timeline[i-1]
			previousTideTime, err = time.Parse("2006-01-02 15:04", previousTide.Time)
			if err != nil {
				return TideAndCurrentData{}, fmt.Errorf("error parsing previous tide time: %w", err)
			}
			break
		}
	}

	if nextTide.Time == "" || previousTide.Time == "" {
		return TideAndCurrentData{}, fmt.Errorf("next tide could not be found")
	}

	totalWindow := nextTideTime.Sub(previousTideTime)
	progress := today.Sub(previousTideTime)

	percentage := progress.Seconds() / totalWindow.Seconds()

	var result TideAndCurrentData

	result.NextHeight = nextTide.Height
	result.NextTime = nextTide.Time
	result.NextType = nextTide.Type

	if nextTide.Type == "H" {
		result.NextType = "High"
	} else {
		result.NextType = "Low"
	}

	// Then calculate current data

	switch {
	case percentage <= 0.15 || percentage >= 0.85:
		result.CurrentSpeed = "Not Moving"
	case (percentage > 0.15 && percentage <= 0.35) || (percentage >= 0.65 && percentage < 0.85):
		result.CurrentSpeed = "Slow"
	default:
		result.CurrentSpeed = "Fast"
	}

	if result.CurrentSpeed == "Not Moving" {
		result.CurrentState = "Slack"
		result.CurrentArrow = ""
	} else if nextTide.Type == "H" {
		result.CurrentState = "Flooding"
		result.CurrentArrow = "→" // Moving East, pushing deeper into the slough
	} else {
		result.CurrentState = "Ebbing"
		result.CurrentArrow = "←" // Moving West, pulling into the ocean
	}

	return result, nil
}

func fetchSwell() (SwellData, error) {
	data := OpenMeteoResponse{}
	url := "https://marine-api.open-meteo.com/v1/marine?latitude=36.80&longitude=-121.79&current=wave_height,wave_period,wave_direction&timezone=auto&length_unit=imperial"

	resp, err := http.Get(url)
	if err != nil {
		return SwellData{}, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SwellData{}, fmt.Errorf("upstream server returned error code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return SwellData{}, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	result := SwellData{
		WaveHeight: data.Current.WaveHeight,
		WavePeriod: data.Current.WavePeriod,
		Direction:  data.Current.WaveDirection,
	}

	return result, nil
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/api/v1/conditions", conditionsHandler)

	r.Get("/api/v1/conditions/mock", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
		"wind": {
			"gust": 10,
			"speed": 10,
			"direction": "W",
			"arrow": "→",
			"is_live": true
		},
		"current": {
			"speed": "Not Moving",
			"state": "Slack",
			"arrow": ""
		},
		"tide": {
			"next_type": "L",
			"next_time": "2026-07-15 06:12",
			"next_height": -1.433
		},
		"swell": {
			"wave_height": 3.412,
			"wave_period": 8.2,
			"direction": "W",
			"arrow": "→"
		}
	}`))
	})

	// 1. Log that the server is trying to start
	log.Println("Starting kayak weather API server on :8080...")

	// 2. Wrap ListenAndServe in log.Fatal so failure prints to your console
	log.Fatal(http.ListenAndServe(":8080", r))
}
