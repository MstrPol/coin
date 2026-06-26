import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import type { BlastRadius, CompositionItem, GPReleaseDetail } from "../api/types";
import BlastRadiusChart from "../components/BlastRadiusChart";
import GpCompositionForm from "../components/GpCompositionForm";
import { useAuth } from "../context/AuthContext";
import { api, getActor, PromoteBlockedError } from "../lib/api";
import { useGpCompositionEditor } from "../lib/useGpCompositionEditor";
import { GP_DRAFT_SLOT_ORDER } from "../lib/gpSlots";

import { platformEditPath } from "../lib/platformComponentPaths";

function componentLink(type: string, name: string, version: string): string | null {
  if ((type === "gp-content" || type === "branching-model") && name && version) {
    return platformEditPath(type, name, version);
  }
  if (type === "agent" && name) {
    return `/platform/runtime`;
  }
  return null;
}

export default function GpReleaseDetailPage() {
  const { name = "", version = "" } = useParams();
  const navigate = useNavigate();
  const { can } = useAuth();
  const [detail, setDetail] = useState<GPReleaseDetail | null>(null);
  const [blast, setBlast] = useState<BlastRadius | null>(null);
  const [promoting, setPromoting] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [blockingPins, setBlockingPins] = useState<
    { type: string; name: string; version: string; status: string }[] | null
  >(null);

  const isDraft = detail?.status === "draft";
  const canEdit = isDraft && can("publisher");
  const hubBase = `/gp/${encodeURIComponent(name)}`;

  const editor = useGpCompositionEditor(name, canEdit ? detail?.composition : undefined);

  const draftPinCount = useMemo(() => {
    if (!canEdit) return 0;
    let count = 0;
    for (const key of GP_DRAFT_SLOT_ORDER) {
      const ver = editor.composition[key];
      if (ver && editor.versionStatuses[key]?.[ver] === "draft") count++;
    }
    return count;
  }, [canEdit, editor.composition, editor.versionStatuses]);

  useEffect(() => {
    if (!name || !version) return;
    setError(null);
    setMessage(null);
    setBlast(null);

    api
      .gpRelease(name, version)
      .then(async (d) => {
        setDetail(d);
        if (d.status !== "draft") {
          const b = await api.blastRadius(name, version).catch(() => null);
          setBlast(b);
        }
      })
      .catch((err: Error) => setError(err.message));
  }, [name, version]);

  async function promote() {
    if (!name || !version) return;
    setPromoting(true);
    setError(null);
    setBlockingPins(null);
    try {
      const result = await api.promoteDraftGPRelease(name, version, getActor() || undefined);
      setMessage(`Promoted → ${result.version}`);
      setTimeout(
        () => navigate(`${hubBase}/releases/${encodeURIComponent(result.version)}`),
        800,
      );
    } catch (err) {
      if (err instanceof PromoteBlockedError) {
        setBlockingPins(err.blockingPins);
        setError(err.message);
      } else {
        setError(err instanceof Error ? err.message : "promote failed");
      }
    } finally {
      setPromoting(false);
    }
  }

  async function deleteDraft() {
    if (!name || !version || !isDraft) return;
    if (!window.confirm(`Удалить draft ${name}@${version}?`)) return;
    setDeleting(true);
    setError(null);
    try {
      await api.deleteGPReleaseDraft(name, version, getActor() || undefined);
      navigate(`${hubBase}/releases`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "delete failed");
    } finally {
      setDeleting(false);
    }
  }

  async function saveComposition() {
    if (!name || !version || !canEdit) return;
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      const updated = await api.updateGPReleaseDraft(name, version, {
        agentStackName: editor.agentStackName,
        gpContentName: editor.gpContentName,
        branchingModelName: editor.branchingModelName,
        composition: editor.composition,
        actor: getActor() || undefined,
      });
      setMessage("Composition сохранена");
      const fresh = await api.gpRelease(updated.name, updated.version);
      setDetail(fresh);
    } catch (err) {
      setError(err instanceof Error ? err.message : "save failed");
    } finally {
      setSaving(false);
    }
  }

  if (error && !detail) {
    return (
      <div className="space-y-4">
        <BackLink name={name} />
        <p className="text-red-400">{error}</p>
      </div>
    );
  }

  if (!detail) {
    return (
      <div className="space-y-4">
        <BackLink name={name} />
        <p className="text-zinc-500">Загрузка…</p>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex items-start justify-between gap-4">
        <div>
          <BackLink name={name} />
          <h1 className="mt-2 text-2xl font-semibold">
            {detail.name}@{detail.version}
          </h1>
          <p className="mt-1 text-zinc-400">
            Golden path release ·{" "}
            <span className={isDraft ? "text-amber-400" : "text-emerald-400"}>{detail.status}</span>
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          {canEdit && (
            <>
              <button
                type="button"
                onClick={deleteDraft}
                disabled={deleting}
                className="rounded border border-red-800 px-4 py-2 text-sm text-red-300 hover:bg-red-950/40 disabled:opacity-50"
              >
                {deleting ? "Deleting…" : "Delete draft"}
              </button>
              <button
                type="button"
                onClick={promote}
                disabled={promoting || draftPinCount > 0}
                title={draftPinCount > 0 ? "Опубликуйте все draft pins перед promote" : undefined}
                className="rounded bg-sky-600 px-4 py-2 text-sm font-medium hover:bg-sky-500 disabled:opacity-50"
              >
                {promoting ? "Promoting…" : "Promote → published"}
              </button>
            </>
          )}
        </div>
      </div>

      {error && <p className="text-red-400">{error}</p>}
      {blockingPins && blockingPins.length > 0 && (
        <div className="rounded border border-red-900/50 bg-red-950/30 px-4 py-3 text-sm text-red-200">
          <p className="font-medium text-red-300">Blocking pins</p>
          <ul className="mt-2 space-y-1 font-mono text-xs">
            {blockingPins.map((pin) => {
              const href = platformEditPath(pin.type, pin.name, pin.version);
              return (
                <li key={`${pin.type}/${pin.name}@${pin.version}`}>
                  {pin.type}/{pin.name}@{pin.version} ({pin.status})
                  {href && (
                    <>
                      {" "}
                      <Link to={href} className="text-sky-400 hover:underline">
                        Publish
                      </Link>
                    </>
                  )}
                </li>
              );
            })}
          </ul>
        </div>
      )}
      {message && <p className="text-emerald-400">{message}</p>}
      {!isDraft && (
        <p className="text-sm text-zinc-500">Published releases are immutable.</p>
      )}

      <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
        <h2 className="font-medium">Metadata</h2>
        <dl className="mt-4 grid gap-3 text-sm sm:grid-cols-2">
          <Field label="Status" value={detail.status} />
          <Field label="Created" value={new Date(detail.createdAt).toLocaleString()} />
          <Field label="Manifest hash" value={detail.manifestHash ?? "—"} mono />
          {detail.manifestUrl && (
            <div className="sm:col-span-2">
              <dt className="text-xs text-zinc-500">Manifest URL</dt>
              <dd className="mt-0.5 break-all font-mono text-zinc-300">{detail.manifestUrl}</dd>
            </div>
          )}
        </dl>
      </section>

      <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 className="font-medium">Composition</h2>
            <p className="mt-1 text-sm text-zinc-500">
              {canEdit
                ? "Редактируйте pins до promote (agent, gp-content, branching-model)."
                : "Pins для этой версии GP (agent, gp-content, branching-model)."}
            </p>
          </div>
          {canEdit && (
            <button
              type="button"
              onClick={saveComposition}
              disabled={saving || editor.loading}
              className="rounded bg-sky-600 px-4 py-2 text-sm font-medium hover:bg-sky-500 disabled:opacity-50"
            >
              {saving ? "Saving…" : "Save composition"}
            </button>
          )}
        </div>

        {canEdit ? (
          editor.loading ? (
            <p className="mt-4 text-sm text-zinc-500">Загрузка каталога…</p>
          ) : (
            <GpCompositionForm
              agentStackName={editor.agentStackName}
              agentStackOptions={editor.agentStackOptions}
              onAgentStackChange={editor.setAgentStackName}
              gpContentName={editor.gpContentName}
              gpContentOptions={editor.gpContentOptions}
              onGpContentChange={editor.setGpContentName}
              branchingModelName={editor.branchingModelName}
              branchingModelOptions={editor.branchingModelOptions}
              onBranchingModelChange={editor.setBranchingModelName}
              composition={editor.composition}
              versionOptions={editor.versionOptions}
              versionStatuses={editor.versionStatuses}
              onCompositionChange={editor.setComposition}
            />
          )
        ) : (
          <CompositionReadOnlyTable
            rows={detail.composition ?? []}
            canLink={can("publisher")}
          />
        )}
      </section>

      {blast && !isDraft && (
        <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
          <h2 className="font-medium">Blast radius</h2>
          <div className="mt-4">
            <BlastRadiusChart blast={blast} highlightVersion={detail.version} />
          </div>
        </section>
      )}
    </div>
  );
}

