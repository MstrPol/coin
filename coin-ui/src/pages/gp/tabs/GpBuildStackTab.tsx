import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import type { ComponentVersion } from "../../../api/types";
import { useAuth } from "../../../context/AuthContext";
import { api } from "../../../lib/api";

export default function GpBuildStackTab() {
  const { name = "" } = useParams();
  const { can } = useAuth();
  const [versions, setVersions] = useState<ComponentVersion[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!name) return;
    api
      .componentVersionsOptional("gp-content", name)
      .then((r) => setVersions(r.items))
      .catch((err: Error) => setError(err.message));
  }, [name]);

  return (
    <div className="space-y-4">
      <p className="text-sm text-zinc-400">
        gp-content / <span className="font-mono text-zinc-200">{name}</span> — Dockerfile, scripts, schema
      </p>
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
            {versions.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-8 text-center text-zinc-500">
                  Нет версий gp-content
                </td>
              </tr>
            ) : (
              versions.map((v) => (
                <tr key={v.version} className="border-b border-zinc-800/60">
                  <td className="px-4 py-3 font-mono">{v.version}</td>
                  <td className="px-4 py-3">{v.status}</td>
                  <td className="px-4 py-3 text-zinc-400">
                    {new Date(v.createdAt).toLocaleString()}
                  </td>
                  <td className="px-4 py-3">
                    {can("publisher") ? (
                      <Link
                        to={`/studio/gp-content/${name}/${encodeURIComponent(v.version)}`}
                        className="text-sky-400 hover:underline"
                      >
                        Studio
                      </Link>
                    ) : (
                      <Link
                        to={`/components/gp-content/${name}/${encodeURIComponent(v.version)}`}
                        className="text-sky-400 hover:underline"
                      >
                        Detail
                      </Link>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
      <Link to="/platform/build-stacks" className="text-sm text-zinc-400 hover:text-zinc-200">
        All build stacks →
      </Link>
    </div>
  );
}
