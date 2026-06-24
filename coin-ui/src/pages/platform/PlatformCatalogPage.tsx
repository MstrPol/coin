import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import type { Component } from "../../api/types";
import ComponentCatalogTable from "../../components/ComponentCatalogTable";
import { useAuth } from "../../context/AuthContext";
import { api } from "../../lib/api";

type CatalogConfig = {
  title: string;
  description: string;
  types?: string[];
  namePrefix?: string;
  emptyLabel: string;
  showType?: boolean;
  hint?: string;
  studioType?: string;
};

function matchesConfig(c: Component, config: CatalogConfig): boolean {
  if (config.types && !config.types.includes(c.type)) return false;
  if (config.namePrefix && !c.name.startsWith(config.namePrefix)) return false;
  return true;
}

export default function PlatformCatalogPage({ config }: { config: CatalogConfig }) {
  const { can } = useAuth();
  const [items, setItems] = useState<Component[]>([]);
  const [error, setError] = useState<string | null>(null);
  const typeKey = (config.types ?? []).join(",");

  useEffect(() => {
    api
      .components()
      .then((r) => setItems(r.items.filter((c) => matchesConfig(c, config))))
      .catch((err: Error) => setError(err.message));
  }, [typeKey, config.namePrefix]);

  const studioLink = useMemo(() => {
    if (!config.studioType || !can("publisher")) return null;
    return `/studio/${config.studioType}`;
  }, [config.studioType, can]);

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-wide text-zinc-500">Platform</p>
          <h1 className="text-2xl font-semibold">{config.title}</h1>
          <p className="mt-1 text-zinc-400">{config.description}</p>
          {config.hint && <p className="mt-2 text-sm text-zinc-500">{config.hint}</p>}
        </div>
        {studioLink && (
          <Link
            to={studioLink}
            className="rounded bg-sky-600 px-4 py-2 text-sm font-medium hover:bg-sky-500"
          >
            Open Studio
          </Link>
        )}
      </div>

      {error && <p className="text-red-400">{error}</p>}
      <ComponentCatalogTable
        items={items}
        emptyLabel={config.emptyLabel}
        showType={config.showType ?? true}
      />
    </div>
  );
}
