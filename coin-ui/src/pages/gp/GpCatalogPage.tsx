import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../../lib/api";
import { useAuth } from "../../context/AuthContext";

type GpRow = {
  name: string;
  slotCount: number;
  latestStable: string;
  latestCanary: string;
  releaseCount: number;
  draftCount: number;
};

export default function GpCatalogPage() {
  const { can } = useAuth();
  const [rows, setRows] = useState<GpRow[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoading(true);
      setError(null);
      try {
        const names = await api.gpNames();
        const enriched = await Promise.all(
          names.items.map(async (name) => {
            const [profile, catalog, releases] = await Promise.all([
              api.gpProfile(name).catch(() => null),
              api.catalog(name).catch(() => null),
              api.gpReleases(name, true).catch(() => ({ items: [] })),
            ]);
            const published = releases.items.filter((r) => r.status === "published");
            const drafts = releases.items.filter((r) => r.status === "draft");
            return {
              name,
              slotCount: profile?.slots.length ?? 0,
              latestStable: catalog?.catalog.latest ?? "—",
              latestCanary: catalog?.catalog.latestCanary ?? "—",
              releaseCount: published.length,
              draftCount: drafts.length,
            };
          }),
        );
        if (!cancelled) {
          setRows(enriched.sort((a, b) => a.name.localeCompare(b.name)));
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "load failed");
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-wide text-zinc-500">Golden Paths</p>
          <h1 className="text-2xl font-semibold">GP Profiles</h1>
          <p className="mt-1 text-zinc-400">Каталог Golden Path profiles — не список версий</p>
        </div>
        {can("publisher") && (
          <Link
            to="/gp/new"
            className="rounded-lg border border-sky-500/70 bg-sky-950/50 px-4 py-2 text-sm font-semibold text-sky-300 hover:border-sky-400"
          >
            + New profile
          </Link>
        )}
      </div>

      {error && <p className="text-red-400">{error}</p>}
      {loading ? (
        <p className="text-zinc-500">Загрузка…</p>
      ) : (
        <div className="overflow-x-auto rounded-lg border border-zinc-800">
          <table className="w-full text-left text-sm">
            <thead className="border-b border-zinc-800 bg-zinc-900/80 text-zinc-500">
              <tr>
                <th className="px-4 py-3 font-medium">Profile</th>
                <th className="px-4 py-3 font-medium">Slots</th>
                <th className="px-4 py-3 font-medium">Latest stable</th>
                <th className="px-4 py-3 font-medium">Latest canary</th>
                <th className="px-4 py-3 font-medium">Releases</th>
                <th className="px-4 py-3 font-medium" />
              </tr>
            </thead>
            <tbody>
              {rows.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-zinc-500">
                    Нет GP profiles
                    {can("publisher") && (
                      <>
                        {" "}
                        —{" "}
                        <Link to="/gp/new" className="text-sky-400 hover:underline">
                          создать первый
                        </Link>
                      </>
                    )}
                  </td>
                </tr>
              ) : (
                rows.map((row) => (
                  <tr key={row.name} className="border-b border-zinc-800/60">
                    <td className="px-4 py-3 font-mono font-medium">{row.name}</td>
                    <td className="px-4 py-3 tabular-nums">{row.slotCount}</td>
                    <td className="px-4 py-3 font-mono text-emerald-400">{row.latestStable}</td>
                    <td className="px-4 py-3 font-mono text-amber-400">{row.latestCanary}</td>
                    <td className="px-4 py-3">
                      {row.releaseCount}
                      {row.draftCount > 0 && (
                        <span className="ml-2 text-xs text-amber-400">+{row.draftCount} draft</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <Link
                        to={`/gp/${encodeURIComponent(row.name)}`}
                        className="text-sky-400 hover:underline"
                      >
                        Open →
                      </Link>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
