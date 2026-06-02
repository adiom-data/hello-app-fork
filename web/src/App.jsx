import { useEffect, useRef, useState } from "react";
import { Calculator, Delete, Music2, RefreshCcw, Server, VolumeX } from "lucide-react";

const noteToFrequency = {
  C3: 130.81,
  D3: 146.83,
  E3: 164.81,
  F3: 174.61,
  G3: 196,
  A3: 220,
  B3: 246.94,
  C4: 261.63,
  D4: 293.66,
  E4: 329.63,
  F4: 349.23,
  G4: 392,
  A4: 440,
  B4: 493.88,
  C5: 523.25,
};

const melody = [
  "E4",
  "G4",
  "A4",
  "C5",
  "B4",
  "G4",
  "E4",
  "D4",
  "C4",
  "E4",
  "G4",
  "B4",
  "A4",
  "F4",
  "D4",
  "C4",
];

const bassline = ["A3", "A3", "G3", "G3", "F3", "F3", "E3", "C3"];

const webRingName = "dummy MCP context ok";
const marqueeMessage = "dummy MCP context ok for repo adiom-data/hello-app-fork in org 4218776c-057b-44f8-aed6-238c199e8903";

function formatHitCount(count) {
  if (!Number.isFinite(count)) {
    return "??????";
  }

  return count.toString().padStart(6, "0");
}

function calculate(firstValue, secondValue, operator) {
  switch (operator) {
    case "+":
      return firstValue + secondValue;
    case "-":
      return firstValue - secondValue;
    case "*":
      return firstValue * secondValue;
    case "/":
      return secondValue === 0 ? null : firstValue / secondValue;
    default:
      return secondValue;
  }
}

function formatCalculatorValue(value) {
  if (!Number.isFinite(value)) {
    return "Error";
  }

  return Number.parseFloat(value.toFixed(10)).toString().slice(0, 14);
}

