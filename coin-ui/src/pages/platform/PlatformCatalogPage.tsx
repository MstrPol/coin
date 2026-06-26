import { FormEvent, useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import type { Component } from "../../api/types";
import ComponentCatalogTable from "../../components/ComponentCatalogTable";
import { useAuth } from "../../context/AuthContext";
import { api, getActor } from "../../lib/api";
import { platformEditPath } from "../../lib/platformComponentPaths";

type CatalogConfig = {
  title: string;
  description: string;
  types?: string[];
  namePrefix?: string;
  emptyLabel: string;
  showType?: boolean;
  hint?: string;
  createType?: "gp-content" | "branching-model";
};

function matchesConfig(c: Component, config: CatalogConfig): boolean {
  if (config.types && !config.types.includes(c.type)) return false;
  if (config.namePrefix && !c.name.startsWith(config.namePrefix)) return false;
  return true;
}

export default function PlatformCatalogPage({ config }: { config: CatalogConfig }) {
  const navigate = useNavigate();
  const { can } = useAuth();
  const [items, setItems] = useState<Component[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState("");
  const [newVersion, setNewVersion] = useState("0.1.0-draft");
  const [creating, setCreating] = useState(false);
  const typeKey = (config.types ?? []).join(",");

  async function reload() {
    const r = await api.components();
    setItems(r.items.filter((c) => matchesConfig(c, config)));
  }

  useEffect(() => {
    reload().catch((err: Error) => setError(err.message));
  }, [typeKey, config.namePrefix]);

  async function onCreateDraft(e: FormEvent) {
    e.preventDefault();
    if (!config.createType || !can("publisher")) return;
    const name = newName.trim();
    const version = newVersion.trim();
    if (!name || !version) return;
    setCreating(true);
    setError(null);
    try {
      await api.createComponent({
        type: config.createType,
        name,
        actor: getActor() || undefined,
      });
      await api.createDraftComponentVersion(config.createType, name, {
        version,
        actor: getActor() || undefined,
      });
      const edit = platformEditPath(config.createType, name, version);
      if (edit) {
        navigate(edit);
        return;
      }
      setShowCreate(false);
      setNewName("");
      await reload();
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
          <p className="text-xs uppercase tracking-wide text-zinc-500">Platform</p>
          <h1 className="text-2xl font-semibold">{config.title}</h1>
          <p className="mt-1 text-zinc-400">{config.description}</p>
          {config.hint && <p className="mt-2 text-sm text-zinc-500">{config.hint}</p>}
        </div>
        {config.createType && can("publisher") && (
          <button
            type="button"
            onClick={() => setShowCreate((v) => !v)}
            className="rounded bg-sky-600 px-4 py-2 text-sm font-medium hover:bg-sky-500"
          >
            {showCreate ? "Отмена" : "Create draft"}
          </button>
        )}
      </div>

      {showCreate && config.createType && (
        <form
          onSubmit={onCreateDraft}
          className="space-y-3 rounded-lg border border-zinc-800 bg-zinc-900 p-4 max-w-md"
        >
          <h2 className="text-sm font-medium">Новый {config.createType}</h2>
          <input
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder="name"
            required
            className="w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
          />
          <input
            value={newVersion}
            onChange={(e) => setNewVersion(e.target.value)}
            placeholder="version"
            required
            className="w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
          />
          <button
            type="submit"
            disabled={creating}
            className="rounded bg-emerald-700 px-4 py-2 text-sm hover:bg-emerald-600 disabled:opacity-50"
          >
            {creating ? "Создание…" : "Create & edit"}
          </button>
        </form>
      )}

      {error && <p className="text-red-400">{error}</p>}
      <ComponentCatalogTable
        items={items}
        emptyLabel={config.emptyLabel}
        showType={config.showType ?? true}
        platformType={config.createType}
      />
      {config.createType && (
        <p className="text-xs text-zinc-500">
          Lifecycle: draft → validate → register → publish. Открыть редактор — из{" "}
          <Link to="/components" className="text-sky-400 hover:underline">
            component detail
          </Link>{" "}
          или Create draft.
        </p>
      )}
    </div>
  );
}
