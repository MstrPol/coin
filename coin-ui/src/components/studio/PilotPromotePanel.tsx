import { useCallback, useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import type { ComponentGPUsage, Project } from "../../api/types";
import { api, getActor } from "../../lib/api";
import {
  GpHealthRow,
  healthBadgeClass,
  pilotProjects,
  promoteGate,
} from "../../lib/promoteGate";

type Props = {
  type: string;
  name: string;
  version: string;
  canEdit: boolean;
  onPromoted: () => void;
};

export default function PilotPromotePanel({ type, name, version, canEdit, onPromoted }: Props) {
  const [gpUsage, setGpUsage] = useState<ComponentGPUsage[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [healthRows, setHealthRows] = useState<GpHealthRow[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(true);
  const [assigning, setAssigning] = useState(false);
  const [promoting, setPromoting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const pilots = pilotProjects(projects);
  const promoteGpName = gpUsage[0]?.gpName ?? healthRows[0]?.gpName ?? "";

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const detail = await api.componentDetail(type, name);
      const usage = detail.gpUsage.filter(
        (u) => u.status === "canary" || u.status === "draft",
      );
      setGpUsage(usage);

      const gpNames = [...new Set(usage.map((u) => u.gpName))];
      const projectLists = await Promise.all(gpNames.map((gp) => api.projects(gp)));
      const merged = new Map<string, Project>();
      for (const list of projectLists) {
        for (const p of list.items) {
          merged.set(p.name, p);
        }
      }
      setProjects([...merged.values()].sort((a, b) => a.name.localeCompare(b.name)));

      const health: GpHealthRow[] = [];
      for (const entry of usage) {
        const h = await api.health(entry.gpName, entry.version, "canary").catch(() => null);
        health.push({ gpName: entry.gpName, gpVersion: entry.version, health: h });
      }
      if (health.length === 0 && gpNames.length > 0) {
        for (const gp of gpNames) {
          const catalog = await api.catalog(gp);
          const canaryVer = catalog.catalog.latestCanary;
          if (!canaryVer) continue;
          const h = await api.health(gp, canaryVer, "canary").catch(() => null);
          health.push({ gpName: gp, gpVersion: canaryVer, health: h });
        }
      }
      setHealthRows(health);
    } catch (err) {
      setError(err instanceof Error ? err.message : "load failed");
    } finally {
      setLoading(false);
    }
  }, [type, name]);

  useEffect(() => {
    void load();
  }, [load]);

  const gate = useMemo(
    () => promoteGate({ healthRows, projects, selectedPilotNames: selected }),
    [healthRows, projects, selected],
  );

  function togglePilot(projectName: string) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(projectName)) next.delete(projectName);
      else next.add(projectName);
      return next;
    });
  }

  async function assignPilots() {
    if (!canEdit || selected.size === 0) return;
    setAssigning(true);
    setError(null);
    setMessage(null);
    try {
      const actor = getActor() || undefined;
      for (const projectName of selected) {
        await api.updateProjectCanaryMode(projectName, "canary", actor);
      }
      setMessage(`Pilot mode включён для: ${[...selected].join(", ")}`);
      setSelected(new Set());
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "assign failed");
    } finally {
      setAssigning(false);
    }
  }

  async function promoteStable() {
    if (!canEdit || !gate.ok) return;
    setPromoting(true);
    setError(null);
    setMessage(null);
    try {
      await api.promoteComponentVersion(type, name, version, getActor() || undefined);
      setMessage(`Promoted → published: ${type}/${name}@${version}`);
      onPromoted();
    } catch (err) {
      setError(err instanceof Error ? err.message : "promote failed");
    } finally {
      setPromoting(false);
    }
  }

  if (loading) {
    return (
      <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-5">
        <p className="text-sm text-zinc-500">Загрузка pilot / health…</p>
      </section>
    );
  }

  return (
    <section className="rounded-lg border border-sky-900/40 bg-zinc-900 p-5 space-y-5">
      <div>
        <h2 className="font-medium">Pilot projects &amp; promote stable</h2>
        <p className="mt-1 text-xs text-zinc-500">
          Назначьте pilot projects с <span className="font-mono">canary_mode=canary</span>, дождитесь
          green health, затем promote component → published.
        </p>
      </div>

      {error && <p className="text-red-400 text-sm">{error}</p>}
      {message && <p className="text-emerald-400 text-sm">{message}</p>}

      {gpUsage.length > 0 ? (
        <div className="text-sm">
          <h3 className="text-xs font-medium text-zinc-500 mb-2">GP composition (canary/draft)</h3>
          <ul className="space-y-1 font-mono text-xs text-zinc-300">
            {gpUsage.map((u) => (
              <li key={`${u.gpName}/${u.version}`}>
                <Link to={`/releases/${u.gpName}/${u.version}`} className="text-sky-400 hover:underline">
                  {u.gpName}@{u.version}
                </Link>{" "}
                <span className="text-zinc-500">({u.status})</span>
              </li>
            ))}
          </ul>
        </div>
      ) : (
        <p className="text-sm text-amber-300">
          Компонент не pin&apos;нут в canary/draft GP release. Добавьте в composition через{" "}
          <Link to="/releases/publish" className="text-sky-400 hover:underline">
            Publish wizard
          </Link>
          .
        </p>
      )}

      {healthRows.length > 0 && (
        <div className="grid gap-3 sm:grid-cols-2">
          {healthRows.map((row) => (
            <HealthCard key={`${row.gpName}/${row.gpVersion}`} row={row} />
          ))}
        </div>
      )}

      <div>
        <h3 className="text-sm font-medium mb-2">
          Pilot projects{" "}
          <span className="text-zinc-500 font-normal">
            ({pilots.length} active{pilots.length === 1 ? "" : "s"})
          </span>
        </h3>
        {projects.length === 0 ? (
          <p className="text-sm text-zinc-500">Нет зарегистрированных projects на связанных GP.</p>
        ) : (
          <div className="max-h-48 overflow-y-auto rounded border border-zinc-800">
            <table className="w-full text-left text-sm">
              <thead className="sticky top-0 bg-zinc-900 text-xs text-zinc-500">
                <tr>
                  <th className="px-3 py-2 w-8" />
                  <th className="px-3 py-2">Project</th>
                  <th className="px-3 py-2">GP</th>
                  <th className="px-3 py-2">canary_mode</th>
                </tr>
              </thead>
              <tbody>
                {projects.map((p) => {
                  const isPilot = p.canaryMode === "canary";
                  return (
                    <tr key={p.name} className="border-t border-zinc-800/60">
                      <td className="px-3 py-2">
                        {canEdit && !isPilot && (
                          <input
                            type="checkbox"
                            checked={selected.has(p.name)}
                            onChange={() => togglePilot(p.name)}
                          />
                        )}
                        {isPilot && <span className="text-emerald-400 text-xs">✓</span>}
                      </td>
                      <td className="px-3 py-2 font-mono">{p.name}</td>
                      <td className="px-3 py-2 text-zinc-400">{p.goldenPath}</td>
                      <td className="px-3 py-2">
                        <span className={isPilot ? "text-sky-400" : "text-zinc-500"}>
                          {p.canaryMode ?? "default"}
                        </span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
        {canEdit && selected.size > 0 && (
          <button
            type="button"
            onClick={() => void assignPilots()}
            disabled={assigning}
            className="mt-3 rounded border border-zinc-600 px-3 py-1.5 text-sm hover:bg-zinc-800 disabled:opacity-50"
          >
            {assigning ? "Назначение…" : `Назначить pilots (${selected.size})`}
          </button>
        )}
      </div>

      {!gate.ok && (
        <ul className="text-sm text-amber-200/90 space-y-1">
          {gate.reasons.map((r) => (
            <li key={r}>• {r}</li>
          ))}
        </ul>
      )}

      {canEdit && (
        <div className="flex flex-wrap gap-2">
          {promoteGpName && (
            <Link
              to={`/promote?gp=${encodeURIComponent(promoteGpName)}`}
              className="rounded border border-sky-700 px-4 py-2 text-sm text-sky-300 hover:bg-sky-950/40"
            >
              Полный promote wizard (catalog + components) →
            </Link>
          )}
          <button
          type="button"
          onClick={() => void promoteStable()}
          disabled={promoting || !gate.ok}
          className="rounded bg-emerald-700 px-4 py-2 text-sm hover:bg-emerald-600 disabled:opacity-50"
        >
          {promoting ? "Promote…" : "Promote component only"}
        </button>
        </div>
      )}
    </section>
  );
}

function HealthCard({ row }: { row: GpHealthRow }) {
  const h = row.health;
  return (
    <div className="rounded border border-zinc-800 bg-zinc-950 p-3 text-sm">
      <div className="font-mono text-xs text-zinc-400">
        {row.gpName}@{row.gpVersion}
      </div>
      {h ? (
        <div className="mt-2 flex flex-wrap items-center gap-2">
          <span className={`rounded px-2 py-0.5 text-xs font-medium ${healthBadgeClass(h.health)}`}>
            {h.health}
          </span>
          <span className="text-zinc-400 text-xs">
            {h.successCount} ok / {h.failureCount} fail ({h.failureRate.toFixed(1)}%)
          </span>
        </div>
      ) : (
        <p className="mt-2 text-xs text-zinc-500">Нет build reports за 24h</p>
      )}
    </div>
  );
}
