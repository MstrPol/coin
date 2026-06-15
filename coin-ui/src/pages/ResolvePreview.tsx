import { FormEvent, useEffect, useMemo, useState } from "react";
import type { CanaryContext, Project, ResolvePreviewResult } from "../api/types";
import { api } from "../lib/api";

type ForceChannel = "" | "canary" | "stable";

export default function ResolvePreview() {
  const [gpNames, setGpNames] = useState<string[]>([]);
  const [gpName, setGpName] = useState("go-app");
  const [pin, setPin] = useState("~1.0.0");
  const [project, setProject] = useState("");
  const [projectQuery, setProjectQuery] = useState("");
  const [projects, setProjects] = useState<Project[]>([]);
  const [showProjectList, setShowProjectList] = useState(false);
  const [canaryCtx, setCanaryCtx] = useState<CanaryContext | null>(null);
  const [forceChannel, setForceChannel] = useState<ForceChannel>("");
  const [result, setResult] = useState<ResolvePreviewResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
    api
      .projects(gpName || undefined)
      .then((r) => setProjects(r.items))
      .catch(() => setProjects([]));
  }, [gpName]);

  useEffect(() => {
    if (!project.trim() || !gpName) {
      setCanaryCtx(null);
      return;
    }
    api
      .canaryContext(gpName, project.trim())
      .then(setCanaryCtx)
      .catch(() => setCanaryCtx(null));
  }, [project, gpName]);

  const filteredProjects = useMemo(() => {
    const q = projectQuery.trim().toLowerCase();
    if (!q) return projects.slice(0, 20);
    return projects.filter((p) => p.name.toLowerCase().includes(q)).slice(0, 20);
  }, [projects, projectQuery]);

  async function onResolve(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      const r = await api.resolvePreview(
        gpName,
        pin.trim(),
        project.trim() || undefined,
        forceChannel || undefined,
      );
      setResult(r);
      if (r.canaryContext) setCanaryCtx(r.canaryContext);
    } catch (err) {
      setError(err instanceof Error ? err.message : "resolve failed");
    } finally {
      setLoading(false);
    }
  }

  function selectProject(name: string) {
    setProject(name);
    setProjectQuery(name);
    setShowProjectList(false);
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Resolve preview</h1>
        <p className="mt-1 text-zinc-400">Тест pin → manifest JSON (тот же engine, что в CI)</p>
      </div>

      <form onSubmit={onResolve} className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
        <div className="grid gap-4 sm:grid-cols-2">
          <label className="block">
            <span className="text-xs text-zinc-500">Golden path</span>
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
          </label>
          <label className="block">
            <span className="text-xs text-zinc-500">Pin</span>
            <input
              value={pin}
              onChange={(e) => setPin(e.target.value)}
              placeholder="=1.0.0, ~1.0.0, ^1.0.0, *"
              className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
              required
            />
          </label>
          <label className="relative block sm:col-span-2">
            <span className="text-xs text-zinc-500">Project (для canary при pin *)</span>
            <input
              value={projectQuery}
              onChange={(e) => {
                setProjectQuery(e.target.value);
                setProject(e.target.value);
                setShowProjectList(true);
              }}
              onFocus={() => setShowProjectList(true)}
              onBlur={() => setTimeout(() => setShowProjectList(false), 150)}
              placeholder="— без проекта — или выберите из списка"
              className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
            />
            {showProjectList && filteredProjects.length > 0 && (
              <ul className="absolute z-10 mt-1 max-h-48 w-full overflow-auto rounded border border-zinc-700 bg-zinc-950 py-1 text-sm shadow-lg">
                <li>
                  <button
                    type="button"
                    className="w-full px-3 py-2 text-left text-zinc-500 hover:bg-zinc-800"
                    onMouseDown={() => selectProject("")}
                  >
                    — без проекта —
                  </button>
                </li>
                {filteredProjects.map((p) => (
                  <li key={p.name}>
                    <button
                      type="button"
                      className="w-full px-3 py-2 text-left hover:bg-zinc-800"
                      onMouseDown={() => selectProject(p.name)}
                    >
                      <span className="font-mono">{p.name}</span>
                      <span className="ml-2 text-xs text-zinc-500">{p.goldenPath}</span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </label>
        </div>

        {project.trim() && canaryCtx && (
          <section className="mt-4 rounded border border-zinc-800 bg-zinc-950/50 p-4 text-sm">
            <h2 className="mb-3 font-medium text-zinc-300">Canary status</h2>
            <div className="grid gap-2 sm:grid-cols-2">
              <label className="flex items-center gap-2">
                <input type="checkbox" checked={canaryCtx.useCanaryLine} disabled className="rounded" />
                <span>В canary-аудитории (pin *)</span>
              </label>
              <div>
                <span className="text-zinc-500">Canary mode: </span>
                <span className="font-mono">{canaryCtx.canaryMode}</span>
              </div>
              <div>
                <span className="text-zinc-500">Rollout: </span>
                <span>
                  {canaryCtx.rolloutEnabled ? `enabled ${canaryCtx.canaryPercent}%` : "disabled"}
                </span>
              </div>
              <div>
                <span className="text-zinc-500">Bucket: </span>
                <span className="font-mono">{canaryCtx.projectBucket}</span>
              </div>
              <div>
                <span className="text-zinc-500">Stable line: </span>
                <span className="font-mono text-emerald-400">{canaryCtx.stableVersion || "—"}</span>
              </div>
              <div>
                <span className="text-zinc-500">Canary line: </span>
                <span className="font-mono text-amber-400">{canaryCtx.canaryVersion || "—"}</span>
              </div>
            </div>
            <div className="mt-3 flex flex-wrap items-center gap-4">
              <span className="text-xs text-zinc-500">Override (только preview):</span>
              <label className="flex items-center gap-1 text-xs">
                <input
                  type="radio"
                  name="forceChannel"
                  checked={forceChannel === ""}
                  onChange={() => setForceChannel("")}
                />
                auto
              </label>
              <label className="flex items-center gap-1 text-xs">
                <input
                  type="radio"
                  name="forceChannel"
                  checked={forceChannel === "stable"}
                  onChange={() => setForceChannel("stable")}
                />
                stable
              </label>
              <label className="flex items-center gap-1 text-xs">
                <input
                  type="radio"
                  name="forceChannel"
                  checked={forceChannel === "canary"}
                  onChange={() => setForceChannel("canary")}
                />
                canary
              </label>
            </div>
          </section>
        )}

        <button
          type="submit"
          disabled={loading}
          className="mt-4 rounded bg-sky-600 px-4 py-2 text-sm hover:bg-sky-500 disabled:opacity-50"
        >
          {loading ? "Resolving…" : "Resolve"}
        </button>
      </form>

      {error && <p className="text-red-400">{error}</p>}

      {result && (
        <section className="space-y-4">
          <div className="flex flex-wrap gap-6 text-sm">
            <div>
              <span className="text-zinc-500">Requested pin: </span>
              <span className="font-mono text-sky-400">{result.requestedPin}</span>
            </div>
            <div>
              <span className="text-zinc-500">Resolved: </span>
              <span className="font-mono text-emerald-400">{result.resolvedVersion}</span>
            </div>
            <div>
              <span className="text-zinc-500">Channel: </span>
              <span
                className={`font-mono ${result.channel === "canary" ? "text-amber-400" : "text-emerald-400"}`}
              >
                {result.channel}
              </span>
            </div>
          </div>
          <pre className="overflow-x-auto rounded-lg border border-zinc-800 bg-zinc-950 p-4 text-xs font-mono text-zinc-300">
            {JSON.stringify(result.manifest, null, 2)}
          </pre>
        </section>
      )}
    </div>
  );
}
