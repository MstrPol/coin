import { useEffect, useState } from "react";
import type { BuildReport } from "../api/types";
import { api } from "../lib/api";

function resultBadge(result: string) {
  const r = result.toLowerCase();
  if (r === "success") {
    return (
      <span className="rounded bg-emerald-950/50 px-2 py-0.5 text-xs text-emerald-400">
        success
      </span>
    );
  }
  if (r === "failure") {
    return (
      <span className="rounded bg-red-950/50 px-2 py-0.5 text-xs text-red-400">failure</span>
    );
  }
  return <span className="text-zinc-400">{result}</span>;
}

export default function BuildReports() {
  const [items, setItems] = useState<BuildReport[]>([]);
  const [project, setProject] = useState("");
  const [goldenPath, setGoldenPath] = useState("");
  const [result, setResult] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .buildReports({
        project: project || undefined,
        goldenPath: goldenPath || undefined,
        result: result || undefined,
        limit: 100,
      })
      .then((r) => setItems(r.items))
      .catch((err: Error) => setError(err.message));
  }, [project, goldenPath, result]);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Build reports</h1>
        <p className="mt-1 text-zinc-400">История CI report из Jenkins Report stage</p>
      </div>

      <div className="flex flex-wrap gap-3">
        <input
          value={project}
          onChange={(e) => setProject(e.target.value)}
          placeholder="Project"
          className="rounded border border-zinc-700 bg-zinc-950 px-3 py-1.5 text-sm"
        />
        <input
          value={goldenPath}
          onChange={(e) => setGoldenPath(e.target.value)}
          placeholder="GP"
          className="rounded border border-zinc-700 bg-zinc-950 px-3 py-1.5 text-sm"
        />
        <select
          value={result}
          onChange={(e) => setResult(e.target.value)}
          className="rounded border border-zinc-700 bg-zinc-950 px-3 py-1.5 text-sm"
        >
          <option value="">Все результаты</option>
          <option value="success">success</option>
          <option value="failure">failure</option>
          <option value="aborted">aborted</option>
        </select>
      </div>

      {error && <p className="text-red-400">{error}</p>}

      <div className="overflow-x-auto rounded-lg border border-zinc-800">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-zinc-800 bg-zinc-900/80 text-zinc-500">
            <tr>
              <th className="px-4 py-3 font-medium">Project</th>
              <th className="px-4 py-3 font-medium">GP</th>
              <th className="px-4 py-3 font-medium">Version</th>
              <th className="px-4 py-3 font-medium">Result</th>
              <th className="px-4 py-3 font-medium">Branch</th>
              <th className="px-4 py-3 font-medium">Reported</th>
              <th className="px-4 py-3 font-medium">Build</th>
            </tr>
          </thead>
          <tbody>
            {items.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-zinc-500">
                  Нет reports
                </td>
              </tr>
            ) : (
              items.map((row) => (
                <tr key={row.id} className="border-b border-zinc-800/60">
                  <td className="px-4 py-3">{row.project}</td>
                  <td className="px-4 py-3">{row.goldenPath}</td>
                  <td className="px-4 py-3 font-mono text-xs">
                    {row.resolvedVersion ?? row.version}
                  </td>
                  <td className="px-4 py-3">{resultBadge(row.result)}</td>
                  <td className="px-4 py-3 text-zinc-400">{row.branch || "—"}</td>
                  <td className="px-4 py-3 text-zinc-400">
                    {new Date(row.reportedAt).toLocaleString()}
                  </td>
                  <td className="px-4 py-3">
                    {row.buildUrl ? (
                      <a
                        href={row.buildUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-sky-400 hover:underline"
                      >
                        link
                      </a>
                    ) : (
                      "—"
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
