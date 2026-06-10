import { useEffect, useState } from "react";
import type { Component } from "../api/types";
import { api } from "../lib/api";

export default function Components() {
  const [items, setItems] = useState<Component[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .components()
      .then((r) => setItems(r.items))
      .catch((err: Error) => setError(err.message));
  }, []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Components</h1>
        <p className="mt-1 text-zinc-400">Component registry (read-only)</p>
      </div>

      {error && <p className="text-red-400">{error}</p>}

      <div className="overflow-x-auto rounded-lg border border-zinc-800">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-zinc-800 bg-zinc-900/80 text-zinc-500">
            <tr>
              <th className="px-4 py-3 font-medium">Type</th>
              <th className="px-4 py-3 font-medium">Name</th>
              <th className="px-4 py-3 font-medium">Latest version</th>
              <th className="px-4 py-3 font-medium">Versions</th>
              <th className="px-4 py-3 font-medium">Updated</th>
            </tr>
          </thead>
          <tbody>
            {items.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-zinc-500">
                  Нет components
                </td>
              </tr>
            ) : (
              items.map((c) => (
                <tr key={`${c.type}/${c.name}`} className="border-b border-zinc-800/60">
                  <td className="px-4 py-3">{c.type}</td>
                  <td className="px-4 py-3 font-mono">{c.name}</td>
                  <td className="px-4 py-3 font-mono text-sky-400">
                    {c.latestVersion || "—"}
                  </td>
                  <td className="px-4 py-3 tabular-nums">{c.versionCount}</td>
                  <td className="px-4 py-3 text-zinc-400">
                    {c.versionCount > 0
                      ? new Date(c.latestCreatedAt).toLocaleDateString()
                      : "—"}
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
