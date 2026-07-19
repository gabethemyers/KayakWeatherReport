# Kayak Weather Report

Aggregates live wind, tide, current, and swell data into a single at-a-glance view for kayak guides on the Elkhorn Slough in Moss Landing, CA.

**[kayakweather.duckdns.org](https://kayakweather.duckdns.org)**

## The Problem

Guides need wind speed, gusts, tide state, current direction, and swell conditions to decide where and when to launch. The government sources (NWS, NOAA buoys, NOAA tides, Open-Meteo) each expose one piece of the picture through separate interfaces and formats. There's no single view.

## What It Does

Aggregates data from four public APIs into a single view:

| Data | Source |
|---|---|
| Wind forecast | NWS hourly gridpoint (`api.weather.gov`) |
| Live wind speed | Moss Landing NDBC station MLSC1 (`ndbc.noaa.gov`) |
| Tide predictions | NOAA station 9413623 (`tidesandcurrents.noaa.gov`) |
| Swell conditions | Open-Meteo marine API (`marine-api.open-meteo.com`) |

The app infers current direction (flooding/ebbing) and speed from the tide curve. No separate current sensor needed.

## Key Decisions

**Mixed live + forecast wind.** Forecast wind comes from the NWS hourly gridpoint API. When the nearest NDBC buoy has recent data, it overrides the forecast with a live reading. The UI shows a "live" indicator so the user knows which source they're seeing.

**Current inferred from tide.** Instead of pulling from a separate current station (which may not exist nearby), it calculates the current phase and intensity from where we are between the last and next tide peak. No extra API call, no extra failure point.

**Caddy for TLS.** Rather than nginx + certbot + cron, Caddy handles Let's Encrypt automatically in a single container. Cert provisioning, renewal, and HTTP→HTTPS redirect with no moving parts.

## Tech Stack

**Frontend:** React, TypeScript, Vite, Tailwind CSS  
**Backend:** Go, Chi  
**Infrastructure:** Docker Compose, Caddy, DuckDNS

## Setup

### Local development

```bash
docker compose up --build
# App at http://localhost:8080
```

### Production deploy

```bash
docker compose --profile production up -d --build
# App at https://kayakweather.duckdns.org
```

### Architecture

```
Browser ──HTTPS──> Caddy ──HTTP──> Nginx ──/api/──> Go API
                  ports 80/443       serves static frontend
```
