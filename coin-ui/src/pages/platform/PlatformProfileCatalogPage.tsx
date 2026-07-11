import { Link } from "react-router-dom";
import type { Component, ComponentVersion } from "../../api/types";
import { useAuth } from "../../context/AuthContext";
import {
  familyHubPath,
  familyNewProfilePath,
  type PlatformFamilyConfig,
} from "../../lib/platformFamilyConfig";

export type ProfileRow = {
  name: string;
  versions: ComponentVersion[];
  gpUsageCount: number;
  latestPublished?: string;
  draftCount: number;
};

function countByStatus(versions: ComponentVersion[]) {
  let draftCount = 0;
  let latestPublished: string | undefined;
  for (const v of versions) {
    if (v.status === "draft") draftCount++;
    if (v.status === "published" && !latestPublished) latestPublished = v.version;
  }
  return { draftCount, latestPublished };
}

export function buildProfileRows(
  components: Component[],
  versionsByName: Record<string, ComponentVersion[]>,
  gpUsageByName: Record<string, number>,
): ProfileRow[] {
  return components
    .map((c) => {
      const versions = versionsByName[c.name] ?? [];
      const { draftCount, latestPublished } = countByStatus(versions);
      return {
        name: c.name,
        versions,
        gpUsageCount: gpUsageByName[c.name] ?? 0,
        latestPublished: latestPublished ?? (c.latestVersion || undefined),
        draftCount,
      };
    })
    .sort((a, b) => a.name.localeCompare(b.name));
}

export default function PlatformProfileCatalogPage({
  family,
  rows,
  loading,
  error,
}: {
  family: PlatformFamilyConfig;
  rows: ProfileRow[];
  loading: boolean;
  error: string | null;
}) {
  const { can } = useAuth();

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-wide text-zinc-500">Platform</p>
          <h1 className="text-2xl font-semibold">{family.catalogTitle}</h1>
          <p className="mt-1 text-zinc-400">{family.catalogDescription}</p>
          {family.hint && <p className="mt-2 text-sm text-zinc-500">{family.hint}</p>}
        </div>
        {can("publisher") && (
          <Link
            to={familyNewProfilePath(family.id)}
            className="rounded bg-sky-600 px-4 py-2 text-sm font-medium hover:bg-sky-500"
          >
            New profile
          </Link>
        )}
      </div>

      {error && <p className="text-red-400">{error}</p>}

      <div className="overflow-x-auto rounded-lg border border-zinc-800">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-zinc-800 bg-zinc-900/80 text-zinc-500">
            <tr>
              <th className="px-4 py-3 font-medium">Profile</th>
              <th className="px-4 py-3 font-medium">Latest published</th>
              <th className="px-4 py-3 font-medium">Drafts</th>
              {family.compType !== "agent" && (
                <th className="px-4 py-3 font-medium">GP usage</th>
              )}
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={family.compType === "agent" ? 3 : 4} className="px-4 py-8 text-center text-zinc-500">
                  Загрузка…
                </td>
              </tr>
            ) : rows.length === 0 ? (
              <tr>
                <td colSpan={family.compType === "agent" ? 3 : 4} className="px-4 py-8 text-center text-zinc-500">
                  Нет профилей
                  {can("publisher") && (
                    <>
                      {" — "}
                      <Link to={familyNewProfilePath(family.id)} className="text-sky-400 hover:underline">
                        New profile
                      </Link>
                    </>
                  )}
                </td>
              </tr>
            ) : (
              rows.map((row) => (
                <tr key={row.name} className="border-b border-zinc-800/60 hover:bg-zinc-900/40">
                  <td className="px-4 py-3">
                    <Link
                      to={familyHubPath(family.id, row.name)}
                      className="font-mono text-sky-400 hover:underline"
                    >
                      {row.name}
                    </Link>
                  </td>
                  <td className="px-4 py-3 font-mono text-zinc-300">{row.latestPublished ?? "—"}</td>
                  <td className="px-4 py-3 text-zinc-400">{row.draftCount || "—"}</td>
                  {family.compType !== "agent" && (
                    <td className="px-4 py-3 text-zinc-400">{row.gpUsageCount || "—"}</td>
                  )}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
