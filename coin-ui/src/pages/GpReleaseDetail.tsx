import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import type { ArtifactMeta, BlastRadius, GPReleaseDetail } from "../api/types";
import BlastRadiusChart from "../components/BlastRadiusChart";
import { useAuth } from "../context/AuthContext";
import { api, getActor } from "../lib/api";

export default function GpReleaseDetailPage() {
  const { name = "", version = "" } = useParams();
  const navigate = useNavigate();
  const { can } = useAuth();
  const [detail, setDetail] = useState<GPReleaseDetail | null>(null);
  const [blast, setBlast] = useState<BlastRadius | null>(null);
  const [artifacts, setArtifacts] = useState<ArtifactMeta[]>([]);
  const [selectedKey, setSelectedKey] = useState<string | null>(null);
  const [editBody, setEditBody] = useState("");
  const [saving, setSaving] = useState(false);
  const [promoting, setPromoting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const isDraft = detail?.status === "draft";

  useEffect(() => {
    if (!name || !version) return;
    setError(null);
    setMessage(null);
    setSelectedKey(null);
    setBlast(null);

    api
      .gpRelease(name, version)
      .then(async (d) => {
        setDetail(d);
        const arts = await api.listArtifacts(name, version).catch(() => ({ items: [] }));
        setArtifacts(arts.items);
        if (arts.items.length > 0) {
          setSelectedKey(arts.items[0].key);
        }
        if (d.status !== "draft") {
          const b = await api.blastRadius(name, version).catch(() => null);
          setBlast(b);
        }
      })
      .catch((err: Error) => setError(err.message));
  }, [name, version]);

  useEffect(() => {
    if (!name || !version || !selectedKey) return;
    api
      .getArtifact(name, version, selectedKey)
      .then((a) => setEditBody(a.body))
      .catch((err: Error) => setError(err.message));
  }, [name, version, selectedKey]);

  async function saveArtifact() {
    if (!name || !version || !selectedKey) return;
    setSaving(true);
    setError(null);
    try {
      await api.saveArtifact(name, version, selectedKey, editBody);
      setMessage("Artifact сохранён");
      const arts = await api.listArtifacts(name, version);
      setArtifacts(arts.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : "save failed");
    } finally {
      setSaving(false);
    }
  }

  async function promote() {
    if (!name || !version) return;
    setPromoting(true);
    setError(null);
    try {
      const result = await api.promoteDraftGPRelease(name, version, getActor() || undefined);
      setMessage(`Promoted → ${result.version}`);
      setTimeout(() => navigate(`/releases/${result.name}/${result.version}`), 800);
    } catch (err) {
      setError(err instanceof Error ? err.message : "promote failed");
    } finally {
      setPromoting(false);
    }
  }

  if (error && !detail) {
    return (
      <div className="space-y-4">
        <BackLink />
        <p className="text-red-400">{error}</p>
      </div>
    );
  }

  if (!detail) {
    return (
      <div className="space-y-4">
        <BackLink />
        <p className="text-zinc-500">Загрузка…</p>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex items-start justify-between gap-4">
        <div>
          <BackLink />
          <h1 className="mt-2 text-2xl font-semibold">
            {detail.name}@{detail.version}
          </h1>
          <p className="mt-1 text-zinc-400">
            Golden path release ·{" "}
            <span
              className={
                isDraft ? "text-amber-400" : "text-emerald-400"
              }
            >
              {detail.status}
            </span>
          </p>
        </div>
        {isDraft && can("publisher") && (
          <button
            type="button"
            onClick={promote}
            disabled={promoting}
            className="rounded bg-sky-600 px-4 py-2 text-sm font-medium hover:bg-sky-500 disabled:opacity-50"
          >
            {promoting ? "Promoting…" : "Promote → published"}
          </button>
        )}
      </div>

      {error && <p className="text-red-400">{error}</p>}
      {message && <p className="text-emerald-400">{message}</p>}

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
        <h2 className="font-medium">Composition</h2>
        <table className="mt-4 w-full text-left text-sm">
          <thead className="text-zinc-500">
            <tr>
              <th className="pb-2 font-medium">Type</th>
              <th className="pb-2 font-medium">Name</th>
              <th className="pb-2 font-medium">Version</th>
            </tr>
          </thead>
          <tbody>
            {detail.composition.map((c) => (
              <tr key={`${c.type}/${c.name}`} className="border-t border-zinc-800/60">
                <td className="py-2">{c.type}</td>
                <td className="py-2">{c.name}</td>
                <td className="py-2 font-mono">{c.version}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>

      {artifacts.length > 0 && (
        <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
          <h2 className="font-medium">GP Content (artifacts)</h2>
          <p className="mt-1 text-sm text-zinc-500">
            {isDraft
              ? "Редактирование доступно только для draft"
              : "Read-only — published release"}
          </p>
          <div className="mt-4 flex flex-wrap gap-2">
            {artifacts.map((a) => (
              <button
                key={a.key}
                type="button"
                onClick={() => setSelectedKey(a.key)}
                className={`rounded px-3 py-1 text-xs font-mono ${
                  selectedKey === a.key
                    ? "bg-sky-900/50 text-sky-300"
                    : "bg-zinc-800 text-zinc-400 hover:text-zinc-200"
                }`}
              >
                {a.key}
                <span className="ml-2 text-zinc-600">({a.size}b)</span>
              </button>
            ))}
          </div>
          {selectedKey && (
            <div className="mt-4 space-y-3">
              <div className="text-xs text-zinc-500 font-mono">{selectedKey}</div>
              <textarea
                value={editBody}
                onChange={(e) => setEditBody(e.target.value)}
                readOnly={!isDraft}
                rows={16}
                className="w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 font-mono text-xs focus:border-sky-600 focus:outline-none"
              />
              {isDraft && can("publisher") && (
                <button
                  type="button"
                  onClick={saveArtifact}
                  disabled={saving}
                  className="rounded bg-sky-600 px-4 py-2 text-sm hover:bg-sky-500 disabled:opacity-50"
                >
                  {saving ? "Saving…" : "Save artifact"}
                </button>
              )}
            </div>
          )}
        </section>
      )}

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

function BackLink() {
  return (
    <Link to="/releases" className="text-sm text-sky-400 hover:underline">
      ← GP Releases
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