function CompositionReadOnlyTable({
  rows,
  canLink,
}: {
  rows: CompositionItem[];
  canLink: boolean;
}) {
  return (
    <table className="mt-4 w-full text-left text-sm">
      <thead className="text-zinc-500">
        <tr>
          <th className="pb-2 font-medium">Type</th>
          <th className="pb-2 font-medium">Name</th>
          <th className="pb-2 font-medium">Version</th>
          <th className="pb-2 font-medium" />
        </tr>
      </thead>
      <tbody>
        {rows.map((c) => {
          const href = componentLink(c.type, c.name, c.version);
          return (
            <tr key={`${c.type}/${c.name}`} className="border-t border-zinc-800/60">
              <td className="py-2">{c.type}</td>
              <td className="py-2">{c.name}</td>
              <td className="py-2 font-mono">{c.version}</td>
              <td className="py-2 text-right">
                {href && canLink && (
                  <Link to={href} className="text-sky-400 hover:underline">
                    Edit
                  </Link>
                )}
                {c.type === "gp-content" && !canLink && (
                  <Link to="/platform/build-stacks" className="text-sky-400 hover:underline">
                    Build stacks
                  </Link>
                )}
              </td>
            </tr>
          );
        })}
      </tbody>
    </table>
  );
}

function BackLink({ name }: { name: string }) {
  return (
    <Link
      to={`/gp/${encodeURIComponent(name)}/releases`}
      className="text-sm text-sky-400 hover:underline"
    >
      ← Releases
    </Link>
  );
}

function Field({
  label,
  value,
  mono,
}: {
  label: string;
  value: string;
  mono?: boolean;
}) {
  return (
    <div>
      <dt className="text-xs text-zinc-500">{label}</dt>
      <dd className={`mt-0.5 ${mono ? "font-mono" : ""} text-zinc-200`}>{value}</dd>
    </div>
  );
}
