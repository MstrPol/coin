import { FormEvent, useEffect, useState, type ReactNode } from "react";
import { Link } from "react-router-dom";
import type { CanaryOverview, HealthSummary } from "../api/types";
import { useAuth } from "../context/AuthContext";
import { api, getActor, setActor } from "../lib/api";

function healthBadge(health: HealthSummary["health"]) {
  const styles = {
    healthy: "bg-emerald-950/50 text-emerald-400",
    degraded: "bg-amber-950/50 text-amber-400",
    critical: "bg-red-950/50 text-red-400",
  };
  return (
    <span className={`rounded px-2 py-0.5 text-xs font-medium ${styles[health]}`}>{health}</span>
  );
}

export default function Canary() {
  const { can } = useAuth();
  const [gpNames, setGpNames] = useState<string[]>([]);
  const [gpName, setGpName] = useState("go-app");
  const [overview, setOverview] = useState<CanaryOverview | null>(null);
  const [health, setHealth] = useState<HealthSummary | null>(null);
  const [enabled, setEnabled] = useState(false);
  const [percent, setPercent] = useState(10);
  const [degradedPct, setDegradedPct] = useState(10);
  const [criticalPct, setCriticalPct] = useState(25);
  const [previewProject, setPreviewProject] = useState("demo-go-app");
  const [previewResult, setPreviewResult] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    api
      .gpNames()
      .then((r) => {
        setGpNames(r.items);
        if (r.items.length > 0 && !r.items.includes(gpName)) setGpName(r.items[0]);
      })
      .catch((err: Error) => setError(err.message));
  }, []);

  useEffect(() => {
    if (!gpName) return;
    setLoading(true);
    setError(null);
    api
      .canary(gpName)
      .then(async (o) => {
        setOverview(o);
        setEnabled(o.policy.enabled);
        setPercent(o.policy.canaryPercent);
        setDegradedPct(o.policy.degradedThresholdPct);
        setCriticalPct(o.policy.criticalThresholdPct);
        if (o.catalog.latestCanary) {
          const h = await api.health(gpName, o.catalog.latestCanary, "canary").catch(() => null);
          setHealth(h);
        } else {
          setHealth(null);
        }
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [gpName]);

  async function onSave(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    setMessage(null);
    setActor(getActor());
    try {
      await api.updateCanary(gpName, {
        enabled,
        canaryPercent: percent,
        degradedThresholdPct: degradedPct,
        criticalThresholdPct: criticalPct,
        criticalConsecutiveFailures: 3,
        actor: getActor() || undefined,
      });
      setMessage("Canary policy сохранён");
      const o = await api.canary(gpName);
      setOverview(o);
    } catch (err) {
      setError(err instanceof Error ? err.message : "save failed");
    } finally {
      setSaving(false);
    }
  }

  async function previewResolve() {
    setPreviewResult(null);
    try {
      const r = await api.resolvePreview(gpName, "*", previewProject.trim());
      setPreviewResult(`${r.channel} → ${r.resolvedVersion}`);
    } catch (err) {
      setPreviewResult(err instanceof Error ? err.message : "preview failed");
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Canary rollout</h1>
        <p className="mt-1 text-zinc-400">
          Pin <span className="font-mono">*</span> + project → stable или canary line
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
          {health?.health === "critical" && (
            <div className="rounded border border-red-900/50 bg-red-950/30 px-4 py-3 text-red-300">
              Canary line critical — рассмотрите откат или pause canary (
              {overview?.catalog.latestCanary})
            </div>
          )}

          <section className="grid gap-4 sm:grid-cols-3">
            <Stat label="Projects in canary" value={`${overview?.inCanary ?? 0} / ${overview?.totalProjects ?? 0}`} />
            <Stat label="Latest canary" value={overview?.catalog.latestCanary || "—"} mono />
            <Stat
              label="Canary health"
              value={health ? `${health.failureRate.toFixed(1)}% fail` : "—"}
              badge={health ? healthBadge(health.health) : undefined}
            />
          </section>

          {can("publisher") && (
            <form onSubmit={onSave} className="rounded-lg border border-zinc-800 bg-zinc-900 p-6 space-y-6">
              <h2 className="font-medium">Rollout policy</h2>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={enabled}
                  onChange={(e) => setEnabled(e.target.checked)}
                />
                Canary enabled
              </label>
              <div>
                <label className="text-xs text-zinc-500">
                  Percent rollout ({percent}%) — default projects
                </label>
                <input
                  type="range"
                  min={0}
                  max={100}
                  value={percent}
                  onChange={(e) => setPercent(Number(e.target.value))}
                  className="mt-2 w-full"
                />
                <p className="mt-1 text-xs text-zinc-500">
                  Preview: ~{overview?.inCanary ?? 0} из {overview?.totalProjects ?? 0} projects
                  получат canary
                </p>
              </div>
              <div className="grid gap-4 sm:grid-cols-2">
                <label className="block text-xs text-zinc-500">
                  Degraded threshold (%)
                  <input
                    type="number"
                    value={degradedPct}
                    onChange={(e) => setDegradedPct(Number(e.target.value))}
                    className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
                  />
                </label>
                <label className="block text-xs text-zinc-500">
                  Critical threshold (%)
                  <input
                    type="number"
                    value={criticalPct}
                    onChange={(e) => setCriticalPct(Number(e.target.value))}
                    className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
                  />
                </label>
              </div>
              <button
                type="submit"
                disabled={saving}
                className="rounded bg-sky-600 px-4 py-2 text-sm hover:bg-sky-500 disabled:opacity-50"
              >
                {saving ? "Saving…" : "Save policy"}
              </button>
            </form>
          )}

          <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
            <h2 className="font-medium">Resolve preview</h2>
            <div className="mt-4 flex flex-wrap gap-3">
              <input
                value={previewProject}
                onChange={(e) => setPreviewProject(e.target.value)}
                placeholder="project name"
                className="rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
              />
              <button
                type="button"
                onClick={previewResolve}
                className="rounded bg-zinc-800 px-4 py-2 text-sm hover:bg-zinc-700"
              >
                pin=* preview
              </button>
              {previewResult && <span className="self-center font-mono text-sm text-sky-400">{previewResult}</span>}
            </div>
          </section>

          {health && health.recentFailures.length > 0 && (
            <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
              <h2 className="font-medium">Recent failures (24h)</h2>
              <table className="mt-4 w-full text-left text-sm">
                <thead className="text-zinc-500">
                  <tr>
                    <th className="pb-2">Project</th>
                    <th className="pb-2">Stage</th>
                    <th className="pb-2">When</th>
                  </tr>
                </thead>
                <tbody>
                  {health.recentFailures.map((f, i) => (
                    <tr key={i} className="border-t border-zinc-800/60">
                      <td className="py-2 font-mono">{f.project}</td>
                      <td className="py-2">{f.failedStage ?? "—"}</td>
                      <td className="py-2 text-zinc-400">
                        {new Date(f.reportedAt).toLocaleString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </section>
          )}

          <p className="text-sm text-zinc-500">
            Allowlist / stable override — на странице{" "}
            <Link to="/projects" className="text-sky-400 hover:underline">
              Projects
            </Link>
            . Latest canary — в{" "}
            <Link to="/catalog" className="text-sky-400 hover:underline">
              Catalog
            </Link>
            .
          </p>
        </>
      )}
    </div>
  );
}

function Stat({
  label,
  value,
  mono,
  badge,
}: {
  label: string;
  value: string;
  mono?: boolean;
  badge?: ReactNode;
}) {
  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900 p-4">
      <div className="text-xs text-zinc-500">{label}</div>
      <div className={`mt-1 flex items-center gap-2 text-lg ${mono ? "font-mono" : ""}`}>
        {value}
        {badge}
      </div>
    </div>
  );
}
