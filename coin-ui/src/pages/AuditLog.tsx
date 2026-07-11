import { FormEvent, Fragment, useCallback, useEffect, useState } from "react";
import type { AuditLogEntry } from "../api/types";
import { api } from "../lib/api";

const PAGE_SIZE = 50;

export default function AuditLog() {
  const [items, setItems] = useState<AuditLogEntry[]>([]);
  const [entityType, setEntityType] = useState("");
  const [action, setAction] = useState("");
  const [filterEntityType, setFilterEntityType] = useState("");
  const [filterAction, setFilterAction] = useState("");
  const [offset, setOffset] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [expanded, setExpanded] = useState<number | null>(null);

  const load = useCallback(() => {
    setError(null);
    api
      .auditLog({
        entityType: filterEntityType || undefined,
        action: filterAction || undefined,
        limit: PAGE_SIZE,
        offset,
      })
      .then((r) => setItems(r.items))
      .catch((err: Error) => setError(err.message));
  }, [filterEntityType, filterAction, offset]);

  useEffect(() => {
    load();
  }, [load]);

  function onFilter(e: FormEvent) {
    e.preventDefault();
    setOffset(0);
    setFilterEntityType(entityType.trim());
    setFilterAction(action.trim());
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Audit log</h1>
        <p className="mt-1 text-zinc-400">Append-only журнал publish mutations</p>
      </div>

      <form onSubmit={onFilter} className="flex flex-wrap gap-3">
        <select
          value={entityType}
          onChange={(e) => setEntityType(e.target.value)}
          className="rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
        >
          <option value="">entityType — все</option>
          <option value="gp_release">gp_release</option>
          <option value="component_version">component_version</option>
        </select>
        <select
          value={action}
          onChange={(e) => setAction(e.target.value)}
          className="rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
        >
          <option value="">action — все</option>
          <option value="publish_gp_release">publish_gp_release</option>
          <option value="publish_component_version">publish_component_version</option>
        </select>
        <button
          type="submit"
          className="rounded bg-zinc-800 px-4 py-2 text-sm hover:bg-zinc-700"
        >
          Фильтр
        </button>
        {(filterEntityType || filterAction) && (
          <button
            type="button"
            onClick={() => {
              setEntityType("");
              setAction("");
              setFilterEntityType("");
              setFilterAction("");
              setOffset(0);
            }}
            className="text-sm text-zinc-500 hover:text-zinc-300"
          >
            Сброс
          </button>
        )}
      </form>

      {error && <p className="text-red-400">{error}</p>}

      <div className="overflow-x-auto rounded-lg border border-zinc-800">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-zinc-800 bg-zinc-900/80 text-zinc-500">
            <tr>
              <th className="px-4 py-3 font-medium">Time</th>
              <th className="px-4 py-3 font-medium">Action</th>
              <th className="px-4 py-3 font-medium">Entity</th>
              <th className="px-4 py-3 font-medium">Actor</th>
              <th className="px-4 py-3 font-medium" />
            </tr>
          </thead>
          <tbody>
            {items.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-zinc-500">
                  Нет записей
                </td>
              </tr>
            ) : (
              items.map((row) => (
                <Fragment key={row.id}>
                  <tr className="border-b border-zinc-800/60">
                    <td className="px-4 py-3 text-zinc-400 whitespace-nowrap">
                      {new Date(row.createdAt).toLocaleString()}
                    </td>
                    <td className="px-4 py-3 font-mono text-xs">{row.action}</td>
                    <td className="px-4 py-3">
                      <span className="text-zinc-500">{row.entityType}/</span>
                      <span className="font-mono">{row.entityKey}</span>
                    </td>
                    <td className="px-4 py-3 text-zinc-400">{row.actor ?? "—"}</td>
                    <td className="px-4 py-3">
                      <button
                        type="button"
                        onClick={() => setExpanded(expanded === row.id ? null : row.id)}
                        className="text-sky-400 hover:underline"
                      >
                        {expanded === row.id ? "Hide" : "Payload"}
                      </button>
                    </td>
                  </tr>
                  {expanded === row.id && (
                    <tr className="border-b border-zinc-800/60 bg-zinc-950/50">
                      <td colSpan={5} className="px-4 py-3">
                        <pre className="overflow-x-auto text-xs text-zinc-400">
                          {JSON.stringify(row.payload, null, 2)}
                        </pre>
                      </td>
                    </tr>
                  )}
                </Fragment>
              ))
            )}
          </tbody>
        </table>
      </div>

      <div className="flex gap-3 text-sm">
        <button
          type="button"
          disabled={offset === 0}
          onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
          className="rounded border border-zinc-700 px-3 py-1.5 disabled:opacity-40 hover:bg-zinc-800"
        >
          ← Prev
        </button>
        <span className="py-1.5 text-zinc-500">offset {offset}</span>
        <button
          type="button"
          disabled={items.length < PAGE_SIZE}
          onClick={() => setOffset(offset + PAGE_SIZE)}
          className="rounded border border-zinc-700 px-3 py-1.5 disabled:opacity-40 hover:bg-zinc-800"
        >
          Next →
        </button>
      </div>
    </div>
  );
}
