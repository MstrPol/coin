import { FormEvent, useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import type { CatalogOverview } from "../../../api/types";
import { useAuth } from "../../../context/AuthContext";
import { api, getActor, setActor } from "../../../lib/api";

function lineBadge(line?: string) {
  if (line === "stable") return <span className="ml-1 text-emerald-400">★ stable</span>;
  if (line === "canary") return <span className="ml-1 text-amber-400">★ canary</span>;
  return null;
}

export default function GpPolicyTab() {
  const { name: gpName = "" } = useParams();
  const { can } = useAuth();
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
    if (!gpName) return;
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
      setOverview(await api.catalog(gpName));
    } catch (err) {
      setError(err instanceof Error ? err.message : "save failed");
    } finally {
      setSaving(false);
    }
  }

  if (loading) return <p className="text-zinc-500">Загрузка…</p>;

  return (
    <div className="space-y-6">
      {error && <p className="text-red-400">{error}</p>}
      {message && <p className="text-emerald-400">{message}</p>}

      {can("publisher") && overview?.catalog.latestCanary && (
        <div className="rounded-lg border border-sky-900/40 bg-sky-950/20 px-4 py-3 text-sm text-zinc-300">
          Canary line: <span className="font-mono text-sky-400">{overview.catalog.latestCanary}</span>
        </div>
      )}

      {can("publisher") && (
        <form onSubmit={onSave} className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
          <h2 className="font-medium">Политика версий</h2>
          <div className="mt-4 grid gap-4 sm:grid-cols-2">
            <PolicySelect label="Latest (stable)" value={latest} onChange={setLatest} options={publishedVersions} />
            <PolicySelect
              label="Latest canary"
              value={latestCanary}
              onChange={setLatestCanary}
              options={publishedVersions}
              emptyLabel="— не задано —"
            />
            <PolicySelect label="Minimum" value={minimum} onChange={setMinimum} options={publishedVersions} />
            <label className="block sm:col-span-2">
              <span className="text-xs text-zinc-500">Deprecated (comma-separated)</span>
              <input
                value={deprecated}
                onChange={(e) => setDeprecated(e.target.value)}
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
        <table className="mt-4 w-full text-left text-sm">
          <thead className="text-zinc-500">
            <tr>
              <th className="pb-2 font-medium">Pin</th>
              <th className="pb-2 font-medium">Аудитория</th>
              <th className="pb-2 font-medium">Resolved</th>
            </tr>
          </thead>
          <tbody>
            {(overview?.pointers ?? []).map((p) => (
              <tr key={`${p.pin}-${p.audience ?? "all"}`} className="border-t border-zinc-800/60">
                <td className="py-2 font-mono text-sky-400">{p.pin}</td>
                <td className="py-2 text-zinc-400">{p.audience ?? "—"}</td>
                <td className="py-2 font-mono">
                  {p.resolvedVersion || "—"}
                  {lineBadge(p.line)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>

      <p className="text-sm text-zinc-500">
        Canary rollout — вкладка{" "}
        <Link to={`/gp/${encodeURIComponent(gpName)}/canary`} className="text-sky-400 hover:underline">
          Canary
        </Link>
        .
      </p>
    </div>
  );
}

function PolicySelect({
  label,
  value,
  onChange,
  options,
  emptyLabel = "—",
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  options: string[];
  emptyLabel?: string;
}) {
  return (
    <label className="block">
      <span className="text-xs text-zinc-500">{label}</span>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
      >
        <option value="">{emptyLabel}</option>
        {options.map((v) => (
          <option key={v} value={v}>
            {v}
          </option>
        ))}
      </select>
    </label>
  );
}
