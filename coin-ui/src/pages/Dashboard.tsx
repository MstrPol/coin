import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import type { DashboardStats } from "../api/types";
import { api } from "../lib/api";
import { appVersion } from "../lib/version";

export default function Dashboard() {
  const [ready, setReady] = useState("…");
  const [apiVersion, setApiVersion] = useState<string | null>(null);
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .ready()
      .then((d) => {
        setReady(d.status);
        setApiVersion(d.version ?? null);
      })
      .catch(() => {
        setReady("down");
        setApiVersion(null);
      });
    api
      .stats()
      .then(setStats)
      .catch((err: Error) => setError(err.message));
  }, []);

  const cards = stats
    ? [
        { label: "Projects", value: stats.projects, to: "/projects" },
        { label: "Stale projects", value: stats.staleProjects, to: "/projects?stale=1" },
        { label: "GP releases", value: stats.gpReleases, to: "/releases" },
        { label: "Build reports", value: stats.buildReports, to: "/build-reports" },
        { label: "Golden paths", value: stats.goldenPaths, to: "/releases" },
      ]
    : [];

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-semibold">Dashboard</h1>
        <p className="mt-1 text-zinc-400">Fleet overview (local samples)</p>
      </div>

      <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
        <h2 className="text-sm font-medium uppercase tracking-wide text-zinc-500">
          Platform status
        </h2>
        <p className="mt-2 font-mono text-2xl capitalize">{ready}</p>
        <div className="mt-4 grid gap-4 sm:grid-cols-2">
          <div>
            <p className="text-xs uppercase tracking-wide text-zinc-500">coin-api</p>
            <p className="mt-1 font-mono text-sky-400">{apiVersion ?? "—"}</p>
          </div>
          <div>
            <p className="text-xs uppercase tracking-wide text-zinc-500">coin-ui</p>
            <p className="mt-1 font-mono text-sky-400">{appVersion}</p>
          </div>
        </div>
      </section>

      {error && <p className="text-red-400">{error}</p>}

      {stats && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {cards.map((c) => (
            <Link
              key={c.label}
              to={c.to}
              className="rounded-lg border border-zinc-800 bg-zinc-900 p-5 transition hover:border-zinc-600"
            >
              <p className="text-sm text-zinc-500">{c.label}</p>
              <p className="mt-2 text-3xl font-semibold tabular-nums">{c.value}</p>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
