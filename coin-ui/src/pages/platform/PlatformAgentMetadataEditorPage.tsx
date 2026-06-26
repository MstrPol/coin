import { FormEvent, useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import type { ComponentVersionDetail } from "../../api/types";
import { useAuth } from "../../context/AuthContext";
import { api, getActor } from "../../lib/api";
import {
  familyReleaseDetailPath,
  type PlatformFamilyId,
} from "../../lib/platformFamilyConfig";

const inputClass =
  "w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm focus:border-sky-600 focus:outline-none";

export default function PlatformAgentMetadataEditorPage({ familyId }: { familyId: PlatformFamilyId }) {
  const { name = "", version = "" } = useParams();
  const navigate = useNavigate();
  const { can } = useAuth();
  const [detail, setDetail] = useState<ComponentVersionDetail | null>(null);
  const [image, setImage] = useState("");
  const [digest, setDigest] = useState("");
  const [goarch, setGoarch] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    if (!name || !version) return;
    api
      .componentVersionDetail("agent", name, version)
      .then((d) => {
        setDetail(d);
        const m = d.metadata ?? {};
        setImage(String(m.image ?? ""));
        setDigest(String(m.digest ?? ""));
        setGoarch(String(m.goarch ?? ""));
      })
      .catch((err: Error) => setError(err.message));
  }, [name, version]);

  const readOnly = !can("publisher") || detail?.status !== "draft";

  async function onSave(e: FormEvent) {
    e.preventDefault();
    if (readOnly) return;
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      await api.patchComponentVersion("agent", name, version, {
        metadata: {
          image: image.trim() || undefined,
          digest: digest.trim() || undefined,
          goarch: goarch.trim() || undefined,
          runtime: name,
        },
        actor: getActor() || undefined,
      });
      setMessage("Saved");
    } catch (err) {
      setError(err instanceof Error ? err.message : "save failed");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div>
        <Link
          to={familyReleaseDetailPath(familyId, name, version)}
          className="text-sm text-sky-400 hover:underline"
        >
          ← {name}@{version}
        </Link>
        <h1 className="mt-2 text-2xl font-semibold font-mono">Edit agent metadata</h1>
        <p className="mt-1 text-sm text-zinc-400">Catch-up после CI image push</p>
      </div>

      <form onSubmit={onSave} className="space-y-4 rounded-lg border border-zinc-800 bg-zinc-900/50 p-6">
        <label className="block space-y-1">
          <span className="text-sm text-zinc-300">Image ref</span>
          <input
            className={inputClass}
            value={image}
            onChange={(e) => setImage(e.target.value)}
            disabled={readOnly}
          />
        </label>
        <label className="block space-y-1">
          <span className="text-sm text-zinc-300">Digest</span>
          <input
            className={inputClass}
            value={digest}
            onChange={(e) => setDigest(e.target.value)}
            disabled={readOnly}
          />
        </label>
        <label className="block space-y-1">
          <span className="text-sm text-zinc-300">GOARCH</span>
          <input
            className={inputClass}
            value={goarch}
            onChange={(e) => setGoarch(e.target.value)}
            disabled={readOnly}
          />
        </label>

        {error && <p className="text-sm text-red-400">{error}</p>}
        {message && <p className="text-sm text-emerald-400">{message}</p>}

        {!readOnly && (
          <button
            type="submit"
            disabled={saving}
            className="rounded bg-sky-600 px-4 py-2 text-sm font-medium text-white hover:bg-sky-500 disabled:opacity-50"
          >
            {saving ? "Saving…" : "Save metadata"}
          </button>
        )}
      </form>

      <button
        type="button"
        onClick={() => navigate(familyReleaseDetailPath(familyId, name, version))}
        className="text-sm text-zinc-400 hover:text-zinc-200"
      >
        Back to release detail
      </button>
    </div>
  );
}
