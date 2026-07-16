import CriticalItem from './components/CriticalItem'

import {useState, useEffect} from 'react';
import { formatRelative, parse } from "date-fns";

export interface TidePeak {
  time: string;
  height: number;
  type: 'H' | 'L'; // Using a union type here ensures you only expect "H" or "L"
}

export interface WeatherData {
  wind: {
    gust: number;
    speed: number;
    direction: string;
    arrow: string;
    is_live: boolean;
  };
  current: {
    state: 'Flooding' | 'Ebbing' | 'Slack'; // "Flooding", "Ebbing", or "Slack"
    speed: 'Fast' | 'Slow' | 'Not Moving'; // "Fast", "Slow", or "Not Moving"
    arrow: string; 
  };
  tide: {
    next_type: 'H' | 'L';
    next_time: string;
    next_height: number;
  };
  swell: {
    wave_height: number;
    wave_period: number;
    direction: string; // cardinal direction
    arrow: string;
  };
}


// Define a type for the return value for better TypeScript support
export interface FormattedTide {
  day: string;
  time: string;
}

function formatTideTime(dateTimeStr: string): FormattedTide {
  const date = parse(dateTimeStr, "yyyy-MM-dd HH:mm", new Date());
  
  // This gets "today at 6:12 AM" or "tomorrow at 6:12 AM"
  const relative = formatRelative(date, new Date());
  
  // Capitalize the first letter
  const capitalizedRelative = relative.charAt(0).toUpperCase() + relative.slice(1);

  // Split by the word " at " to separate the day from the time
  // Example: "Today at 6:12 AM" becomes ["Today", "6:12 AM"]
  const parts = capitalizedRelative.split(' at ');

  return {
    day: parts[0], // "Today"
    time: parts[1] || '', // "6:12 AM" (fallback to empty string if split fails)
  };
}

function App() {
  const today = new Date()
  const [weather, setWeather] = useState<WeatherData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetch("/api/v1/conditions")
      .then(response => response.json())
      .then(data => {
        setWeather(data);
        setLoading(false);
      })
      .catch(err => {
        setError(err.message);
        setLoading(false);
      });
  }, []);

  if (loading) {
    return <div className="text-center p-10">Loading weather...</div>;
  }

  if (error) {
    return <div className="text-center p-10 text-red-500">Error: {error}</div>;
  }


  return (
    <div className="flex flex-col max-w-sm mx-auto min-h-screen bg-gradient-to-br from-sky-200 to-blue-400 rounded-xl gap-3">
      {/* HEADER BAR */}
      <div className="flex justify-between items-center pt-2 px-2">
        <p className="text-md font-bold text-slate-800">📍Moss Landing</p>
        <p className="text-base font-medium text-slate-600">{today.toLocaleDateString([], {month: 'long', day: 'numeric', year: 'numeric'})}</p>
      </div>

      {/* Title */}
      <div className="text-center py-2">
        <h1 className="text-sky-900 text-3xl font-extrabold tracking-tight">
          Carolyn's Kayaking <br/>
          <span className="text-sky-900">Weather Report</span>
        </h1>
      </div>

      {/* HERO */}
      <div className="flex flex-col justify-center items-stretch mx-6 my-2 pb-2 rounded-2xl bg-white overflow-hidden shadow-md">
        <p className="text-[12px] font-bold uppercase tracking-widest text-center text-sky-900 bg-sky-300 py-0.5">Wind</p>
        <div className="grid grid-cols-[1fr_auto_1fr] items-center">
          <div className="flex flex-col justify-center items-center ">
             <div className="flex items-center gap-1.5 text-lg text-gray-700 h-7">
              {weather.wind.is_live && (
                <span className="text-[10px] font-bold tracking-wider text-emerald-800 bg-emerald-100 border border-emerald-300 px-1 rounded">
                  LIVE
                </span>
              )}
            </div>
            <p className="text-[64px] leading-none font-bold">{weather.wind.speed}<span className="text-base text-gray-500 font-normal"> mph</span></p>
          </div>
          <div className="flex flex-col justify-center items-center">
            <p className="text-5xl">{weather.wind.arrow}</p>
            <p className="text-lg">{weather.wind.direction}</p>
          </div>
          <div className="flex flex-col justify-center items-center px-4">
            <p className="text-lg text-gray-700">Gust</p>
            <p className="text-[64px] leading-none font-bold whitespace-nowrap">{weather.wind.gust}<span className="text-base text-gray-500 font-normal"> mph</span></p>
          </div>
        </div>
      </div>

      {/* Current */}
      <div className="flex flex-col justify-center items-stretch mx-6 my-2 rounded-2xl bg-white overflow-hidden shadow-md">
        <p className="text-[12px] font-bold uppercase tracking-widest text-center text-sky-900 bg-sky-300 py-0.5">Current</p>
        <div className="grid grid-cols-3 items-center bg-white px-8 pt-2 pb-4 gap-1">
          <div className="flex justify-center items-center">
            <p className="text-[24px] leading-none font-medium text-center">{weather.current.speed}</p>
          </div>
          <div className="flex justify-center items-center gap-1">
            <p className="text-5xl">{weather.current.arrow}</p>
          </div>
          <div className="flex justify-center items-center">
            <p className="text-[24px] leading-none font-medium text-center">{weather.current.state}</p>
          </div>
        </div>
      </div>

      {/* Swell */}
      <div className="flex flex-col justify-center items-stretch mx-6 my-2 rounded-2xl bg-white overflow-hidden shadow-md">
        <p className="text-[12px] font-bold uppercase tracking-widest text-center text-sky-900 bg-sky-300 py-0.5">Swell</p>
        <div className="grid grid-cols-3 items-center bg-white px-8 pb-2">
          {/* Left Column: Wave Height */}
          <div className="flex justify-start items-center">
            <p className="text-2xl p-2 whitespace-nowrap font-medium">{weather.swell.wave_height.toFixed(2)} <span className="text-base text-gray-500 font-normal">ft</span></p>
          </div>
          {/* Center Column: Direction */}
          <div className="flex flex-col justify-center items-center">
            <p className="text-4xl">{weather.swell.arrow}</p>
            <p className="text-lg">{weather.swell.direction}</p>
          </div>
          {/* Right Column: Wave Period */}
          <div className="flex justify-end items-center">
            <p className="text-2xl p-2 whitespace-nowrap">{weather.swell.wave_period} <span className="text-base text-gray-500 font-normal">sec</span></p>
          </div>
        </div>
      </div>

      {/* Tide */}
      <div className="flex flex-col justify-center items-stretch mx-6 my-2 rounded-2xl bg-white overflow-hidden shadow-md">
        <p className="text-[12px] font-bold uppercase tracking-widest text-center text-sky-900 bg-sky-300 py-0.5">Tide</p>

        <div className="grid grid-cols-3 items-center bg-white px-8 pb-2">
          {/* Left Column: Formatted Next Tide Time */}
          <div className="flex flex-col items-start">
            <p className="text-sm text-gray-500">{formatTideTime(weather.tide.next_time).day} at</p>
            <p className="text-lg">{formatTideTime(weather.tide.next_time).time}</p>
          </div>
          {/* Center Column: Next Tide Type (H or L) */}
          <p className="text-3xl text-bold text-center">{weather.tide.next_type}</p>
          {/* Right Column: Next Tide Height */}
          <p className="text-2xl p-2 whitespace-nowrap">{weather.tide.next_height.toFixed(2)} <span className="text-base text-gray-500 font-normal">ft</span></p>
        </div>
      </div>

    </div>

  );
}

export default App;
