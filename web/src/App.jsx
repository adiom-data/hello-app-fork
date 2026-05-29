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
          <p className="eyebrow">Go + React fork</p>
          <h1>Hello App Fork</h1>
          <p className="lede">
            A bright little frontend calling a tiny API, ready for experiments.
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
              <dl className="facts">
                <div>
                  <dt>Database</dt>
                  <dd>{hello?.dbEnabled ? "connected" : "not configured"}</dd>
                </div>
                {hello?.hitCount ? (
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
    </main>
  );
}
