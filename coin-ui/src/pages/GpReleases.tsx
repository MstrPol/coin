import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import type { GPRelease } from "../api/types";
import { useAuth } from "../context/AuthContext";
import { api } from "../lib/api";

function statusBadge(status: string) {
  if (status === "draft") {
    return (
      <span className="rounded bg-amber-950/50 px-2 py-0.5 text-xs text-amber-400">draft</span>
    );
  }
  return (
    <span className="rounded bg-emerald-950/50 px-2 py-0.5 text-xs text-emerald-400">
      {status}
    </span>
  );
}

export default function GpReleases() {
  const { can } = useAuth();
  const [items, setItems] = useState<GPRelease[]>([]);
  const [includeDrafts, setIncludeDrafts] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .gpReleases(undefined, includeDrafts)
      .then((r) => setItems(r.items))
      .catch((err: Error) => setError(err.message));
  }, [includeDrafts]);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">GP Releases</h1>
          <p className="mt-1 text-zinc-400">Published releases и draft snapshots</p>
        </div>
        {can("publisher") && (
          <Link
            to="/releases/publish"
            className="rounded bg-sky-600 px-4 py-2 text-sm font-medium hover:bg-sky-500"
          >
            Publish
          </Link>
        )}
      </div>

      <label className="flex items-center gap-2 text-sm text-zinc-400">
        <input
          type="checkbox"
          checked={includeDrafts}
          onChange={(e) => setIncludeDrafts(e.target.checked)}
          className="rounded border-zinc-600"
        />
        Показать drafts
      </label>

      {error && <p className="text-red-400">{error}</p>}

      <div className="overflow-x-auto rounded-lg border border-zinc-800">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-zinc-800 bg-zinc-900/80 text-zinc-500">
            <tr>
              <th className="px-4 py-3 font-medium">GP</th>
              <th className="px-4 py-3 font-medium">Version</th>
              <th className="px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 font-medium">Created</th>
              <th className="px-4 py-3 font-medium" />
            </tr>
          </thead>
          <tbody>
            {items.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-zinc-500">
                  Нет releases
                </td>
              </tr>
            ) : (
              items.map((r) => (
                <tr key={`${r.name}@${r.version}`} className="border-b border-zinc-800/60">
                  <td className="px-4 py-3">{r.name}</td>
                  <td className="px-4 py-3 font-mono">{r.version}</td>
                  <td className="px-4 py-3">{statusBadge(r.status)}</td>
                  <td className="px-4 py-3 text-zinc-400">
                    {new Date(r.createdAt).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3">
                    <Link
                      to={`/releases/${r.name}/${encodeURIComponent(r.version)}`}
                      className="text-sky-400 hover:underline"
                    >
                      Detail
                    </Link>
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
