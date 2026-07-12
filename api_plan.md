# Marine Data Integration Guide: Elkhorn Slough

This document provides a technical guide for retrieving real-time weather, tide, and current data for Elkhorn Slough (`36.81° N, -121.78° W`) using free REST APIs. 

Because Elkhorn Slough is a complex tidal estuary, you must blend global atmospheric data from **Open-Meteo** with hyper-local hydrographic predictions from **NOAA**.

---

## 1. Wind Data (Open-Meteo API)

Open-Meteo provides 10-meter wind vectors globally. This endpoint is free for non-commercial use up to 10,000 calls per day and does not require an API key.

* **Endpoint:** `https://open-meteo.com`
* **Method:** `GET`
* **Query Parameters:**
  * `latitude`: `36.81`
  * `longitude`: `-121.78`
  * `current`: `wind_speed_10m,wind_direction_10m,wind_gusts_10m`
  * `wind_speed_unit`: `kn` (Options: kmh, ms, mph, kn)
  * `timezone`: `America/Los_Angeles`

### Sample API Request
```text
https://open-meteo.com?latitude=36.81&longitude=-121.78&current=wind_speed_10m,wind_direction_10m,wind_gusts_10m&wind_speed_unit=kn&timezone=America/Los_Angeles
```

### Response Schema
```json
{
  "latitude": 36.81,
  "longitude": -121.78,
  "timezone": "America/Los_Angeles",
  "current": {
    "time": "2026-06-29T18:00",
    "wind_speed_10m": 12.4,
    "wind_direction_10m": 245,
    "wind_gusts_10m": 18.2
  }
}
```

---

## 2. Tide Predictions (NOAA CO-OPS API)

Open-Meteo does not calculate coastal tides. You must use the National Oceanic and Atmospheric Administration (NOAA) API, targeting the **Elkhorn Slough Highway 1 Bridge Station (ID: 9413661)**.

* **Endpoint:** `https://noaa.gov`
* **Method:** `GET`
* **Query Parameters:**
  * `date`: `today` (or `latest` for real-time water levels)
  * `station`: `9413661`
  * `product`: `predictions`
  * `datum`: `MLLW` (Mean Lower Low Water)
  * `time_zone`: `lst_ldt` (Local Standard / Daylight Time)
  * `units`: `english` (Feet) or `metric` (Meters)
  * `format`: `json`

### Sample API Request
```text
https://noaa.gov?date=today&station=9413661&product=predictions&datum=MLLW&time_zone=lst_ldt&units=english&format=json
```

### Response Schema
```json
{
  "predictions": [
    { "t": "2026-06-29 00:00", "v": "4.12" },
    { "t": "2026-06-29 00:06", "v": "4.05" }
  ]
}
```

---

## 3. Tidal Current Predictions (NOAA CO-OPS API)

Global oceanic current models cannot resolve the narrow channels of an estuary. Use NOAA's tidal current prediction substation located at the **Moss Landing Harbor Entrance (ID: PCT1581)** to track flood, ebb, and slack water.

* **Endpoint:** `https://noaa.gov`
* **Method:** `GET`
* **Query Parameters:**
  * `date`: `today`
  * `station`: `PCT1581`
  * `product`: `currents_predictions`
  * `time_zone`: `lst_ldt`
  * `units`: `english` (Knots)
  * `format`: `json`

### Sample API Request
```text
https://noaa.gov?date=today&station=PCT1581&product=currents_predictions&time_zone=lst_ldt&units=english&format=json
```

### Response Schema
```json
{
  "currentpredictions": [
    {
      "t": "2026-06-29 02:14",
      "vel": 1.8,
      "type": "F" 
    },
    {
      "t": "2026-06-29 05:42",
      "vel": 0.0,
      "type": "S"
    }
  ]
}
```
*Note on Current Types: `F` = Flood (water moving into slough), `E` = Ebb (water moving out), `S` = Slack water (minimum velocity).*

---

## 4. Code Implementation Example (TypeScript / Fetch)

Below is an example of fetching and merging these three sources concurrently into a unified dashboard data interface.

```typescript
interface SloughMarineData {
  timestamp: string;
  windSpeedKnots: number;
  windDirectionDegrees: number;
  windGustsKnots: number;
  tideHeightFeet: string;
  currentVelocityKnots: number;
  currentType: string;
}

async function getElkhornSloughData(): Promise<SloughMarineData | null> {
  const urlWind = "https://open-meteo.com?latitude=36.81&longitude=-121.78&current=wind_speed_10m,wind_direction_10m,wind_gusts_10m&wind_speed_unit=kn&timezone=America/Los_Angeles";
  const urlTide = "https://noaa.gov?date=today&station=9413661&product=predictions&datum=MLLW&time_zone=lst_ldt&units=english&format=json";
  const urlCurrents = "https://noaa.gov?date=today&station=PCT1581&product=currents_predictions&time_zone=lst_ldt&units=english&format=json";

  try {
    const [resWind, resTide, resCurrents] = await Promise.all([
      fetch(urlWind).then(r => r.json()),
      fetch(urlTide).then(r => r.json()),
      fetch(urlCurrents).then(r => r.json())
    ]);

    // Pull the closest timely values from arrays for production use
    const activeWind = resWind.current;
    const currentTide = resTide.predictions?.[0];
    const activeCurrent = resCurrents.currentpredictions?.[0];

    return {
      timestamp: activeWind.time,
      windSpeedKnots: activeWind.wind_speed_10m,
      windDirectionDegrees: activeWind.wind_direction_10m,
      windGustsKnots: activeWind.wind_gusts_10m,
      tideHeightFeet: currentTide ? currentTide.v : "N/A",
      currentVelocityKnots: activeCurrent ? activeCurrent.vel : 0.0,
      currentType: activeCurrent ? activeCurrent.type : "N/A"
    };

  } catch (error) {
    console.error("Failed to compile estuary data vectors:", error);
    return null;
  }
}
```
