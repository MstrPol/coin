import { FormEvent, useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import type { ComponentDetail, ComponentVersion, ComponentVersionDetail } from "../api/types";
import { useAuth } from "../context/AuthContext";
import { api, getActor } from "../lib/api";
import { isStudioType } from "../lib/componentStudio";

export default function ComponentDetailPage() {
  const { type = "", name = "", version: versionParam } = useParams();
  const navigate = useNavigate();
  const { can } = useAuth();
  const [detail, setDetail] = useState<ComponentDetail | null>(null);
  const [versions, setVersions] = useState<ComponentVersion[]>([]);
  const [selectedVersion, setSelectedVersion] = useState<string | null>(versionParam ?? null);
  const [versionDetail, setVersionDetail] = useState<ComponentVersionDetail | null>(null);
  const [showPublish, setShowPublish] = useState(false);
  const [pubVersion, setPubVersion] = useState("");
  const [pubMetadata, setPubMetadata] = useState("{}");
  const [pubContentRef, setPubContentRef] = useState("");
  const [publishing, setPublishing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    if (!type || !name) return;
    setError(null);
    setMessage(null);
    api
      .componentDetail(type, name)
      .then(async (d) => {
        setDetail(d);
        const vers = await api.componentVersions(type, name);
        setVersions(vers.items);
        const initial = versionParam ?? vers.items[0]?.version ?? null;
        setSelectedVersion(initial);
      })
      .catch((err: Error) => setError(err.message));
  }, [type, name, versionParam]);

  useEffect(() => {
    if (!type || !name || !selectedVersion) {
      setVersionDetail(null);
      return;
    }
    api
      .componentVersionDetail(type, name, selectedVersion)
      .then(setVersionDetail)
      .catch((err: Error) => setError(err.message));
  }, [type, name, selectedVersion]);

  async function onPublish(e: FormEvent) {
    e.preventDefault();
    if (!can("publisher") || !type || !name) return;
    setPublishing(true);
    setError(null);
    try {
      let metadata: Record<string, unknown> = {};
      let contentRef: Record<string, unknown> | undefined;
      metadata = JSON.parse(pubMetadata || "{}");
      if (pubContentRef.trim()) {
        contentRef = JSON.parse(pubContentRef);
      }
      await api.publishComponentVersion(type, name, {
        version: pubVersion.trim(),
        metadata,
        contentRef,
        actor: getActor() || undefined,
      });
      setMessage(`Published ${pubVersion}`);
      setShowPublish(false);
      const vers = await api.componentVersions(type, name);
      setVersions(vers.items);
      const d = await api.componentDetail(type, name);
      setDetail(d);
      setSelectedVersion(pubVersion.trim());
    } catch (err) {
      setError(err instanceof Error ? err.message : "publish failed");
    } finally {
      setPublishing(false);
    }
  }

  if (!type || !name) return null;

  return (
    <div className="space-y-6">
      <div>
        <Link to="/components" className="text-sm text-sky-400 hover:underline">
          ← Components
        </Link>
        <h1 className="mt-2 text-2xl font-semibold font-mono">
          {type}/{name}
        </h1>
        {detail && (
          <p className="mt-1 text-zinc-400">
            Latest: <span className="font-mono text-sky-400">{detail.latestVersion || "—"}</span>
            {" · "}
            {detail.versionCount} versions
          </p>
        )}
      </div>

      {error && <p className="text-red-400">{error}</p>}
      {message && <p className="text-emerald-400">{message}</p>}

        {can("publisher") && (
          <div className="flex flex-wrap gap-2">
            {isStudioType(type) && (
              <Link
                to="/studio"
                className="rounded border border-zinc-700 px-4 py-2 text-sm hover:bg-zinc-800"
              >
                Component Studio
              </Link>
            )}
            <button
          type="button"
          onClick={() => setShowPublish((v) => !v)}
          className="rounded bg-sky-600 px-4 py-2 text-sm hover:bg-sky-500"
        >
            Publish new version
          </button>
          </div>
        )}

      {showPublish && (
        <form onSubmit={onPublish} className="space-y-3 rounded-lg border border-zinc-800 bg-zinc-900 p-4">
          <input
            value={pubVersion}
            onChange={(e) => setPubVersion(e.target.value)}
            placeholder="version"
            required
            className="w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
          />
          <label className="block text-xs text-zinc-500">
            metadata (JSON)
            <textarea
              value={pubMetadata}
              onChange={(e) => setPubMetadata(e.target.value)}
              rows={4}
              className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 font-mono text-xs"
            />
          </label>
          <label className="block text-xs text-zinc-500">
            contentRef (JSON, optional)
            <textarea
              value={pubContentRef}
              onChange={(e) => setPubContentRef(e.target.value)}
              rows={3}
              className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 font-mono text-xs"
            />
          </label>
          <button
            type="submit"
            disabled={publishing}
            className="rounded bg-emerald-700 px-4 py-2 text-sm hover:bg-emerald-600 disabled:opacity-50"
          >
            {publishing ? "Publishing…" : "Publish"}
          </button>
        </form>
      )}

      <div className="grid gap-6 lg:grid-cols-2">
        <section>
          <h2 className="mb-3 font-medium">Versions</h2>
          <div className="overflow-x-auto rounded-lg border border-zinc-800">
            <table className="w-full text-left text-sm">
              <thead className="border-b border-zinc-800 bg-zinc-900/80 text-zinc-500">
                <tr>
                  <th className="px-3 py-2">Version</th>
                  <th className="px-3 py-2">Status</th>
                  <th className="px-3 py-2">Created</th>
                </tr>
              </thead>
              <tbody>
                {versions.length === 0 ? (
                  <tr>
                    <td colSpan={3} className="px-3 py-6 text-center text-zinc-500">
                      Нет версий
                    </td>
                  </tr>
                ) : (
                  versions.map((v) => (
                    <tr
                      key={v.version}
                      className={`cursor-pointer border-b border-zinc-800/60 ${
                        selectedVersion === v.version ? "bg-zinc-800/50" : ""
                      }`}
                      onClick={() => {
                        setSelectedVersion(v.version);
                        navigate(`/components/${type}/${name}/${encodeURIComponent(v.version)}`, {
                          replace: true,
                        });
                      }}
                    >
                      <td className="px-3 py-2 font-mono">{v.version}</td>
                      <td className="px-3 py-2">
                        <span
                          className={
                            v.status === "draft"
                              ? "text-amber-400"
                              : v.status === "canary"
                                ? "text-sky-400"
                                : ""
                          }
                        >
                          {v.status}
                        </span>
                        {v.status === "draft" && can("publisher") && isStudioType(type) && (
                          <>
                            {" "}
                            <Link
                              to={`/studio/${type}/${name}/${encodeURIComponent(v.version)}`}
                              className="text-sky-400 hover:underline"
                              onClick={(e) => e.stopPropagation()}
                            >
                              Studio
                            </Link>
                          </>
                        )}
                      </td>
                      <td className="px-3 py-2 text-zinc-400">
                        {new Date(v.createdAt).toLocaleString()}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </section>

        <section>
          <h2 className="mb-3 font-medium">Version detail</h2>
          {versionDetail ? (
            <div className="space-y-3">
              <pre className="overflow-x-auto rounded-lg border border-zinc-800 bg-zinc-950 p-3 text-xs font-mono text-zinc-300">
                {JSON.stringify(versionDetail.metadata, null, 2)}
              </pre>
              {versionDetail.contentRef && (
                <>
                  <p className="text-xs text-zinc-500">contentRef</p>
                  <pre className="overflow-x-auto rounded-lg border border-zinc-800 bg-zinc-950 p-3 text-xs font-mono text-zinc-300">
                    {JSON.stringify(versionDetail.contentRef, null, 2)}
                  </pre>
                </>
              )}
            </div>
          ) : (
            <p className="text-zinc-500">Выберите версию</p>
          )}
        </section>
      </div>

      {detail && detail.gpUsage.length > 0 && (
        <section>
          <h2 className="mb-3 font-medium">Использование в GP</h2>
          <div className="overflow-x-auto rounded-lg border border-zinc-800">
            <table className="w-full text-left text-sm">
              <thead className="border-b border-zinc-800 bg-zinc-900/80 text-zinc-500">
                <tr>
                  <th className="px-3 py-2">GP</th>
                  <th className="px-3 py-2">Release</th>
                  <th className="px-3 py-2">Status</th>
                </tr>
              </thead>
              <tbody>
                {detail.gpUsage.map((u) => (
                  <tr key={`${u.gpName}/${u.version}`} className="border-b border-zinc-800/60">
                    <td className="px-3 py-2">{u.gpName}</td>
                    <td className="px-3 py-2">
                      <Link
                        to={`/gp/${encodeURIComponent(u.gpName)}/releases/${encodeURIComponent(u.version)}`}
                        className="font-mono text-sky-400 hover:underline"
                      >
                        {u.version}
                      </Link>
                    </td>
                    <td className="px-3 py-2">{u.status}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}
    </div>
  );
}
