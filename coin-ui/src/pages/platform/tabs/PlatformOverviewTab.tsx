import { useEffect, useState } from "react";
import { Link, useOutletContext, useParams, useSearchParams } from "react-router-dom";
import type { ComponentVersion } from "../../../api/types";
import { api } from "../../../lib/api";
import {
  familyHubPath,
  familyNewDraftPath,
  PLATFORM_FAMILIES,
  type PlatformFamilyId,
} from "../../../lib/platformFamilyConfig";

type HubContext = { familyId: PlatformFamilyId; compType: string };

export default function PlatformOverviewTab() {
  const { name = "" } = useParams();
  const { familyId, compType } = useOutletContext<HubContext>();
  const [searchParams, setSearchParams] = useSearchParams();
  const showWelcome = searchParams.get("welcome") === "1";
  const [versions, setVersions] = useState<ComponentVersion[]>([]);
  const [error, setError] = useState<string | null>(null);
  const family = PLATFORM_FAMILIES[familyId];
  const base = familyHubPath(familyId, name);

  useEffect(() => {
    if (!name) return;
    api
      .componentVersions(compType, name)
      .then((r) => setVersions(r.items))
      .catch((err: Error) => setError(err.message));
  }, [compType, name]);

  function dismissWelcome() {
    const next = new URLSearchParams(searchParams);
    next.delete("welcome");
    setSearchParams(next, { replace: true });
  }

  const published = versions.filter((v) => v.status === "published");
  const drafts = versions.filter((v) => v.status === "draft");
  const latestPublished = published[0]?.version ?? "—";

  return (
    <div className="space-y-6">
      {showWelcome && (
        <div className="rounded-lg border border-emerald-900/50 bg-emerald-950/30 px-4 py-4 flex flex-wrap items-center justify-between gap-3">
          <p className="text-sm text-emerald-200">
            Profile <span className="font-mono">{name}</span> создан. Создайте первый draft.
          </p>
          <div className="flex gap-2">
            <Link
              to={familyNewDraftPath(familyId, name)}
              className="rounded bg-sky-600 px-3 py-1.5 text-sm hover:bg-sky-500"
            >
              Create first draft
            </Link>
            <button
              type="button"
              onClick={dismissWelcome}
              className="text-sm text-zinc-400 hover:text-zinc-200"
            >
              Dismiss
            </button>
          </div>
        </div>
      )}

      {error && <p className="text-red-400">{error}</p>}

      {versions.length === 0 && !error && (
        <div className="rounded-lg border border-zinc-800 bg-zinc-900 p-6 text-sm text-zinc-400">
          Нет версий.{" "}
          <Link to={familyNewDraftPath(familyId, name)} className="text-sky-400 hover:underline">
            New draft
          </Link>
        </div>
      )}

      <section className="grid gap-4 sm:grid-cols-3">
        <Stat label="Latest published" value={latestPublished} mono />
        <Stat label="Drafts" value={String(drafts.length)} />
        <Stat label="Total versions" value={String(versions.length)} />
      </section>

      {family.runbookHref && (
        <p className="text-sm text-zinc-500">
          CI publish runbook:{" "}
          <a href={family.runbookHref} className="text-sky-400 hover:underline">
            agent-build-model
          </a>
        </p>
      )}

      <section className="flex flex-wrap gap-4 text-sm">
        <Link to={`${base}/releases`} className="text-sky-400 hover:underline">
          Releases →
        </Link>
      </section>
    </div>
  );
}

function Stat({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900 p-4">
      <div className="text-xs text-zinc-500">{label}</div>
      <div className={`mt-1 text-lg ${mono ? "font-mono" : ""}`}>{value}</div>
    </div>
  );
}
