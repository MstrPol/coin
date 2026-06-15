import { FormEvent, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import type { CatalogOverview } from "../api/types";
import { useAuth } from "../context/AuthContext";
import { api, getActor, setActor } from "../lib/api";

function lineBadge(line?: string) {
  if (line === "stable") {
    return <span className="ml-1 text-emerald-400">★ stable</span>;
  }
  if (line === "canary") {
    return <span className="ml-1 text-amber-400">★ canary</span>;
  }
  return null;
}

export default function Catalog() {
  const { can } = useAuth();
  const [gpNames, setGpNames] = useState<string[]>([]);
  const [gpName, setGpName] = useState("");
  const [publishedVersions, setPublishedVersions] = useState<string[]>([]);
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
        if (r.items.length === 0) {
          setGpName("");
        } else {
          setGpName((prev) => (prev && r.items.includes(prev) ? prev : r.items[0]));
        }
      })
      .catch((err: Error) => setError(err.message));
  }, []);

  useEffect(() => {
    if (!gpName) {
      setLoading(false);
      setOverview(null);
      setPublishedVersions([]);
      setLatest("");
      setLatestCanary("");
      setMinimum("");
      setDeprecated("");
      return;
    }
    setLoading(true);
    setError(null);
    Promise.all([api.catalog(gpName), api.gpReleases(gpName, false)])
      .then(([o, releases]) => {
        const published = releases.items
          .filter((r) => r.status === "published" && !r.version.includes("-snapshot."))
          .map((r) => r.version)
          .sort();
        setPublishedVersions(published);
        setOverview(o);
        setLatest(o.catalog.latest ?? "");
        setLatestCanary(o.catalog.latestCanary ?? "");
        setMinimum(o.catalog.minimum ?? "");
        setDeprecated((o.catalog.deprecated ?? []).join(", "));
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
      setMessage("Политика версий обновлена");
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
        <h1 className="text-2xl font-semibold">Политика версий GP</h1>
        <p className="mt-1 text-zinc-400">
          Какие версии стабильны, устарели и сняты с эксплуатации
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

      {gpNames.length === 0 && (
        <p className="text-zinc-500">
          Нет golden paths.{" "}
          <Link to="/releases/new-gp" className="text-sky-400 hover:underline">
            Создайте GP profile
          </Link>
          , затем опубликуйте release.
        </p>
      )}

      {loading ? (
        <p className="text-zinc-500">Загрузка…</p>
      ) : gpName ? (
        <>
          {can("publisher") && (
            <form onSubmit={onSave} className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
              <h2 className="font-medium">Политика версий</h2>
              <div className="mt-4 grid gap-4 sm:grid-cols-2">
                <label className="block">
                  <span className="text-xs text-zinc-500">Latest (stable)</span>
                  <p className="mt-0.5 text-xs text-zinc-600">
                    Pin * резолвится сюда. Только published без snapshot.
                  </p>
                  <select
                    value={latest}
                    onChange={(e) => setLatest(e.target.value)}
                    className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
                  >
                    <option value="">—</option>
                    {publishedVersions.map((v) => (
                      <option key={v} value={v}>
                        {v}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="block">
                  <span className="text-xs text-zinc-500">Latest canary</span>
                  <p className="mt-0.5 text-xs text-zinc-600">
                    Canary-линия для pin * при rollout и canary mode проекта.
                  </p>
                  <select
                    value={latestCanary}
                    onChange={(e) => setLatestCanary(e.target.value)}
                    className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
                  >
                    <option value="">— не задано —</option>
                    {publishedVersions.map((v) => (
                      <option key={v} value={v}>
                        {v}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="block">
                  <span className="text-xs text-zinc-500">Minimum</span>
                  <p className="mt-0.5 text-xs text-zinc-600">
                    Ниже — validate fail, пайп останавливается до test/build.
                  </p>
                  <select
                    value={minimum}
                    onChange={(e) => setMinimum(e.target.value)}
                    className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
                  >
                    <option value="">—</option>
                    {publishedVersions.map((v) => (
                      <option key={v} value={v}>
                        {v}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="block sm:col-span-2">
                  <span className="text-xs text-zinc-500">Deprecated (comma-separated)</span>
                  <p className="mt-0.5 text-xs text-zinc-600">
                    Билд разрешён, warning в validate и resolve headers.
                  </p>
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
            <h2 className="font-medium">Превью resolve</h2>
            <p className="mt-1 text-sm text-zinc-500">
              Pin * — две строки (stable / canary). Pin =, ~, ^ — stable-линия для всех проектов.
            </p>
            <table className="mt-4 w-full text-left text-sm">
              <thead className="text-zinc-500">
                <tr>
                  <th className="pb-2 font-medium">Pin</th>
                  <th className="pb-2 font-medium">Аудитория</th>
                  <th className="pb-2 font-medium">Resolved version</th>
                  <th className="pb-2 font-medium">Manifest hash</th>
                </tr>
              </thead>
              <tbody>
                {(overview?.pointers ?? []).map((p) => (
                  <tr
                    key={`${p.pin}-${p.audience ?? "all"}`}
                    className="border-t border-zinc-800/60"
                  >
                    <td className="py-2 font-mono text-sky-400">{p.pin}</td>
                    <td className="py-2 text-zinc-400">{p.audience ?? "—"}</td>
                    <td className="py-2 font-mono">
                      {p.resolvedVersion || "—"}
                      {lineBadge(p.line)}
                    </td>
                    <td className="py-2 font-mono text-xs text-zinc-500">
                      {p.manifestHash ? p.manifestHash.slice(0, 20) + "…" : "—"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </section>
        </>
      ) : null}
    </div>
  );
}
