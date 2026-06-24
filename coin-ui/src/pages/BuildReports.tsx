import { FormEvent, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import type { BuildReport } from "../api/types";
import TablePagination, { pageToOffset } from "../components/TablePagination";
import { api, downloadCsv } from "../lib/api";

const DEFAULT_PAGE_SIZE = 50;

function intParam(sp: URLSearchParams, key: string, fallback: number): number {
  const v = parseInt(sp.get(key) ?? "", 10);
  return Number.isFinite(v) && v > 0 ? v : fallback;
}

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
  const [searchParams, setSearchParams] = useSearchParams();
  const [items, setItems] = useState<BuildReport[]>([]);
  const [total, setTotal] = useState(0);
  const [project, setProject] = useState(searchParams.get("project") ?? "");
  const [goldenPath, setGoldenPath] = useState(searchParams.get("goldenPath") ?? "");
  const [result, setResult] = useState(searchParams.get("result") ?? "");
  const [dateFrom, setDateFrom] = useState(searchParams.get("reportedAfter") ?? "");
  const [dateTo, setDateTo] = useState(searchParams.get("reportedBefore") ?? "");
  const filterProject = searchParams.get("project") ?? "";
  const filterGp = searchParams.get("goldenPath") ?? "";
  const filterResult = searchParams.get("result") ?? "";
  const filterAfter = searchParams.get("reportedAfter") ?? "";
  const filterBefore = searchParams.get("reportedBefore") ?? "";
  const page = intParam(searchParams, "page", 1);
  const pageSize = intParam(searchParams, "pageSize", DEFAULT_PAGE_SIZE);
  const [error, setError] = useState<string | null>(null);
  const [exporting, setExporting] = useState(false);

  useEffect(() => {
    setError(null);
    api
      .buildReports({
        project: filterProject || undefined,
        goldenPath: filterGp || undefined,
        result: filterResult || undefined,
        reportedAfter: filterAfter || undefined,
        reportedBefore: filterBefore || undefined,
        limit: pageSize,
        offset: pageToOffset(page, pageSize),
      })
      .then((r) => {
        setItems(r.items);
        setTotal(r.total);
      })
      .catch((err: Error) => setError(err.message));
  }, [filterProject, filterGp, filterResult, filterAfter, filterBefore, page, pageSize]);

  function onFilter(e: FormEvent) {
    e.preventDefault();
    const next = new URLSearchParams(searchParams);
    const p = project.trim();
    const gp = goldenPath.trim();
    const r = result.trim();
    const from = dateFrom.trim();
    const to = dateTo.trim();
    if (p) next.set("project", p);
    else next.delete("project");
    if (gp) next.set("goldenPath", gp);
    else next.delete("goldenPath");
    if (r) next.set("result", r);
    else next.delete("result");
    if (from) next.set("reportedAfter", from);
    else next.delete("reportedAfter");
    if (to) next.set("reportedBefore", to);
    else next.delete("reportedBefore");
    next.set("page", "1");
    setSearchParams(next);
  }

  function setPage(nextPage: number) {
    const next = new URLSearchParams(searchParams);
    next.set("page", String(nextPage));
    setSearchParams(next);
  }

  function setPageSize(nextSize: number) {
    const next = new URLSearchParams(searchParams);
    next.set("pageSize", String(nextSize));
    next.set("page", "1");
    setSearchParams(next);
  }

  async function exportCsv() {
    setExporting(true);
    setError(null);
    try {
      const path = api.buildReportsExportPath({
        project: filterProject || undefined,
        goldenPath: filterGp || undefined,
        result: filterResult || undefined,
        reportedAfter: filterAfter || undefined,
        reportedBefore: filterBefore || undefined,
      });
      await downloadCsv(path, `build-reports-${new Date().toISOString().slice(0, 10)}.csv`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "export failed");
    } finally {
      setExporting(false);
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">Build reports</h1>
          <p className="mt-1 text-zinc-400">История CI report из Jenkins Report stage</p>
        </div>
        <button
          type="button"
          onClick={exportCsv}
          disabled={exporting}
          className="rounded border border-zinc-700 px-4 py-2 text-sm hover:bg-zinc-800 disabled:opacity-50"
        >
          {exporting ? "Экспорт…" : "Export CSV"}
        </button>
      </div>

      <form onSubmit={onFilter} className="flex flex-wrap gap-3 items-end">
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
        <label className="flex flex-col gap-1 text-xs text-zinc-500">
          From
          <input
            type="date"
            value={dateFrom}
            onChange={(e) => setDateFrom(e.target.value)}
            className="rounded border border-zinc-700 bg-zinc-950 px-3 py-1.5 text-sm text-zinc-200"
          />
        </label>
        <label className="flex flex-col gap-1 text-xs text-zinc-500">
          To
          <input
            type="date"
            value={dateTo}
            onChange={(e) => setDateTo(e.target.value)}
            className="rounded border border-zinc-700 bg-zinc-950 px-3 py-1.5 text-sm text-zinc-200"
          />
        </label>
        <button
          type="submit"
          className="rounded bg-zinc-800 px-4 py-2 text-sm hover:bg-zinc-700"
        >
          Фильтр
        </button>
      </form>

      {error && <p className="text-red-400">{error}</p>}

      <TablePagination
        page={page}
        pageSize={pageSize}
        total={total}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
      />

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
