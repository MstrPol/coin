import { FormEvent, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import type { Project } from "../api/types";
import { useAuth } from "../context/AuthContext";
import { api, getActor } from "../lib/api";

const MODES = ["default", "canary", "stable"] as const;

export default function Projects() {
  const { can } = useAuth();
  const [searchParams, setSearchParams] = useSearchParams();
  const [items, setItems] = useState<Project[]>([]);
  const [gp, setGp] = useState(searchParams.get("goldenPath") ?? "");
  const [ver, setVer] = useState(searchParams.get("version") ?? "");
  const filterGp = searchParams.get("goldenPath") ?? "";
  const filterVer = searchParams.get("version") ?? "";
  const staleOnly = searchParams.get("stale") === "1" || searchParams.get("stale") === "true";
  const [error, setError] = useState<string | null>(null);
  const [updating, setUpdating] = useState<string | null>(null);

  function load(goldenPath: string, version: string, stale: boolean) {
    setError(null);
    api
      .projects(goldenPath || undefined, version || undefined, stale)
      .then((r) => setItems(r.items))
      .catch((err: Error) => setError(err.message));
  }

  useEffect(() => {
    load(filterGp, filterVer, staleOnly);
  }, [filterGp, filterVer, staleOnly]);

  function onFilter(e: FormEvent) {
    e.preventDefault();
    const next = new URLSearchParams(searchParams);
    const g = gp.trim();
    const v = ver.trim();
    if (g) next.set("goldenPath", g);
    else next.delete("goldenPath");
    if (v) next.set("version", v);
    else next.delete("version");
    setSearchParams(next);
  }

  async function setMode(projectName: string, mode: string) {
    if (!can("publisher")) return;
    setUpdating(projectName);
    try {
      await api.updateProjectCanaryMode(projectName, mode, getActor() || undefined);
      load(filterGp, filterVer, staleOnly);
    } catch (err) {
      setError(err instanceof Error ? err.message : "update failed");
    } finally {
      setUpdating(null);
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Projects</h1>
        <p className="mt-1 text-zinc-400">
          Регистрация при первом билде, обновление при каждом report
        </p>
      </div>

      <form onSubmit={onFilter} className="flex flex-wrap gap-3">
        <input
          value={gp}
          onChange={(e) => setGp(e.target.value)}
          placeholder="goldenPath"
          className="rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
        />
        <input
          value={ver}
          onChange={(e) => setVer(e.target.value)}
          placeholder="version pin"
          className="rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
        />
        <button
          type="submit"
          className="rounded bg-zinc-800 px-4 py-2 text-sm hover:bg-zinc-700"
        >
          Фильтр
        </button>
        {staleOnly && (
          <span className="self-center text-sm text-amber-400">Только stale (&gt;90d)</span>
        )}
      </form>

      {error && <p className="text-red-400">{error}</p>}

      <div className="overflow-x-auto rounded-lg border border-zinc-800">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-zinc-800 bg-zinc-900/80 text-zinc-500">
            <tr>
              <th className="px-3 py-3 font-medium">Name</th>
              <th className="px-3 py-3 font-medium">groupId</th>
              <th className="px-3 py-3 font-medium">artifactId</th>
              <th className="px-3 py-3 font-medium">git repo</th>
              <th className="px-3 py-3 font-medium">GP</th>
              <th className="px-3 py-3 font-medium">Version pin</th>
              <th className="px-3 py-3 font-medium">Canary</th>
              <th className="px-3 py-3 font-medium">Last build</th>
              <th className="px-3 py-3 font-medium">Branch</th>
            </tr>
          </thead>
          <tbody>
            {items.length === 0 ? (
              <tr>
                <td colSpan={9} className="px-4 py-8 text-center text-zinc-500">
                  Нет projects
                </td>
              </tr>
            ) : (
              items.map((p) => (
                <tr key={p.name} className="border-b border-zinc-800/60">
                  <td className="px-3 py-3 font-medium">{p.name}</td>
                  <td className="px-3 py-3 font-mono text-xs">{p.groupId || "—"}</td>
                  <td className="px-3 py-3 font-mono text-xs">{p.artifactId || p.name}</td>
                  <td className="px-3 py-3">
                    {p.gitRepoUrl ? (
                      <a
                        href={p.gitRepoUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-sky-400 hover:underline"
                      >
                        {p.gitRepoName || p.gitRepoUrl}
                      </a>
                    ) : (
                      "—"
                    )}
                  </td>
                  <td className="px-3 py-3">{p.goldenPath}</td>
                  <td className="px-3 py-3 font-mono text-xs">{p.version}</td>
                  <td className="px-3 py-3">
                    {can("publisher") ? (
                      <select
                        value={p.canaryMode ?? "default"}
                        disabled={updating === p.name}
                        onChange={(e) => setMode(p.name, e.target.value)}
                        className="rounded border border-zinc-700 bg-zinc-950 px-2 py-1 text-xs"
                      >
                        {MODES.map((m) => (
                          <option key={m} value={m}>
                            {m}
                          </option>
                        ))}
                      </select>
                    ) : (
                      p.canaryMode ?? "default"
                    )}
                  </td>
                  <td className="px-3 py-3 text-zinc-400">
                    {p.lastBuildAt ? new Date(p.lastBuildAt).toLocaleString() : "—"}
                  </td>
                  <td className="px-3 py-3 text-zinc-400">{p.branch || "—"}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
