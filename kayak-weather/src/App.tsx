import CriticalItem from './components/CriticalItem'

import {useState, useEffect} from 'react';

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
    arrow: string;                          // "↑", "↓", or "↔"
  };
  tide: {
    next_type: 'High' | 'Low';
    next_time: string;
    next_height: number;
  };
  swell: {
    wave_height: number;
    wave_period: number;
    direction: number; // Heading in degrees (e.g., 270)
  };
}

function App() {
  const today = new Date()
  const [weather, setWeather] = useState<WeatherData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetch("http://100.94.171.1:8080/api/weather")
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
    <div className="flex flex-col max-w-sm mx-auto min-h-screen bg-sky-300 rounded-xl">
      {/* HEADER BAR */}
      <div className="flex justify-between items-center p-4">
        <p>Elkhorn Slough</p>
        <p>Date: {today.toLocaleDateString()}</p>
      </div>

      {/* HERO */}
      <div className="flex flex-col justify-center items-stretch mx-6 my-2 rounded-2xl bg-white overflow-hidden">
        <p className="text-sm text-uppercase text-center text-grey-500 bg-sky-500">Wind</p>
        <div className="flex justify-evenly items-center bg-white">
          <div className="flex flex-col justify-center items-center px-4 pb-2">
            <p className="text-lg text-gray-700">Gust</p>
            <p className="text-[64px] leading-none font-bold">{weather.wind.gust}<span className="text-base text-gray-500 font-normal"> mph</span></p>
          </div>
          <div className="flex flex-col justify-center items-center">
            <p className="text-5xl">{weather.wind.arrow}</p>
            <p className="text-lg">{weather.wind.direction}</p>
          </div>
          <div className="flex flex-col justify-center items-center px-4 pb-2">
             <div className="flex items-center gap-1.5 text-lg text-gray-700 h-7">
              {weather.wind.is_live && (
                <span className="text-[10px] font-bold tracking-wider text-emerald-800 bg-emerald-100 border border-emerald-300 px-1 rounded">
                  LIVE
                </span>
              )}
              <span>Speed</span>
            </div>
            <p className="text-[64px] leading-none font-bold">{weather.wind.speed}<span className="text-base text-gray-500 font-normal"> mph</span></p>
          </div>
        </div>
      </div>


      <div className="flex flex-col justify-center items-stretch mx-6 my-2 rounded-2xl bg-white overflow-hidden">
        <p className="text-sm text-uppercase text-center text-grey-500 bg-sky-500">Current</p>
        <div className="flex justify-between items-center bg-white px-8 pb-2">
          <div className="flex justify-center items-center gap-1">
            <p className="text-5xl">{weather.current.arrow}</p>
            <p className="text-4xl">{weather.current.direction}</p>
          </div>
          <div>
          </div>
          <div className="flex flex-col justify-center items-center pb-2">
            <p className="text-[64px] leading-none font-bold">{weather.current.speed}<span className="text-base text-gray-500 font-normal"> kts</span></p>
          </div>
        </div>
      </div>

      <div className="flex justify-around p-4 gap-2">
        <div className="flex flex-col justify-center items-stretch my-2 rounded-2xl bg-white overflow-hidden">
          <p className="text-sm text-uppercase text-center text-grey-500 bg-sky-500">Tide</p>
          <div className="flex justify-center items-center bg-white px-8 pb-2">
            <p className="text-2xl p-2">{weather.tide} ft</p>
          </div>
        </div>

        <div className="flex flex-col justify-center items-stretch my-2 rounded-2xl bg-white overflow-hidden">
          <p className="text-sm text-uppercase text-center text-grey-500 bg-sky-500">Swell</p>
          <div className="flex justify-center items-center bg-white px-8 pb-2">
            <p className="text-2xl p-2">{weather.swell} ft</p>
          </div>
        </div>
      </div>
    </div>

  );
}

export default App;
