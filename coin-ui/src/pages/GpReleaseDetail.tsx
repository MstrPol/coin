import { useEffect, useState, type ReactNode } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import type { BlastRadius, ComponentVersion, GPReleaseDetail } from "../api/types";
import BlastRadiusChart from "../components/BlastRadiusChart";
import { useAuth } from "../context/AuthContext";
import { api, getActor } from "../lib/api";

type Tab = "overview" | "build-stack";

export default function GpReleaseDetailPage() {
  const { name = "", version = "" } = useParams();
  const navigate = useNavigate();
  const { can } = useAuth();
  const [tab, setTab] = useState<Tab>("overview");
  const [detail, setDetail] = useState<GPReleaseDetail | null>(null);
  const [blast, setBlast] = useState<BlastRadius | null>(null);
  const [buildStackVersions, setBuildStackVersions] = useState<ComponentVersion[]>([]);
  const [promoting, setPromoting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const isDraft = detail?.status === "draft";

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

  useEffect(() => {
    if (!name) return;
    api
      .componentVersionsOptional("gp-content", name)
      .then((r) => setBuildStackVersions(r.items))
      .catch(() => setBuildStackVersions([]));
  }, [name]);

  const pinnedGpContent = detail?.composition?.find((c) => c.type === "gp-content");

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

      <div className="flex gap-1 border-b border-zinc-800">
        <TabButton active={tab === "overview"} onClick={() => setTab("overview")}>
          Overview
        </TabButton>
        <TabButton active={tab === "build-stack"} onClick={() => setTab("build-stack")}>
          Build stack
        </TabButton>
      </div>

      {tab === "overview" && (
        <>
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
        </>
      )}

      {tab === "build-stack" && (
        <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6 space-y-4">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h2 className="font-medium">gp-content / {name}</h2>
              <p className="mt-1 text-sm text-zinc-400">
                Build stack для GP profile · primary path к Dockerfile и scripts
              </p>
              {pinnedGpContent && (
                <p className="mt-2 text-sm">
                  Pinned в release:{" "}
                  <span className="font-mono text-sky-400">{pinnedGpContent.version}</span>
                </p>
              )}
            </div>
            {can("publisher") && (
              <Link
                to={`/platform/build-stacks`}
                className="text-sm text-zinc-400 hover:text-zinc-200"
              >
                All build stacks →
              </Link>
            )}
          </div>
          <table className="w-full text-left text-sm">
            <thead className="text-zinc-500">
              <tr>
                <th className="pb-2 font-medium">Version</th>
                <th className="pb-2 font-medium">Status</th>
                <th className="pb-2 font-medium">Created</th>
                <th className="pb-2 font-medium"></th>
              </tr>
            </thead>
            <tbody>
              {buildStackVersions.length === 0 ? (
                <tr>
                  <td colSpan={4} className="py-6 text-zinc-500">
                    Нет версий gp-content для {name}
                  </td>
                </tr>
              ) : (
                buildStackVersions.map((v) => (
                  <tr key={v.version} className="border-t border-zinc-800/60">
                    <td className="py-2 font-mono">{v.version}</td>
                    <td className="py-2">{v.status}</td>
                    <td className="py-2 text-zinc-400">
                      {new Date(v.createdAt).toLocaleString()}
                    </td>
                    <td className="py-2">
                      {can("publisher") ? (
                        <Link
                          to={`/studio/gp-content/${name}/${encodeURIComponent(v.version)}`}
                          className="text-sky-400 hover:underline"
                        >
                          Studio
                        </Link>
                      ) : (
                        <Link
                          to={`/components/gp-content/${name}/${encodeURIComponent(v.version)}`}
                          className="text-sky-400 hover:underline"
                        >
                          Detail
                        </Link>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
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

function TabButton({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`border-b-2 px-4 py-2 text-sm ${
        active
          ? "border-sky-500 text-sky-400"
          : "border-transparent text-zinc-500 hover:text-zinc-300"
      }`}
    >
      {children}
    </button>
  );
}
