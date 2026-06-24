import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import type { BlastRadius, GPReleaseDetail } from "../api/types";
import BlastRadiusChart from "../components/BlastRadiusChart";
import { useAuth } from "../context/AuthContext";
import { api, getActor } from "../lib/api";

export default function GpReleaseDetailPage() {
  const { name = "", version = "" } = useParams();
  const navigate = useNavigate();
  const { can } = useAuth();
  const [detail, setDetail] = useState<GPReleaseDetail | null>(null);
  const [blast, setBlast] = useState<BlastRadius | null>(null);
  const [promoting, setPromoting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const isDraft = detail?.status === "draft";
  const hubBase = `/gp/${encodeURIComponent(name)}`;

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
    try {
      const result = await api.promoteDraftGPRelease(name, version, getActor() || undefined);
      setMessage(`Promoted → ${result.version}`);
      setTimeout(
        () => navigate(`${hubBase}/releases/${encodeURIComponent(result.version)}`),
        800,
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "promote failed");
    } finally {
      setPromoting(false);
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
        <div className="flex items-start justify-between gap-4">
          <h2 className="font-medium">Composition</h2>
          <Link to={`${hubBase}/build-stack`} className="text-sm text-sky-400 hover:underline">
            Build stack →
          </Link>
        </div>
        <table className="mt-4 w-full text-left text-sm">
          <thead className="text-zinc-500">
            <tr>
              <th className="pb-2 font-medium">Type</th>
              <th className="pb-2 font-medium">Name</th>
              <th className="pb-2 font-medium">Version</th>
            </tr>
          </thead>
          <tbody>
            {(detail.composition ?? []).map((c) => (
              <tr key={`${c.type}/${c.name}`} className="border-t border-zinc-800/60">
                <td className="py-2">{c.type}</td>
                <td className="py-2">{c.name}</td>
                <td className="py-2 font-mono">{c.version}</td>
              </tr>
            ))}
          </tbody>
        </table>
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
