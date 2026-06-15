import { FormEvent, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import type { Component } from "../api/types";
import { useAuth } from "../context/AuthContext";
import { api, getActor } from "../lib/api";

export default function Components() {
  const { can } = useAuth();
  const [items, setItems] = useState<Component[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [newType, setNewType] = useState("gp-content");
  const [newName, setNewName] = useState("");
  const [creating, setCreating] = useState(false);

  function load() {
    api
      .components()
      .then((r) => setItems(r.items))
      .catch((err: Error) => setError(err.message));
  }

  useEffect(() => {
    load();
  }, []);

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setCreating(true);
    setError(null);
    setMessage(null);
    try {
      await api.createComponent({
        type: newType.trim(),
        name: newName.trim(),
        actor: getActor() || undefined,
      });
      setMessage(`Компонент ${newType}/${newName} создан`);
      setNewName("");
      setShowCreate(false);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "create failed");
    } finally {
      setCreating(false);
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">Components</h1>
          <p className="mt-1 text-zinc-400">Component registry — SoT в coin-api</p>
        </div>
        {can("publisher") && (
          <button
            type="button"
            onClick={() => setShowCreate((v) => !v)}
            className="rounded bg-sky-600 px-4 py-2 text-sm font-medium hover:bg-sky-500"
          >
            {showCreate ? "Отмена" : "Создать компонент"}
          </button>
        )}
      </div>

      {error && <p className="text-red-400">{error}</p>}
      {message && <p className="text-emerald-400">{message}</p>}

      {showCreate && can("publisher") && (
        <form
          onSubmit={handleCreate}
          className="rounded-lg border border-zinc-800 bg-zinc-900 p-4 space-y-3 max-w-md"
        >
          <div>
            <label className="block text-xs text-zinc-500 mb-1">Type</label>
            <input
              value={newType}
              onChange={(e) => setNewType(e.target.value)}
              className="w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
              placeholder="lib | gp-content | agent | executor"
              required
            />
          </div>
          <div>
            <label className="block text-xs text-zinc-500 mb-1">Name</label>
            <input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              className="w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
              placeholder="coin-lib | go-app | go"
              required
            />
          </div>
          <button
            type="submit"
            disabled={creating}
            className="rounded bg-sky-600 px-4 py-2 text-sm hover:bg-sky-500 disabled:opacity-50"
          >
            {creating ? "Создание…" : "Создать"}
          </button>
        </form>
      )}

      <div className="overflow-x-auto rounded-lg border border-zinc-800">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-zinc-800 bg-zinc-900/80 text-zinc-500">
            <tr>
              <th className="px-4 py-3 font-medium">Type</th>
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
                <td colSpan={6} className="px-4 py-8 text-center text-zinc-500">
                  Нет components
                </td>
              </tr>
            ) : (
              items.map((c) => (
                <tr key={`${c.type}/${c.name}`} className="border-b border-zinc-800/60">
                  <td className="px-4 py-3">{c.type}</td>
                  <td className="px-4 py-3 font-mono">{c.name}</td>
                  <td className="px-4 py-3">
                    <Link
                      to={`/components/${c.type}/${c.name}`}
                      className="text-sky-400 hover:underline"
                    >
                      Detail →
                    </Link>
                  </td>
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
