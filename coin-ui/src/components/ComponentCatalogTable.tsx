import { Link } from "react-router-dom";
import type { Component } from "../api/types";
import { platformEditPath } from "../lib/platformComponentPaths";

type Props = {
  items: Component[];
  emptyLabel?: string;
  showType?: boolean;
  platformType?: "gp-content" | "branching-model";
};

export default function ComponentCatalogTable({
  items,
  emptyLabel = "Нет компонентов",
  showType = true,
  platformType,
}: Props) {
  return (
    <div className="overflow-x-auto rounded-lg border border-zinc-800">
      <table className="w-full text-left text-sm">
        <thead className="border-b border-zinc-800 bg-zinc-900/80 text-zinc-500">
          <tr>
            {showType && <th className="px-4 py-3 font-medium">Type</th>}
            <th className="px-4 py-3 font-medium">Name</th>
            <th className="px-4 py-3 font-medium"></th>
            <th className="px-4 py-3 font-medium">Latest version</th>
            <th className="px-4 py-3 font-medium">Versions</th>
            <th className="px-4 py-3 font-medium">Updated</th>
          </tr>
        </thead>
        <tbody>
          {items.length === 0 ? (
            <tr>
              <td colSpan={showType ? 6 : 5} className="px-4 py-8 text-center text-zinc-500">
                {emptyLabel}
              </td>
            </tr>
          ) : (
            items.map((c) => (
              <tr key={`${c.type}/${c.name}`} className="border-b border-zinc-800/60">
                {showType && <td className="px-4 py-3">{c.type}</td>}
                <td className="px-4 py-3 font-mono">{c.name}</td>
                <td className="px-4 py-3">
                  <Link
                    to={`/components/${c.type}/${c.name}`}
                    className="text-sky-400 hover:underline"
                  >
                    Detail →
                  </Link>
                  {platformType && c.latestVersion && platformEditPath(c.type, c.name, c.latestVersion) && (
                    <>
                      {" · "}
                      <Link
                        to={platformEditPath(c.type, c.name, c.latestVersion)!}
                        className="text-zinc-400 hover:text-sky-400 hover:underline"
                      >
                        Edit
                      </Link>
                    </>
                  )}
                </td>
                <td className="px-4 py-3 font-mono text-sky-400">{c.latestVersion || "—"}</td>
                <td className="px-4 py-3 tabular-nums">{c.versionCount}</td>
                <td className="px-4 py-3 text-zinc-400">
                  {c.versionCount > 0 ? new Date(c.latestCreatedAt).toLocaleDateString() : "—"}
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}
