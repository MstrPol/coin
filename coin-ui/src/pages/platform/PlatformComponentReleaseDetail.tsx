import { useEffect, useState } from "react";
import { Link, useNavigate, useOutletContext, useParams } from "react-router-dom";
import type { ComponentVersionDetail } from "../../api/types";
import { useAuth } from "../../context/AuthContext";
import { api, getActor } from "../../lib/api";
import { platformEditPath, supportsDraftDelete } from "../../lib/platformComponentPaths";
import {
  familyHubPath,
  type PlatformFamilyId,
} from "../../lib/platformFamilyConfig";
import PlatformComponentEditor from "../../components/platform/PlatformComponentEditor";

type HubContext = { familyId: PlatformFamilyId; compType: string };

function statusBadge(status: string) {
  if (status === "draft") {
    return (
      <span className="rounded bg-amber-950/50 px-2 py-0.5 text-xs text-amber-400">draft</span>
    );
  }
  return (
    <span className="rounded bg-emerald-950/50 px-2 py-0.5 text-xs text-emerald-400">{status}</span>
  );
}

export default function PlatformComponentReleaseDetail() {
  const { name = "", version = "" } = useParams();
  const { familyId, compType } = useOutletContext<HubContext>();
  const navigate = useNavigate();
  const { can } = useAuth();
  const [detail, setDetail] = useState<ComponentVersionDetail | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [promoting, setPromoting] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const hubBase = familyHubPath(familyId, name);
  const isAgent = compType === "agent";
  const canDeleteDraft = supportsDraftDelete(compType);
  const isEditableArtifact = compType === "gp-content" || compType === "branching-model";

  useEffect(() => {
    if (!name || !version) return;
    api
      .componentVersionDetail(compType, name, version)
      .then(setDetail)
      .catch((err: Error) => setError(err.message));
  }, [compType, name, version]);

  async function promote() {
    if (!name || !version) return;
    setPromoting(true);
    setError(null);
    setMessage(null);
    try {
      if (isEditableArtifact) {
        const v = await api.validateComponentPackage(compType, name, version);
        if (!v.valid) {
          setError("Validation failed — fix artifacts before publish");
          return;
        }
        if (!detail?.contentRef) {
          await api.registerComponentPackage(compType, name, version, {
            actor: getActor() || undefined,
          });
        }
      }
      await api.promoteComponentVersion(compType, name, version, getActor() || undefined);
      setMessage("Published");
      const refreshed = await api.componentVersionDetail(compType, name, version);
      setDetail(refreshed);
    } catch (err) {
      setError(err instanceof Error ? err.message : "promote failed");
    } finally {
      setPromoting(false);
    }
  }

  async function deleteDraft() {
    if (!name || !version || detail?.status !== "draft" || !canDeleteDraft) return;
    if (!window.confirm(`Удалить draft ${name}@${version}?`)) return;
    setDeleting(true);
    setError(null);
    setMessage(null);
    try {
      await api.deleteComponentVersionDraft(compType, name, version, getActor() || undefined);
      navigate(`${hubBase}/releases`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "delete failed");
    } finally {
      setDeleting(false);
    }
  }

  if (isEditableArtifact && detail) {
    return (
      <div className="space-y-4">
        <Link to={`${hubBase}/releases`} className="text-sm text-sky-400 hover:underline">
          ← Releases
        </Link>
        <PlatformComponentEditor
          type={compType}
          name={name}
          version={version}
          canEdit={detail.status === "draft" && can("publisher")}
        />
      </div>
    );
  }

  const meta = detail?.metadata ?? {};
  const derived =
    detail?.derivedExecutorPin ??
  (isAgent ? { type: "executor", name: "coin-executor", version } : null);

  return (
    <div className="space-y-6">
      <div>
        <Link to={`${hubBase}/releases`} className="text-sm text-sky-400 hover:underline">
          ← Releases
        </Link>
        <div className="mt-2 flex flex-wrap items-center gap-3">
          <h1 className="text-2xl font-semibold font-mono">
            {name}@{version}
          </h1>
          {detail && statusBadge(detail.status)}
        </div>
      </div>

      {error && <p className="text-red-400">{error}</p>}
      {message && <p className="text-emerald-400">{message}</p>}

      {isAgent && (
        <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6 space-y-3">
          <h2 className="font-medium">Agent metadata</h2>
          {detail?.status === "draft" && (
            <p className="text-xs text-zinc-500">
              CI path: draft от <code className="text-zinc-400">publish-agent.sh</code> → проверьте image и
              digest → Publish. Ручной catch-up: Edit metadata.
            </p>
          )}
          <dl className="grid gap-2 text-sm sm:grid-cols-2">
            <div>
              <dt className="text-zinc-500">Image</dt>
              <dd className="font-mono text-zinc-300 break-all">{String(meta.image ?? "—")}</dd>
            </div>
            <div>
              <dt className="text-zinc-500">Digest</dt>
              <dd className="font-mono text-zinc-300 break-all">{String(meta.digest ?? "—")}</dd>
            </div>
          </dl>
          {derived && (
            <div className="pt-2 border-t border-zinc-800">
              <p className="text-xs text-zinc-500 mb-1">Derived executor pin (read-only)</p>
              <p className="font-mono text-sm text-zinc-300">
                {derived.type}/{derived.name}@{derived.version}
              </p>
            </div>
          )}
        </section>
      )}

      {!isAgent && detail && (
        <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6 text-sm text-zinc-400">
          Read-only published release.{" "}
          {detail.status === "draft" && platformEditPath(compType, name, version) && (
            <Link to={platformEditPath(compType, name, version)!} className="text-sky-400 hover:underline">
              Edit draft
            </Link>
          )}
        </section>
      )}

      {detail?.status === "draft" && can("publisher") && (
        <div className="flex flex-wrap gap-3">
          {isAgent && platformEditPath(compType, name, version) && (
            <Link
              to={platformEditPath(compType, name, version)!}
              className="rounded border border-zinc-700 px-4 py-2 text-sm hover:bg-zinc-800"
            >
              Edit metadata
            </Link>
          )}
          {canDeleteDraft && isAgent && (
            <button
              type="button"
              disabled={deleting}
              onClick={() => void deleteDraft()}
              className="rounded border border-red-800 px-4 py-2 text-sm text-red-300 hover:bg-red-950/40 disabled:opacity-50"
            >
              {deleting ? "Deleting…" : "Delete draft"}
            </button>
          )}
          <button
            type="button"
            disabled={promoting}
            onClick={() => void promote()}
            className="rounded bg-emerald-700 px-4 py-2 text-sm font-medium hover:bg-emerald-600 disabled:opacity-50"
          >
            {promoting ? "Publishing…" : "Publish"}
          </button>
        </div>
      )}

      {detail?.status === "published" && (
        <button
          type="button"
          onClick={() => navigate(`${hubBase}/releases`)}
          className="text-sm text-zinc-400 hover:text-zinc-200"
        >
          Back to releases
        </button>
      )}
    </div>
  );
}
