import { useEffect, useState } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import type { CatalogOverview, GPProfile, GPProfileSlot } from "../../../api/types";
import { api } from "../../../lib/api";
import { SLOT_LABELS } from "../../../lib/gpSlots";

export default function GpOverviewTab() {
  const { name = "" } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const showWelcome = searchParams.get("welcome") === "1";
  const [profile, setProfile] = useState<GPProfile | null>(null);
  const [catalog, setCatalog] = useState<CatalogOverview | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!name) return;
    setError(null);
    Promise.all([api.gpProfile(name), api.catalog(name)])
      .then(([p, c]) => {
        setProfile(p);
        setCatalog(c);
      })
      .catch((err: Error) => setError(err.message));
  }, [name]);

  function dismissWelcome() {
    const next = new URLSearchParams(searchParams);
    next.delete("welcome");
    setSearchParams(next, { replace: true });
  }

  const base = `/gp/${encodeURIComponent(name)}`;

  return (
    <div className="space-y-6">
      {showWelcome && (
        <div className="rounded-lg border border-emerald-900/50 bg-emerald-950/30 px-4 py-4 flex flex-wrap items-center justify-between gap-3">
          <p className="text-sm text-emerald-200">
            Profile <span className="font-mono">{name}</span> создан. Опубликуйте initial release,
            когда composition готов.
          </p>
          <div className="flex gap-2">
            <Link
              to={`${base}/releases/new`}
              className="rounded bg-sky-600 px-3 py-1.5 text-sm hover:bg-sky-500"
            >
              Publish initial release
            </Link>
            <button
              type="button"
              onClick={dismissWelcome}
              className="text-sm text-zinc-400 hover:text-zinc-200"
            >
              Dismiss
            </button>
          </div>
        </div>
      )}

      {error && <p className="text-red-400">{error}</p>}

      <section className="grid gap-4 sm:grid-cols-3">
        <Stat label="Latest stable" value={catalog?.catalog.latest || "—"} mono />
        <Stat label="Latest canary" value={catalog?.catalog.latestCanary || "—"} mono />
        <Stat label="Minimum" value={catalog?.catalog.minimum || "—"} mono />
      </section>

      <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
        <h2 className="font-medium">Composition slots</h2>
        <table className="mt-4 w-full text-left text-sm">
          <thead className="text-zinc-500">
            <tr>
              <th className="pb-2 font-medium">Slot</th>
              <th className="pb-2 font-medium">Type</th>
              <th className="pb-2 font-medium">Component</th>
            </tr>
          </thead>
          <tbody>
            {(profile?.slots ?? []).map((s: GPProfileSlot) => (
              <tr key={s.key} className="border-t border-zinc-800/60">
                <td className="py-2 font-mono text-sky-400">{s.key}</td>
                <td className="py-2">{SLOT_LABELS[s.key] ?? s.type}</td>
                <td className="py-2 font-mono text-xs">
                  {s.type}/{s.name}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>

      <section className="flex flex-wrap gap-4 text-sm">
        <Link to={`${base}/releases`} className="text-sky-400 hover:underline">
          Releases →
        </Link>
        <Link to={`${base}/policy`} className="text-sky-400 hover:underline">
          Version policy →
        </Link>
        <Link to={`${base}/canary`} className="text-sky-400 hover:underline">
          Canary rollout →
        </Link>
        <Link to={`${base}/build-stack`} className="text-sky-400 hover:underline">
          Build stack →
        </Link>
        <Link to={`/resolve?gp=${encodeURIComponent(name)}`} className="text-zinc-400 hover:underline">
          Resolve preview ↗
        </Link>
      </section>
    </div>
  );
}

function Stat({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900 p-4">
      <div className="text-xs text-zinc-500">{label}</div>
      <div className={`mt-1 text-lg ${mono ? "font-mono" : ""}`}>{value}</div>
    </div>
  );
}
