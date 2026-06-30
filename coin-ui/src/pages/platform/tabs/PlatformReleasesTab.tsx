import { useEffect, useState } from "react";
import { Link, useOutletContext, useParams } from "react-router-dom";
import type { ComponentVersion } from "../../../api/types";
import { useAuth } from "../../../context/AuthContext";
import { api, getActor } from "../../../lib/api";
import { platformEditPath } from "../../../lib/platformComponentPaths";
import {
  familyNewDraftPath,
  familyReleaseDetailPath,
  type PlatformFamilyId,
} from "../../../lib/platformFamilyConfig";

type HubContext = { familyId: PlatformFamilyId; compType: string };

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

export default function PlatformReleasesTab() {
  const { name = "" } = useParams();
  const { familyId, compType } = useOutletContext<HubContext>();
  const { can } = useAuth();
  const [items, setItems] = useState<ComponentVersion[]>([]);
  const [includeDrafts, setIncludeDrafts] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deletingVersion, setDeletingVersion] = useState<string | null>(null);
  const isAgent = compType === "agent";

  function reload() {
    if (!name) return;
    api
      .componentVersions(compType, name)
      .then((r) => {
        const sorted = [...r.items].sort(
          (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
        );
        setItems(includeDrafts ? sorted : sorted.filter((v) => v.status === "published"));
      })
      .catch((err: Error) => setError(err.message));
  }

  useEffect(() => {
    reload();
  }, [name, compType, includeDrafts]);

  async function deleteDraft(ver: string) {
    if (!name || !isAgent) return;
    if (!window.confirm(`Удалить draft ${name}@${ver}?`)) return;
    setDeletingVersion(ver);
    setError(null);
    try {
      await api.deleteComponentVersionDraft(compType, name, ver, getActor() || undefined);
      reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : "delete failed");
    } finally {
      setDeletingVersion(null);
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
          <Link to={familyNewDraftPath(familyId, name)} className="text-sm text-sky-400 hover:underline">
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
                      to={familyReleaseDetailPath(familyId, name, r.version)}
                      className="text-sky-400 hover:underline"
                    >
                      Detail
                    </Link>
                    {r.status === "draft" && can("publisher") && platformEditPath(compType, name, r.version) && (
                      <>
                        {" · "}
                        <Link
                          to={platformEditPath(compType, name, r.version)!}
                          className="text-sky-400 hover:underline"
                        >
                          Edit
                        </Link>
                      </>
                    )}
                    {r.status === "draft" && isAgent && can("publisher") && (
                      <>
                        {" · "}
                        <button
                          type="button"
                          disabled={deletingVersion === r.version}
                          onClick={() => void deleteDraft(r.version)}
                          className="text-red-400 hover:underline disabled:opacity-50"
                        >
                          {deletingVersion === r.version ? "Deleting…" : "Delete"}
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
