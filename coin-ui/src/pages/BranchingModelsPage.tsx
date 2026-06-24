import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import type { Component, ComponentGPUsage, ComponentVersion } from "../api/types";
import { useAuth } from "../context/AuthContext";
import { api } from "../lib/api";

const COMP_TYPE = "branching-model";

type ModelRow = {
  name: string;
  versions: ComponentVersion[];
  gpUsage: ComponentGPUsage[];
  latestByStatus: Record<string, string>;
};

function latestByStatus(versions: ComponentVersion[]): Record<string, string> {
  const out: Record<string, string> = {};
  for (const v of versions) {
    if (!out[v.status]) {
      out[v.status] = v.version;
    }
  }
  return out;
}

function statusPill(status: string): string {
  switch (status) {
    case "draft":
      return "text-amber-400";
    case "canary":
      return "text-sky-400";
    case "published":
      return "text-emerald-400";
    default:
      return "text-zinc-400";
  }
}

function studioTarget(row: ModelRow): string | null {
  return (
    row.latestByStatus.draft ??
    row.latestByStatus.canary ??
    row.latestByStatus.published ??
    row.versions[0]?.version ??
    null
  );
}

export default function BranchingModelsPage() {
  const { can } = useAuth();
  const [rows, setRows] = useState<ModelRow[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoading(true);
      setError(null);
      try {
        const all = await api.components();
        const models = all.items.filter((c: Component) => c.type === COMP_TYPE);
        const enriched = await Promise.all(
          models.map(async (c) => {
            const [vers, detail] = await Promise.all([
              api.componentVersions(COMP_TYPE, c.name),
              api.componentDetail(COMP_TYPE, c.name),
            ]);
            return {
              name: c.name,
              versions: vers.items,
              gpUsage: detail.gpUsage,
              latestByStatus: latestByStatus(vers.items),
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
        if (!cancelled) {
          setLoading(false);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const gpNamesFor = (usage: ComponentGPUsage[]) =>
    [...new Set(usage.map((u) => u.gpName))].sort();

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">Branching Models</h1>
          <p className="mt-1 text-zinc-400">
            Каталог моделей ветвления · canary в PostgreSQL, Nexus на promote
          </p>
        </div>
        {can("publisher") && (
          <Link
            to="/studio"
            className="rounded bg-sky-600 px-4 py-2 text-sm font-medium hover:bg-sky-500"
          >
            Component Studio
          </Link>
        )}
      </div>

      {error && <p className="text-red-400">{error}</p>}

      <div className="overflow-x-auto rounded-lg border border-zinc-800">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-zinc-800 bg-zinc-900/80 text-zinc-500">
            <tr>
              <th className="px-4 py-3 font-medium">Model</th>
              <th className="px-4 py-3 font-medium">Draft</th>
              <th className="px-4 py-3 font-medium">Canary</th>
              <th className="px-4 py-3 font-medium">Published</th>
              <th className="px-4 py-3 font-medium">GP profiles</th>
              <th className="px-4 py-3 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-zinc-500">
                  Загрузка…
                </td>
              </tr>
            ) : rows.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-zinc-500">
                  Нет branching models — создайте в{" "}
                  <Link to="/studio" className="text-sky-400 hover:underline">
                    Studio
                  </Link>
                </td>
              </tr>
            ) : (
              rows.map((row) => {
                const target = studioTarget(row);
                const gpNames = gpNamesFor(row.gpUsage);
                return (
                  <tr key={row.name} className="border-b border-zinc-800/60">
                    <td className="px-4 py-3 font-mono">{row.name}</td>
                    <td className="px-4 py-3 font-mono">
                      {row.latestByStatus.draft ? (
                        <Link
                          to={`/studio/${COMP_TYPE}/${row.name}/${encodeURIComponent(row.latestByStatus.draft)}`}
                          className="text-amber-400 hover:underline"
                        >
                          {row.latestByStatus.draft}
                        </Link>
                      ) : (
                        "—"
                      )}
                    </td>
                    <td className="px-4 py-3 font-mono">
                      {row.latestByStatus.canary ? (
                        <Link
                          to={`/studio/${COMP_TYPE}/${row.name}/${encodeURIComponent(row.latestByStatus.canary)}`}
                          className="text-sky-400 hover:underline"
                        >
                          {row.latestByStatus.canary}
                        </Link>
                      ) : (
                        "—"
                      )}
                    </td>
                    <td className="px-4 py-3 font-mono text-emerald-400">
                      {row.latestByStatus.published ?? "—"}
                    </td>
                    <td className="px-4 py-3 text-zinc-400">
                      {gpNames.length > 0 ? (
                        <span className="font-mono text-xs">{gpNames.join(", ")}</span>
                      ) : (
                        "—"
                      )}
                    </td>
                    <td className="px-4 py-3 space-x-3">
                      {target && (
                        <Link
                          to={`/studio/${COMP_TYPE}/${row.name}/${encodeURIComponent(target)}`}
                          className="text-sky-400 hover:underline"
                        >
                          Studio
                        </Link>
                      )}
                      <Link
                        to={`/components/${COMP_TYPE}/${row.name}`}
                        className="text-zinc-400 hover:text-zinc-200 hover:underline"
                      >
                        Detail
                      </Link>
                      {row.latestByStatus.canary && can("publisher") && (
                        <Link
                          to={`/studio/${COMP_TYPE}/${row.name}/${encodeURIComponent(row.latestByStatus.canary)}`}
                          className="text-emerald-400 hover:underline"
                        >
                          Promote
                        </Link>
                      )}
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      <p className="text-xs text-zinc-500">
        Lifecycle: draft → validate → register (PG) → canary → promote (Nexus). Статусы:{" "}
        <span className={statusPill("draft")}>draft</span>,{" "}
        <span className={statusPill("canary")}>canary</span>,{" "}
        <span className={statusPill("published")}>published</span>.
      </p>
    </div>
  );
}
