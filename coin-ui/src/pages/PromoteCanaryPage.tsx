import { useCallback, useEffect, useState, type ReactNode } from "react";
import { Link, useSearchParams } from "react-router-dom";
import type { PromoteWizardPlan } from "../lib/promoteWizard";
import { buildPromoteChecks, pilotProjects, summarizePlan } from "../lib/promoteWizard";
import { healthBadgeClass } from "../lib/promoteGate";
import { useAuth } from "../context/AuthContext";
import { api, getActor, setActor } from "../lib/api";

export default function PromoteWizard() {
  const { can } = useAuth();
  const [searchParams, setSearchParams] = useSearchParams();
  const [gpNames, setGpNames] = useState<string[]>([]);
  const [gpName, setGpName] = useState(searchParams.get("gp") ?? "");
  const [plan, setPlan] = useState<PromoteWizardPlan | null>(null);
  const [actor, setActorField] = useState(getActor());
  const [loading, setLoading] = useState(true);
  const [executing, setExecuting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    api
      .gpNames()
      .then((r) => {
        setGpNames(r.items);
        if (r.items.length === 0) {
          setGpName("");
          return;
        }
        const fromUrl = searchParams.get("gp");
        setGpName((prev) => {
          if (fromUrl && r.items.includes(fromUrl)) return fromUrl;
          if (prev && r.items.includes(prev)) return prev;
          return r.items[0];
        });
      })
      .catch((err: Error) => setError(err.message));
  }, []);

  const loadPlan = useCallback(async () => {
    if (!gpName) {
      setPlan(null);
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const catalog = await api.catalog(gpName);
      const canaryVersion = catalog.catalog.latestCanary ?? "";
      const currentLatest = catalog.catalog.latest ?? "";

      let composition: PromoteWizardPlan["composition"] = [];
      const canaryComponents: PromoteWizardPlan["canaryComponents"] = [];

      if (canaryVersion) {
        const release = await api.gpRelease(gpName, canaryVersion);
        composition = release.composition;
        await Promise.all(
          composition.map(async (slot) => {
            const versions = await api.componentVersions(slot.type, slot.name);
            const row = versions.items.find((v) => v.version === slot.version);
            if (row?.status === "canary") {
              canaryComponents.push({
                type: slot.type,
                name: slot.name,
                version: slot.version,
              });
            }
          }),
        );
      }

      const health = canaryVersion
        ? await api.health(gpName, canaryVersion, "canary").catch(() => null)
        : null;
      const projects = (await api.projects(gpName)).items;
      const pilots = pilotProjects(projects);

      const { checks, ready, blockers } = buildPromoteChecks({
        canaryVersion,
        currentLatest,
        health,
        pilots,
        canaryComponents,
      });

      setPlan({
        gpName,
        catalog,
        currentLatest,
        canaryVersion,
        composition,
        canaryComponents,
        health,
        pilots,
        checks,
        ready,
        blockers,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "load failed");
      setPlan(null);
    } finally {
      setLoading(false);
    }
  }, [gpName]);

  useEffect(() => {
    void loadPlan();
  }, [loadPlan]);

  function onGpChange(name: string) {
    setGpName(name);
    const next = new URLSearchParams(searchParams);
    if (name) next.set("gp", name);
    else next.delete("gp");
    setSearchParams(next, { replace: true });
  }

  async function executePromote() {
    if (!plan || !plan.ready || !can("publisher")) return;
    setExecuting(true);
    setError(null);
    setMessage(null);
    setActor(actor);
    const actorVal = actor.trim() || undefined;

    try {
      for (const c of plan.canaryComponents) {
        await api.promoteComponentVersion(c.type, c.name, c.version, actorVal);
      }
      if (plan.currentLatest !== plan.canaryVersion) {
        await api.updateCatalog(plan.gpName, {
          latest: plan.canaryVersion,
          latestCanary: plan.canaryVersion,
          minimum: plan.catalog.catalog.minimum,
          deprecated: plan.catalog.catalog.deprecated ?? [],
          actor: actorVal,
        });
      }
      setMessage(`Promote завершён: ${plan.gpName} stable → ${plan.canaryVersion}`);
      await loadPlan();
    } catch (err) {
      setError(err instanceof Error ? err.message : "promote failed");
    } finally {
      setExecuting(false);
    }
  }

  const steps = plan ? summarizePlan(plan) : [];

  return (
    <div className="space-y-6">
      <div>
        <Link to="/catalog" className="text-sm text-sky-400 hover:underline">
          ← GP Policy
        </Link>
        <h1 className="mt-2 text-2xl font-semibold">Promote canary → stable</h1>
        <p className="mt-1 text-zinc-400">
          Единый wizard: components canary→published + catalog latest_canary→latest
        </p>
      </div>

      <div className="max-w-xs">
        <label className="block text-xs text-zinc-500">Golden path</label>
        <select
          value={gpName}
          onChange={(e) => onGpChange(e.target.value)}
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
        <p className="text-zinc-500">Загрузка checklist…</p>
      ) : plan ? (
        <>
          <section className="grid gap-4 sm:grid-cols-3">
            <Stat label="Stable (latest)" value={plan.currentLatest || "—"} mono />
            <Stat label="Canary (latest_canary)" value={plan.canaryVersion || "—"} mono />
            <Stat
              label="Canary health"
              value={plan.health ? plan.health.health : "—"}
              badge={
                plan.health ? (
                  <span
                    className={`rounded px-2 py-0.5 text-xs ${healthBadgeClass(plan.health.health)}`}
                  >
                    {plan.health.failureRate.toFixed(0)}% fail
                  </span>
                ) : undefined
              }
            />
          </section>

          <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-5">
            <h2 className="font-medium mb-4">Checklist</h2>
            <ul className="space-y-3">
              {plan.checks.map((c) => (
                <li key={c.id} className="flex gap-3 text-sm">
                  <span className={c.ok ? "text-emerald-400" : "text-red-400"}>
                    {c.ok ? "✓" : "✗"}
                  </span>
                  <div>
                    <div className="text-zinc-200">{c.label}</div>
                    <div className="text-xs text-zinc-500 font-mono mt-0.5">{c.detail}</div>
                  </div>
                </li>
              ))}
            </ul>
            {!plan.ready && plan.blockers.length > 0 && (
              <ul className="mt-4 text-sm text-amber-200/90 space-y-1">
                {plan.blockers.map((b) => (
                  <li key={b}>• {b}</li>
                ))}
              </ul>
            )}
          </section>

          {steps.length > 0 && (
            <section className="rounded-lg border border-zinc-800 bg-zinc-950 p-5">
              <h2 className="text-sm font-medium text-zinc-400 mb-2">Порядок выполнения</h2>
              <ol className="list-decimal list-inside text-sm text-zinc-300 space-y-1">
                {steps.map((s) => (
                  <li key={s}>{s}</li>
                ))}
              </ol>
            </section>
          )}

          {can("publisher") && (
            <div className="flex flex-wrap items-center gap-3">
              <label className="text-xs text-zinc-500">
                Actor
                <input
                  value={actor}
                  onChange={(e) => setActorField(e.target.value)}
                  className="ml-2 rounded border border-zinc-700 bg-zinc-950 px-3 py-1.5 text-sm"
                />
              </label>
              <button
                type="button"
                disabled={executing || !plan.ready}
                onClick={() => void executePromote()}
                className="rounded bg-emerald-700 px-5 py-2 text-sm hover:bg-emerald-600 disabled:opacity-50"
              >
                {executing ? "Выполнение…" : "Выполнить promote"}
              </button>
              <Link
                to={`/canary`}
                className="text-sm text-sky-400 hover:underline"
              >
                Canary policy →
              </Link>
              <Link
                to={`/studio`}
                className="text-sm text-sky-400 hover:underline"
              >
                Component Studio →
              </Link>
            </div>
          )}
        </>
      ) : null}
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
      <div className={`mt-1 flex items-center gap-2 ${mono ? "font-mono text-sky-400" : ""}`}>
        {value}
        {badge}
      </div>
    </div>
  );
}
