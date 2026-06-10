import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import type { DashboardStats } from "../api/types";
import { api } from "../lib/api";

export default function Dashboard() {
  const [ready, setReady] = useState("…");
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.ready().then((d) => setReady(d.status)).catch(() => setReady("down"));
    api
      .stats()
      .then(setStats)
      .catch((err: Error) => setError(err.message));
  }, []);

  const cards = stats
    ? [
        { label: "Projects", value: stats.projects, to: "/projects" },
        { label: "GP releases", value: stats.gpReleases, to: "/releases" },
        { label: "Build reports", value: stats.buildReports, to: "/projects" },
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
          coin-api
        </h2>
        <p className="mt-2 font-mono text-2xl capitalize">{ready}</p>
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
