import { FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api, getActor, setActor } from "../lib/api";

const inputClass =
  "w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm focus:border-sky-600 focus:outline-none";

export default function CreateGPProfile() {
  const navigate = useNavigate();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [actor, setActorField] = useState(getActor());
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    const trimmed = name.trim();
    if (!trimmed) {
      setError("Имя GP обязательно");
      return;
    }

    setSubmitting(true);
    setActor(actor);
    try {
      await api.createGPProfile({
        name: trimmed,
        description: description.trim() || undefined,
        actor: actor.trim() || undefined,
      });
      navigate(`/gp/${encodeURIComponent(trimmed)}?welcome=1`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка создания GP");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div>
        <Link to="/gp" className="text-sm text-sky-400 hover:underline">
          ← GP Profiles
        </Link>
        <h1 className="mt-2 text-2xl font-semibold">Новый Golden Path</h1>
        <p className="mt-1 text-sm text-zinc-400">
          Создайте GP profile — только имя и описание. Composition (agent, gp-content, branching-model)
          задаётся в первом draft.
        </p>
      </div>

      <form onSubmit={onSubmit} className="space-y-4 rounded-lg border border-zinc-800 bg-zinc-900/50 p-6">
        <label className="block space-y-1">
          <span className="text-sm text-zinc-300">Имя GP (coin.goldenPath)</span>
          <input
            className={inputClass}
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="go-app"
            required
          />
        </label>

        <label className="block space-y-1">
          <span className="text-sm text-zinc-300">Описание (опционально)</span>
          <textarea
            className={`${inputClass} min-h-[4rem] resize-y`}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Краткое описание стека"
            rows={3}
          />
        </label>

        <label className="block space-y-1">
          <span className="text-sm text-zinc-300">Actor (audit)</span>
          <input
            className={inputClass}
            value={actor}
            onChange={(e) => setActorField(e.target.value)}
            placeholder="operator@local"
          />
        </label>

        {error && <p className="text-sm text-red-400">{error}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="rounded bg-sky-600 px-4 py-2 text-sm font-medium text-white hover:bg-sky-500 disabled:opacity-50"
        >
          {submitting ? "Создание…" : "Создать GP profile"}
        </button>
      </form>
    </div>
  );
}
