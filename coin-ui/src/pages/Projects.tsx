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
  const [error, setError] = useState<string | null>(null);
  const [updating, setUpdating] = useState<string | null>(null);

  function load(goldenPath: string, version: string) {
    setError(null);
    api
      .projects(goldenPath || undefined, version || undefined)
      .then((r) => setItems(r.items))
      .catch((err: Error) => setError(err.message));
  }

  useEffect(() => {
    load(filterGp, filterVer);
  }, [filterGp, filterVer]);

  function onFilter(e: FormEvent) {
    e.preventDefault();
    const next = new URLSearchParams();
    const g = gp.trim();
    const v = ver.trim();
    if (g) next.set("goldenPath", g);
    if (v) next.set("version", v);
    setSearchParams(next);
  }

  async function setMode(projectName: string, mode: string) {
    if (!can("publisher")) return;
    setUpdating(projectName);
    try {
      await api.updateProjectCanaryMode(projectName, mode, getActor() || undefined);
      load(filterGp, filterVer);
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
          GP binding и canary mode (allowlist / stable override для pin *)
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
          placeholder="version"
          className="rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
        />
        <button
          type="submit"
          className="rounded bg-zinc-800 px-4 py-2 text-sm hover:bg-zinc-700"
        >
          Фильтр
        </button>
        {(filterGp || filterVer) && (
          <button
            type="button"
            onClick={() => {
              setGp("");
              setVer("");
              setSearchParams({});
            }}
            className="text-sm text-zinc-500 hover:text-zinc-300"
          >
            Сброс
          </button>
        )}
      </form>

      {error && <p className="text-red-400">{error}</p>}

      <div className="overflow-x-auto rounded-lg border border-zinc-800">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-zinc-800 bg-zinc-900/80 text-zinc-500">
            <tr>
              <th className="px-4 py-3 font-medium">Project</th>
              <th className="px-4 py-3 font-medium">GP</th>
              <th className="px-4 py-3 font-medium">Version</th>
              <th className="px-4 py-3 font-medium">Canary mode</th>
              <th className="px-4 py-3 font-medium">Last seen</th>
            </tr>
          </thead>
          <tbody>
            {items.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-zinc-500">
                  Нет projects
                </td>
              </tr>
            ) : (
              items.map((p) => (
                <tr key={p.name} className="border-b border-zinc-800/60">
                  <td className="px-4 py-3 font-mono">{p.name}</td>
                  <td className="px-4 py-3">{p.goldenPath}</td>
                  <td className="px-4 py-3 font-mono">{p.version}</td>
                  <td className="px-4 py-3">
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
                      <span className="font-mono text-zinc-400">{p.canaryMode ?? "default"}</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-zinc-400">
                    {new Date(p.lastSeenAt).toLocaleString()}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
