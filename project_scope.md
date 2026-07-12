# Kayak Weather Report - Project Scope

## Overview
A single-page dashboard for Kayak Connection in Moss Landing, CA. Displays real-time conditions for Elkhorn Slough to help staff decide if kayaking is safe **right now**.

## Tech Stack
- **Frontend:** React (Vite) + Tailwind CSS
- **Backend:** Go (REST API)
- **Deployment:** TBD (Vercel for frontend, Railway/Fly.io for backend)

## Data Priorities

### High Priority (The Decision Makers)
1. **Wind Speed** (The Dealbreaker)
   - Green: `< 10 knots`
   - Orange: `10 - 20 knots`
   - Red: `> 20 knots`
2. **Wind Direction**
3. **Current**
   - *Note:* If wind and current oppose → choppy conditions. If aligned → hard to paddle against.

### Details (Supporting Context)
- Tide
- Swell
- Temperature
- Conditions (Sunny/Cloudy/Rain)

## UI States
- **Loading:** Skeleton placeholders for all data points.
- **Success:** Full data displayed with color-coded wind speed.
- **Stale (data > 10 min):** Show yellow indicator next to timestamp.
- **Partial Failure:** Missing fields display `—` (dash) in place of value.
- **Total Failure (API down):** Show friendly error message + Retry button.

## Development Plan

### Phase 1: Backend (Go)
- Build a single endpoint: `GET /api/weather?lat=&lon=`
- Proxy requests to NOAA / OpenWeather APIs.
- Aggregate and return a clean JSON response matching the UI schema.

### Phase 2: Frontend (React + Tailwind)
- Map Figma layout to React components.
- Implement Tailwind utility classes for all spacing/colors.
- Fetch data from Go API using `useEffect`.
- Apply conditional rendering for all UI states.

### Phase 3: Integration & Polish
- Connect frontend to backend.
- Test all states (loading, stale, errors).
- Deploy and share with girlfriend for feedback.

## Data Sources (Potential)
- NOAA Tides & Currents API: https://tidesandcurrents.noaa.gov/web_services_info.html
- OpenWeatherMap (for conditions/temp)
- Stormglass.io (marine weather)
