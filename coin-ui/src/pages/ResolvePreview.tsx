import { FormEvent, useEffect, useState } from "react";
import type { ResolvePreviewResult } from "../api/types";
import { api } from "../lib/api";

export default function ResolvePreview() {
  const [gpNames, setGpNames] = useState<string[]>([]);
  const [gpName, setGpName] = useState("go-app");
  const [pin, setPin] = useState("~1.0.0");
  const [project, setProject] = useState("");
  const [result, setResult] = useState<ResolvePreviewResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .gpNames()
      .then((r) => {
        setGpNames(r.items);
        if (r.items.length > 0 && !r.items.includes(gpName)) {
          setGpName(r.items[0]);
        }
      })
      .catch((err: Error) => setError(err.message));
  }, []);

  async function onResolve(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      const r = await api.resolvePreview(gpName, pin.trim(), project.trim() || undefined);
      setResult(r);
    } catch (err) {
      setError(err instanceof Error ? err.message : "resolve failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Resolve preview</h1>
        <p className="mt-1 text-zinc-400">Тест pin → manifest JSON (тот же engine, что в CI)</p>
      </div>

      <form onSubmit={onResolve} className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
        <div className="grid gap-4 sm:grid-cols-2">
          <label className="block">
            <span className="text-xs text-zinc-500">Golden path</span>
            <select
              value={gpName}
              onChange={(e) => setGpName(e.target.value)}
              className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
            >
              {gpNames.map((n) => (
                <option key={n} value={n}>
                  {n}
                </option>
              ))}
            </select>
          </label>
          <label className="block">
            <span className="text-xs text-zinc-500">Pin</span>
            <input
              value={pin}
              onChange={(e) => setPin(e.target.value)}
              placeholder="=1.0.0, ~1.0.0, ^1.0.0, *"
              className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
              required
            />
          </label>
          <label className="block sm:col-span-2">
            <span className="text-xs text-zinc-500">Project (optional, для canary в MVP-3)</span>
            <input
              value={project}
              onChange={(e) => setProject(e.target.value)}
              placeholder="my-service"
              className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
            />
          </label>
        </div>
        <button
          type="submit"
          disabled={loading}
          className="mt-4 rounded bg-sky-600 px-4 py-2 text-sm hover:bg-sky-500 disabled:opacity-50"
        >
          {loading ? "Resolving…" : "Resolve"}
        </button>
      </form>

      {error && <p className="text-red-400">{error}</p>}

      {result && (
        <section className="space-y-4">
          <div className="flex flex-wrap gap-6 text-sm">
            <div>
              <span className="text-zinc-500">Requested pin: </span>
              <span className="font-mono text-sky-400">{result.requestedPin}</span>
            </div>
            <div>
              <span className="text-zinc-500">Resolved: </span>
              <span className="font-mono text-emerald-400">{result.resolvedVersion}</span>
            </div>
            <div>
              <span className="text-zinc-500">Channel: </span>
              <span
                className={`font-mono ${result.channel === "canary" ? "text-amber-400" : "text-emerald-400"}`}
              >
                {result.channel}
              </span>
            </div>
          </div>
          <pre className="overflow-x-auto rounded-lg border border-zinc-800 bg-zinc-950 p-4 text-xs font-mono text-zinc-300">
            {JSON.stringify(result.manifest, null, 2)}
          </pre>
        </section>
      )}
    </div>
  );
}
