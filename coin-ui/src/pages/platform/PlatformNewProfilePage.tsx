import { FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api, getActor } from "../../lib/api";
import {
  familyCatalogPath,
  familyHubPath,
  PLATFORM_FAMILIES,
  type PlatformFamilyId,
} from "../../lib/platformFamilyConfig";

const inputClass =
  "w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm focus:border-sky-600 focus:outline-none";

export default function PlatformNewProfilePage({ familyId }: { familyId: PlatformFamilyId }) {
  const navigate = useNavigate();
  const family = PLATFORM_FAMILIES[familyId];
  const [name, setName] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    const trimmed = name.trim();
    if (!trimmed) {
      setError("Имя профиля обязательно");
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      await api.createComponent({
        type: family.compType,
        name: trimmed,
        actor: getActor() || undefined,
      });
      navigate(`${familyHubPath(familyId, trimmed)}?welcome=1`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "create failed");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div>
        <Link to={familyCatalogPath(familyId)} className="text-sm text-sky-400 hover:underline">
          ← {family.catalogTitle}
        </Link>
        <h1 className="mt-2 text-2xl font-semibold">New {family.profileLabel.toLowerCase()}</h1>
        <p className="mt-1 text-sm text-zinc-400">
          Создайте профиль компонента. Первую версию добавьте как draft на hub.
        </p>
      </div>

      <form onSubmit={onSubmit} className="space-y-4 rounded-lg border border-zinc-800 bg-zinc-900/50 p-6">
        <label className="block space-y-1">
          <span className="text-sm text-zinc-300">Имя профиля</span>
          <input
            className={inputClass}
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={family.compType === "agent" ? "coin-agent-arm" : "go-app"}
            required
          />
        </label>

        {error && <p className="text-sm text-red-400">{error}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="rounded bg-sky-600 px-4 py-2 text-sm font-medium text-white hover:bg-sky-500 disabled:opacity-50"
        >
          {submitting ? "Создание…" : "Создать profile"}
        </button>
      </form>
    </div>
  );
}

// For route without prop — unused; prefer explicit familyId on Route element
