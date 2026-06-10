import { FormEvent, useEffect, useState } from "react";
import type { CatalogOverview } from "../api/types";
import { useAuth } from "../context/AuthContext";
import { api, getActor, setActor } from "../lib/api";

export default function Catalog() {
  const { can } = useAuth();
  const [gpNames, setGpNames] = useState<string[]>([]);
  const [gpName, setGpName] = useState("go-app");
  const [overview, setOverview] = useState<CatalogOverview | null>(null);
  const [latest, setLatest] = useState("");
  const [latestCanary, setLatestCanary] = useState("");
  const [minimum, setMinimum] = useState("");
  const [deprecated, setDeprecated] = useState("");
  const [actor, setActorField] = useState(getActor());
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

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

  useEffect(() => {
    if (!gpName) return;
    setLoading(true);
    setError(null);
    api
      .catalog(gpName)
      .then((o) => {
        setOverview(o);
        setLatest(o.catalog.latest);
        setLatestCanary(o.catalog.latestCanary ?? "");
        setMinimum(o.catalog.minimum);
        setDeprecated(o.catalog.deprecated.join(", "));
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [gpName]);

  async function onSave(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    setMessage(null);
    setActor(actor);
    const dep = deprecated
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
    try {
      await api.updateCatalog(gpName, {
        latest: latest.trim(),
        latestCanary: latestCanary.trim(),
        minimum: minimum.trim(),
        deprecated: dep,
        actor: actor.trim() || undefined,
      });
      setMessage("Catalog policy обновлён");
      const o = await api.catalog(gpName);
      setOverview(o);
    } catch (err) {
      setError(err instanceof Error ? err.message : "save failed");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Catalog & pointers</h1>
        <p className="mt-1 text-zinc-400">
          latest / minimum / deprecated и статус Nexus pointers
        </p>
      </div>

      <div className="max-w-xs">
        <label className="block text-xs text-zinc-500">Golden path</label>
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
      </div>

      {error && <p className="text-red-400">{error}</p>}
      {message && <p className="text-emerald-400">{message}</p>}

      {loading ? (
        <p className="text-zinc-500">Загрузка…</p>
      ) : (
        <>
          {can("publisher") && (
            <form onSubmit={onSave} className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
              <h2 className="font-medium">Catalog policy</h2>
              <div className="mt-4 grid gap-4 sm:grid-cols-2">
                <label className="block">
                  <span className="text-xs text-zinc-500">Latest (stable)</span>
                  <input
                    value={latest}
                    onChange={(e) => setLatest(e.target.value)}
                    className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
                  />
                </label>
                <label className="block">
                  <span className="text-xs text-zinc-500">Latest canary</span>
                  <input
                    value={latestCanary}
                    onChange={(e) => setLatestCanary(e.target.value)}
                    placeholder="1.0.1"
                    className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
                  />
                </label>
                <label className="block">
                  <span className="text-xs text-zinc-500">Minimum</span>
                  <input
                    value={minimum}
                    onChange={(e) => setMinimum(e.target.value)}
                    className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
                  />
                </label>
                <label className="block sm:col-span-2">
                  <span className="text-xs text-zinc-500">Deprecated (comma-separated)</span>
                  <input
                    value={deprecated}
                    onChange={(e) => setDeprecated(e.target.value)}
                    placeholder="1.0.0, 1.0.1"
                    className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
                  />
                </label>
                <label className="block">
                  <span className="text-xs text-zinc-500">Actor</span>
                  <input
                    value={actor}
                    onChange={(e) => setActorField(e.target.value)}
                    className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
                  />
                </label>
              </div>
              <button
                type="submit"
                disabled={saving}
                className="mt-4 rounded bg-sky-600 px-4 py-2 text-sm hover:bg-sky-500 disabled:opacity-50"
              >
                {saving ? "Saving…" : "Save policy"}
              </button>
            </form>
          )}

          <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
            <h2 className="font-medium">Pointer status</h2>
            <p className="mt-1 text-sm text-zinc-500">
              Куда resolve укажет каждый pin при текущем catalog.latest
            </p>
            <table className="mt-4 w-full text-left text-sm">
              <thead className="text-zinc-500">
                <tr>
                  <th className="pb-2 font-medium">Pin</th>
                  <th className="pb-2 font-medium">Resolved version</th>
                  <th className="pb-2 font-medium">Manifest hash</th>
                </tr>
              </thead>
              <tbody>
                {(overview?.pointers ?? []).map((p) => (
                  <tr key={p.pin} className="border-t border-zinc-800/60">
                    <td className="py-2 font-mono text-sky-400">{p.pin}</td>
                    <td className="py-2 font-mono">{p.resolvedVersion || "—"}</td>
                    <td className="py-2 font-mono text-xs text-zinc-500">
                      {p.manifestHash ? p.manifestHash.slice(0, 20) + "…" : "—"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </section>
        </>
      )}
    </div>
  );
}
