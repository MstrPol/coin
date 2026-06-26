import { FormEvent, useState } from "react";
import { Link, useNavigate, useOutletContext, useParams } from "react-router-dom";
import { api, getActor } from "../../lib/api";
import { platformEditPath } from "../../lib/platformComponentPaths";
import {
  familyHubPath,
  familyReleaseDetailPath,
  PLATFORM_FAMILIES,
  type PlatformFamilyId,
} from "../../lib/platformFamilyConfig";

type HubContext = { familyId: PlatformFamilyId; compType: string };

const inputClass =
  "w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm focus:border-sky-600 focus:outline-none";

export default function PlatformNewDraftPage() {
  const { name = "" } = useParams();
  const { familyId, compType } = useOutletContext<HubContext>();
  const navigate = useNavigate();
  const family = PLATFORM_FAMILIES[familyId];
  const [version, setVersion] = useState("0.1.0-draft");
  const [image, setImage] = useState("");
  const [digest, setDigest] = useState("");
  const [goarch, setGoarch] = useState("amd64");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const isAgent = compType === "agent";

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    const ver = version.trim();
    if (!ver) return;
    setSubmitting(true);
    setError(null);
    try {
      const metadata =
        isAgent && (image.trim() || digest.trim())
          ? {
              image: image.trim() || undefined,
              digest: digest.trim() || undefined,
              goarch: goarch.trim() || undefined,
              runtime: name,
            }
          : undefined;
      await api.createDraftComponentVersion(compType, name, {
        version: ver,
        metadata,
        actor: getActor() || undefined,
      });
      const edit = platformEditPath(compType, name, ver);
      if (edit && !isAgent) {
        navigate(edit);
        return;
      }
      navigate(familyReleaseDetailPath(familyId, name, ver));
    } catch (err) {
      setError(err instanceof Error ? err.message : "create failed");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div>
        <Link to={familyHubPath(familyId, name)} className="text-sm text-sky-400 hover:underline">
          ← {name}
        </Link>
        <h1 className="mt-2 text-2xl font-semibold">New draft</h1>
        <p className="mt-1 text-sm text-zinc-400">{family.profileLabel}</p>
      </div>

      <form onSubmit={onSubmit} className="space-y-4 rounded-lg border border-zinc-800 bg-zinc-900/50 p-6">
        <label className="block space-y-1">
          <span className="text-sm text-zinc-300">Version</span>
          <input
            className={inputClass}
            value={version}
            onChange={(e) => setVersion(e.target.value)}
            required
          />
        </label>

        {isAgent && (
          <>
            <p className="text-xs text-zinc-500">
              Опционально: metadata для catch-up после CI push (image, digest, goarch).
            </p>
            <label className="block space-y-1">
              <span className="text-sm text-zinc-300">Image ref</span>
              <input className={inputClass} value={image} onChange={(e) => setImage(e.target.value)} />
            </label>
            <label className="block space-y-1">
              <span className="text-sm text-zinc-300">Digest</span>
              <input className={inputClass} value={digest} onChange={(e) => setDigest(e.target.value)} />
            </label>
            <label className="block space-y-1">
              <span className="text-sm text-zinc-300">GOARCH</span>
              <input className={inputClass} value={goarch} onChange={(e) => setGoarch(e.target.value)} />
            </label>
          </>
        )}

        {error && <p className="text-sm text-red-400">{error}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="rounded bg-sky-600 px-4 py-2 text-sm font-medium text-white hover:bg-sky-500 disabled:opacity-50"
        >
          {submitting ? "Создание…" : "Create draft"}
        </button>
      </form>
    </div>
  );
}
