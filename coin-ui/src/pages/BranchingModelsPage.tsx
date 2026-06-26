import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import type { Component, ComponentGPUsage, ComponentVersion } from "../api/types";
import { api } from "../lib/api";
import { platformEditPath } from "../lib/platformComponentPaths";

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

function editPath(row: ModelRow, version: string): string | null {
  return platformEditPath(COMP_TYPE, row.name, version);
}

export default function BranchingModelsPage() {
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
          <p className="text-xs uppercase tracking-wide text-zinc-500">Platform</p>
          <h1 className="text-2xl font-semibold">Branching models</h1>
          <p className="mt-1 text-zinc-400">
            Каталог моделей ветвления · draft → published
          </p>
        </div>
      </div>

      {error && <p className="text-red-400">{error}</p>}

      <div className="overflow-x-auto rounded-lg border border-zinc-800">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-zinc-800 bg-zinc-900/80 text-zinc-500">
            <tr>
              <th className="px-4 py-3 font-medium">Model</th>
              <th className="px-4 py-3 font-medium">Draft</th>
              <th className="px-4 py-3 font-medium">Published</th>
              <th className="px-4 py-3 font-medium">GP profiles</th>
              <th className="px-4 py-3 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-zinc-500">
                  Загрузка…
                </td>
              </tr>
            ) : rows.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-zinc-500">
                  Нет branching models — создайте draft через API или Platform editor
                </td>
              </tr>
            ) : (
              rows.map((row) => {
                const gpNames = gpNamesFor(row.gpUsage);
                const draftVer = row.latestByStatus.draft;
                const draftLink = draftVer ? editPath(row, draftVer) : null;
                return (
                  <tr key={row.name} className="border-b border-zinc-800/60">
                    <td className="px-4 py-3 font-mono">{row.name}</td>
                    <td className="px-4 py-3 font-mono">
                      {draftLink ? (
                        <Link to={draftLink} className="text-amber-400 hover:underline">
                          {draftVer}
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
                      {draftLink && (
                        <Link to={draftLink} className="text-sky-400 hover:underline">
                          Edit
                        </Link>
                      )}
                      <Link
                        to={`/components/${COMP_TYPE}/${row.name}`}
                        className="text-zinc-400 hover:text-zinc-200 hover:underline"
                      >
                        Detail
                      </Link>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      <p className="text-xs text-zinc-500">
        Lifecycle: draft → validate → register → publish. Статусы:{" "}
        <span className={statusPill("draft")}>draft</span>,{" "}
        <span className={statusPill("published")}>published</span>.
      </p>
    </div>
  );
}
