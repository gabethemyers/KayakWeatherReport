# kayak-weather-api

Go server that aggregates marine weather data for Elkhorn Slough, Moss Landing, CA.

## Endpoint

**`GET /api/v1/conditions`**

Returns a JSON object with wind, tide, current, and swell data.

Also available: `GET /api/v1/conditions/mock` returns a hardcoded response for development.

### Response Schema

```jsonc
{
  "wind": {
    "gust": 12,          // mph
    "speed": 8,          // mph
    "direction": "W",    // cardinal or intercardinal (N, NE, E, etc.)
    "arrow": "→",        // Unicode arrow representing wind direction
    "is_live": true      // true = reading from NDBC buoy, false = NWS forecast
  },
  "current": {
    "speed": "Slow",     // "Not Moving" | "Slow" | "Fast"
    "state": "Ebbing",   // "Flooding" | "Ebbing" | "Slack"
    "arrow": "←"         // "" (Slack) | "→" (Flooding) | "←" (Ebbing)
  },
  "tide": {
    "next_type": "High",   // "High" | "Low"
    "next_time": "2026-07-18 14:23",  // lst_ldt (local standard/daylight)
    "next_height": 4.2     // feet relative to MLLW
  },
  "swell": {
    "wave_height": 3.4,   // feet
    "wave_period": 8.2,   // seconds
    "direction": "W",     // cardinal direction
    "arrow": "→"          // Unicode arrow — points where the swell is traveling
  }
}
```

### Error Handling

If any upstream fetch fails, the endpoint returns a **500** with a message indicating which source failed. Note: a failed live wind fetch does not return a 500 — the API falls back to the NWS forecast and sets `is_live` to `false`.

## Data Sources

### Wind Forecast — NWS API

Fetches the current hour from the NWS hourly gridpoint forecast for Monterey Bay.

- **URL:** `https://api.weather.gov/gridpoints/MTR/98,58/forecast/hourly`
- **Wind:** Speed, gust, and direction from `periods[0]`
- **Timeout:** 6 seconds
- **Edge case:** If the NWS response omits `windGust`, the API falls back to the wind speed value.

### Live Wind — NDBC Station MLSC1

Scrapes the latest observation from the Moss Landing NDBC station. Overrides the NWS forecast speed when available.

- **URL:** `https://www.ndbc.noaa.gov/data/realtime2/MLSC1.txt`
- **Format:** Tabular text file, reads the third data line (after two header rows)
- **Timeout:** 8 seconds
- **Edge cases:**
  - Field `6` (wind speed) may be `"MM"` (missing) — returns an error and the NWS forecast is used instead
  - Malformed lines return an error

### Tide Predictions — NOAA Tides & Currents

Pulls 72 hours of hi/lo tide predictions and finds the next tide relative to the current time.

- **URL:** `https://api.tidesandcurrents.noaa.gov/api/prod/datagetter?station=9413623&product=predictions&datum=MLLW&interval=hilo&time_zone=lst_ldt&units=english&format=json&begin_date={yesterday}&range=72`
- **Station:** 9413623 (Monterey Harbor)
- **Datum:** MLLW (Mean Lower Low Water)

### Current Inference

Current direction and speed are derived from the tide timeline, not from a separate sensor.

**Direction logic:**
- If next tide is **High** → Flooding (water moving East into the slough) → `→`
- If next tide is **Low** → Ebbing (water moving West to the ocean) → `←`
- If slack (not moving) → arrow is empty string

**Speed zones** are calculated by where we are in the window between the previous and next tide peak:

| Position in window | Speed |
|---|---|
| First/last 15% | Not Moving (Slack) |
| 15–35% or 65–85% | Slow |
| 35–65% | Fast |

### Swell — Open-Meteo Marine API

- **URL:** `https://marine-api.open-meteo.com/v1/marine?latitude=36.80&longitude=-121.79&current=wave_height,wave_period,wave_direction&length_unit=imperial`
- **Fields:** Wave height (ft), period (s), direction (degrees → cardinal)

## Arrow Semantics

Wind and swell arrows point in the direction the wind or swell is moving toward.

| Wind direction | Arrow |
|---|---|
| N | ↓ |
| S | ↑ |
| E | ← |
| W | → |

The mapping also supports 16-point compass directions (NE, SE, SW, NW, etc.).

## Upstream Dependencies

All four sources are public APIs with no authentication required. The API has no database, no cache, and no external configuration beyond the `TZ` environment variable (should be `America/Los_Angeles` to match the tide station's timezone).

## Running

```bash
TZ=America/Los_Angeles go run .
# Server starts on :8080
```
