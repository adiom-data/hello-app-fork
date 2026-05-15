import { useEffect, useState } from "react";
import { RefreshCcw, Server } from "lucide-react";

export default function App() {
  const [hello, setHello] = useState(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

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

  return (
    <main className="app-shell">
      <section className="hero">
        <div className="hero-copy">
          <p className="eyebrow">Go + React sample</p>
          <h1>Hello App</h1>
          <p className="lede">
            A tiny frontend calling a tiny API, ready for experiments.
          </p>
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
            </div>
          )}
        </div>
      </section>
    </main>
  );
}