export default function App() {
  const [hello, setHello] = useState(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [musicPlaying, setMusicPlaying] = useState(false);
  const [calculatorValue, setCalculatorValue] = useState("0");
  const [storedCalculatorValue, setStoredCalculatorValue] = useState(null);
  const [calculatorOperator, setCalculatorOperator] = useState(null);
  const [waitingForCalculatorValue, setWaitingForCalculatorValue] = useState(false);
  const audioRef = useRef(null);
  const hitCounterValue = formatHitCount(hello?.hitCount);

  async function loadHello() {
    setLoading(true);
    setError("");

    try {
      const response = await fetch("/api/hello");
      if (!response.ok) {
        throw new Error(`Request failed with ${response.status}`);
      }

      setHello(await response.json());
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong");
      setHello(null);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadHello();
  }, []);

  useEffect(() => {
    return () => {
      if (audioRef.current?.timer) {
        window.clearInterval(audioRef.current.timer);
      }
      audioRef.current?.context.close();
      audioRef.current = null;
    };
  }, []);

  function playTone(context, destination, frequency, startAt, duration, type = "square", volume = 0.12) {
    const oscillator = context.createOscillator();
    const gain = context.createGain();

    oscillator.type = type;
    oscillator.frequency.setValueAtTime(frequency, startAt);
    gain.gain.setValueAtTime(0.0001, startAt);
    gain.gain.exponentialRampToValueAtTime(volume, startAt + 0.015);
    gain.gain.exponentialRampToValueAtTime(0.0001, startAt + duration);

    oscillator.connect(gain);
    gain.connect(destination);
    oscillator.start(startAt);
    oscillator.stop(startAt + duration + 0.02);
  }

  function scheduleSceneLoop(engine, startAt) {
    const step = 0.18;

    melody.forEach((note, index) => {
      playTone(engine.context, engine.master, noteToFrequency[note], startAt + index * step, step * 0.82);
      if (index % 4 === 2) {
        playTone(
          engine.context,
          engine.master,
          noteToFrequency[note] * 2,
          startAt + index * step + step * 0.5,
          step * 0.35,
          "triangle",
          0.045,
        );
      }
    });

    bassline.forEach((note, index) => {
      playTone(
        engine.context,
        engine.master,
        noteToFrequency[note] / 2,
        startAt + index * step * 2,
        step * 1.5,
        "sawtooth",
        0.055,
      );
    });
  }

  function stopSceneMusic() {
    if (audioRef.current?.timer) {
      window.clearInterval(audioRef.current.timer);
    }
    audioRef.current?.context.close();
    audioRef.current = null;
    setMusicPlaying(false);
  }

  async function toggleSceneMusic() {
    if (audioRef.current) {
      stopSceneMusic();
      return;
    }

    const AudioContext = window.AudioContext || window.webkitAudioContext;
    if (!AudioContext) {
      return;
    }

    const context = new AudioContext();
    const master = context.createGain();
    master.gain.value = 0.22;
    master.connect(context.destination);

    const engine = { context, master, timer: null };
    audioRef.current = engine;
    await context.resume();

    scheduleSceneLoop(engine, context.currentTime + 0.04);
    engine.timer = window.setInterval(() => {
      scheduleSceneLoop(engine, context.currentTime + 0.04);
    }, 2880);

    setMusicPlaying(true);
  }

  function enterCalculatorDigit(digit) {
    if (calculatorValue === "Error" || waitingForCalculatorValue) {
      setCalculatorValue(digit);
      setWaitingForCalculatorValue(false);
      return;
    }

    setCalculatorValue((currentValue) =>
      currentValue === "0" ? digit : `${currentValue}${digit}`.slice(0, 14),
    );
  }

  function enterCalculatorDecimal() {
    if (calculatorValue === "Error" || waitingForCalculatorValue) {
      setCalculatorValue("0.");
      setWaitingForCalculatorValue(false);
      return;
    }

    if (!calculatorValue.includes(".")) {
      setCalculatorValue((currentValue) => `${currentValue}.`);
    }
  }

  function clearCalculator() {
    setCalculatorValue("0");
    setStoredCalculatorValue(null);
    setCalculatorOperator(null);
    setWaitingForCalculatorValue(false);
  }

  function backspaceCalculator() {
    if (waitingForCalculatorValue || calculatorValue === "Error") {
      setCalculatorValue("0");
      setWaitingForCalculatorValue(false);
      return;
    }

    setCalculatorValue((currentValue) =>
      currentValue.length > 1 ? currentValue.slice(0, -1) : "0",
    );
  }

  function toggleCalculatorSign() {
    if (calculatorValue === "0" || calculatorValue === "Error") {
      return;
    }

    setCalculatorValue((currentValue) =>
      currentValue.startsWith("-") ? currentValue.slice(1) : `-${currentValue}`,
    );
  }

  function applyCalculatorPercent() {
    if (calculatorValue === "Error") {
      return;
    }

    setCalculatorValue(formatCalculatorValue(Number(calculatorValue) / 100));
  }

  function chooseCalculatorOperator(nextOperator) {
    if (calculatorValue === "Error") {
      clearCalculator();
      return;
    }

    const inputValue = Number(calculatorValue);

    if (calculatorOperator && storedCalculatorValue !== null && !waitingForCalculatorValue) {
      const result = calculate(storedCalculatorValue, inputValue, calculatorOperator);

      if (result === null) {
        setCalculatorValue("Error");
        setStoredCalculatorValue(null);
        setCalculatorOperator(null);
        setWaitingForCalculatorValue(true);
        return;
      }

      setCalculatorValue(formatCalculatorValue(result));
      setStoredCalculatorValue(result);
    } else {
      setStoredCalculatorValue(inputValue);
    }

    setCalculatorOperator(nextOperator);
    setWaitingForCalculatorValue(true);
  }

  function solveCalculator() {
    if (!calculatorOperator || storedCalculatorValue === null || calculatorValue === "Error") {
      return;
    }

    const result = calculate(storedCalculatorValue, Number(calculatorValue), calculatorOperator);

    if (result === null) {
      setCalculatorValue("Error");
    } else {
      setCalculatorValue(formatCalculatorValue(result));
    }

    setStoredCalculatorValue(null);
    setCalculatorOperator(null);
    setWaitingForCalculatorValue(true);
  }

  return (
    <main className="app-shell">
      <section className="hero">
        <div className="hero-copy">
          <p className="eyebrow">Go + React fork</p>
          <h1>Hello App Fork</h1>
          <p className="lede">
            A bright little frontend calling a tiny API, ready for experiments.
          </p>
          <button
            type="button"
            className="music-toggle"
            aria-pressed={musicPlaying}
            onClick={toggleSceneMusic}
          >
            {musicPlaying ? <VolumeX aria-hidden="true" /> : <Music2 aria-hidden="true" />}
            <span>{musicPlaying ? "Mute scene music" : "Play scene music"}</span>
          </button>
        </div>
      </section>

      <section className="workspace" aria-label="API response">
        <div className="response-card">
          <div className="response-header">
            <div className="response-title">
              <Server aria-hidden="true" />
              <span>API Response</span>
            </div>
            <button type="button" onClick={loadHello} disabled={loading}>
              <RefreshCcw aria-hidden="true" />
              <span>{loading ? "Loading" : "Refresh"}</span>
            </button>
          </div>

          {error ? (
            <p className="error">{error}</p>
          ) : (
            <div className="response-body">
              <p className="message">
                {loading ? "Waiting for the server..." : hello?.message}
              </p>
              {hello?.time ? (
                <time dateTime={hello.time}>
                  {new Date(hello.time).toLocaleString()}
                </time>
              ) : null}
              <dl className="facts">
                <div>
                  <dt>Database</dt>
                  <dd>{hello?.dbEnabled ? "connected" : "not configured"}</dd>
                </div>
                {hello?.hitCount != null ? (
                  <div>
                    <dt>Stored hits</dt>
                    <dd>{hello.hitCount}</dd>
                  </div>
                ) : null}
                {hello?.lastHitAt ? (
                  <div>
                    <dt>Last DB write</dt>
                    <dd>{new Date(hello.lastHitAt).toLocaleString()}</dd>
                  </div>
                ) : null}
              </dl>
              {hello?.dbError ? <p className="error inline">{hello.dbError}</p> : null}
            </div>
          )}
        </div>
      </section>

      <footer className="web-ring" aria-label={`${webRingName} web ring`}>
        <div className="web-ring-panel">
          <div className="web-ring-banner">{webRingName} web ring</div>
          <div className="awesome-marquee" aria-label={marqueeMessage}>
            <div className="awesome-marquee-track">
              <span>{marqueeMessage}</span>
              <span>{marqueeMessage}</span>
              <span>{marqueeMessage}</span>
              <span>{marqueeMessage}</span>
            </div>
          </div>
          <p className="web-ring-title">This site is a proud member of the {webRingName} ring.</p>
          <nav className="web-ring-links" aria-label="Web ring navigation">
            <a href="#previous-site">Prev</a>
            <a href="#random-site">Random</a>
            <a href="#next-site">Next</a>
          </nav>
          <div className="hit-counter" aria-label={`Visitor counter: ${hitCounterValue}`}>
            <span className="hit-counter-label">You are visitor number</span>
            <span className="hit-counter-digits" aria-hidden="true">
              {hitCounterValue.split("").map((character, index) => (
                <span className="hit-counter-digit" key={`${character}-${index}`}>
                  {character}
                </span>
              ))}
            </span>
          </div>
          <div className="web-ring-badges" aria-hidden="true">
            <span className="badge badge-hot">HOT!</span>
            <span className="badge badge-html">HTML 4.0</span>
            <span className="badge badge-counter">Netscape Now</span>
          </div>
        </div>
      </footer>

      <section className="calculator-section" aria-labelledby="calculator-title">
        <div className="calculator-panel">
          <div className="calculator-header">
            <div className="calculator-title" id="calculator-title">
              <Calculator aria-hidden="true" />
              <span>Calculator</span>
            </div>
            <div className="calculator-status" aria-hidden="true">
              {calculatorOperator ? `${calculatorOperator} ready` : "online"}
            </div>
          </div>
          <output className="calculator-display" aria-live="polite">
            {calculatorValue}
          </output>
          <div className="calculator-keys">
            <button type="button" className="calculator-key utility" onClick={clearCalculator}>
              AC
            </button>
            <button type="button" className="calculator-key utility" onClick={toggleCalculatorSign}>
              +/-
            </button>
            <button type="button" className="calculator-key utility" onClick={applyCalculatorPercent}>
              %
            </button>
            <button
              type="button"
              className="calculator-key operator"
              aria-label="Divide"
              onClick={() => chooseCalculatorOperator("/")}
            >
              /
            </button>

            <button type="button" className="calculator-key" onClick={() => enterCalculatorDigit("7")}>
              7
            </button>
            <button type="button" className="calculator-key" onClick={() => enterCalculatorDigit("8")}>
              8
            </button>
            <button type="button" className="calculator-key" onClick={() => enterCalculatorDigit("9")}>
              9
            </button>
            <button
              type="button"
              className="calculator-key operator"
              aria-label="Multiply"
              onClick={() => chooseCalculatorOperator("*")}
            >
              x
            </button>

            <button type="button" className="calculator-key" onClick={() => enterCalculatorDigit("4")}>
              4
            </button>
            <button type="button" className="calculator-key" onClick={() => enterCalculatorDigit("5")}>
              5
            </button>
            <button type="button" className="calculator-key" onClick={() => enterCalculatorDigit("6")}>
              6
            </button>
            <button
              type="button"
              className="calculator-key operator"
              aria-label="Subtract"
              onClick={() => chooseCalculatorOperator("-")}
            >
              -
            </button>

            <button type="button" className="calculator-key" onClick={() => enterCalculatorDigit("1")}>
              1
            </button>
            <button type="button" className="calculator-key" onClick={() => enterCalculatorDigit("2")}>
              2
            </button>
            <button type="button" className="calculator-key" onClick={() => enterCalculatorDigit("3")}>
              3
            </button>
            <button
              type="button"
              className="calculator-key operator"
              aria-label="Add"
              onClick={() => chooseCalculatorOperator("+")}
            >
              +
            </button>

            <button
              type="button"
              className="calculator-key utility"
              aria-label="Backspace"
              onClick={backspaceCalculator}
            >
              <Delete aria-hidden="true" />
            </button>
            <button type="button" className="calculator-key" onClick={() => enterCalculatorDigit("0")}>
              0
            </button>
            <button type="button" className="calculator-key" onClick={enterCalculatorDecimal}>
              .
            </button>
            <button type="button" className="calculator-key equals" onClick={solveCalculator}>
              =
            </button>
          </div>
        </div>
      </section>
    </main>
  );
}
