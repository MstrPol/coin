import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import type { GPRelease } from "../../../api/types";
import { useAuth } from "../../../context/AuthContext";
import { api, getActor } from "../../../lib/api";

function statusBadge(status: string) {
  if (status === "draft") {
    return (
      <span className="rounded bg-amber-950/50 px-2 py-0.5 text-xs text-amber-400">draft</span>
    );
  }
  return (
    <span className="rounded bg-emerald-950/50 px-2 py-0.5 text-xs text-emerald-400">{status}</span>
  );
}

export default function GpReleasesTab() {
  const { name = "" } = useParams();
  const { can } = useAuth();
  const [items, setItems] = useState<GPRelease[]>([]);
  const [includeDrafts, setIncludeDrafts] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deleting, setDeleting] = useState<string | null>(null);
  const base = `/gp/${encodeURIComponent(name)}`;

  function reload() {
    if (!name) return;
    api
      .gpReleases(name, includeDrafts)
      .then((r) => setItems(r.items))
      .catch((err: Error) => setError(err.message));
  }

  useEffect(() => {
    reload();
  }, [name, includeDrafts]);

  async function deleteDraft(version: string) {
    if (!window.confirm(`Удалить draft ${name}@${version}?`)) return;
    setDeleting(version);
    setError(null);
    try {
      await api.deleteGPReleaseDraft(name, version, getActor() || undefined);
      reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : "delete failed");
    } finally {
      setDeleting(null);
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <label className="flex items-center gap-2 text-sm text-zinc-400">
          <input
            type="checkbox"
            checked={includeDrafts}
            onChange={(e) => setIncludeDrafts(e.target.checked)}
            className="rounded border-zinc-600"
          />
          Показать drafts
        </label>
        {can("publisher") && (
          <Link to={`${base}/releases/new-draft`} className="text-sm text-sky-400 hover:underline">
            + New draft
          </Link>
        )}
      </div>

      {error && <p className="text-red-400">{error}</p>}

      <div className="overflow-x-auto rounded-lg border border-zinc-800">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-zinc-800 bg-zinc-900/80 text-zinc-500">
            <tr>
              <th className="px-4 py-3 font-medium">Version</th>
              <th className="px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 font-medium">Created</th>
              <th className="px-4 py-3 font-medium" />
            </tr>
          </thead>
          <tbody>
            {items.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-8 text-center text-zinc-500">
                  Нет releases
                </td>
              </tr>
            ) : (
              items.map((r) => (
                <tr key={r.version} className="border-b border-zinc-800/60">
                  <td className="px-4 py-3 font-mono">{r.version}</td>
                  <td className="px-4 py-3">{statusBadge(r.status)}</td>
                  <td className="px-4 py-3 text-zinc-400">
                    {new Date(r.createdAt).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <Link
                      to={`${base}/releases/${encodeURIComponent(r.version)}`}
                      className="text-sky-400 hover:underline"
                    >
                      Detail
                    </Link>
                    {r.status === "draft" && can("publisher") && (
                      <>
                        {" · "}
                        <button
                          type="button"
                          disabled={deleting === r.version}
                          onClick={() => deleteDraft(r.version)}
                          className="text-red-400 hover:underline disabled:opacity-50"
                        >
                          {deleting === r.version ? "…" : "Delete"}
                        </button>
                      </>
                    )}
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
