import { FormEvent, useEffect, useState } from "react";
import type { PlatformSettings } from "../api/types";
import { useAuth } from "../context/AuthContext";
import { api, getActor } from "../lib/api";

const inputClass =
  "mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono";

export default function PlatformSettingsPage() {
  const { can } = useAuth();
  const [settings, setSettings] = useState<PlatformSettings | null>(null);
  const [mavenBase, setMavenBase] = useState("");
  const [credentialsId, setCredentialsId] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    api
      .platformSettings()
      .then((s) => {
        setSettings(s);
        setMavenBase(s.nexusMavenBase);
        setCredentialsId(s.nexusCredentialsId);
      })
      .catch((err: Error) => setError(err.message));
  }, []);

  async function onSave(e: FormEvent) {
    e.preventDefault();
    if (!can("publisher")) return;
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      await api.updatePlatformSettings({
        nexusMavenBase: mavenBase.trim(),
        nexusCredentialsId: credentialsId.trim(),
        actor: getActor() || undefined,
      });
      const s = await api.platformSettings();
      setSettings(s);
      setMessage("Сохранено");
    } catch (err) {
      setError(err instanceof Error ? err.message : "save failed");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Platform settings</h1>
        <p className="mt-1 text-zinc-400">Глобальные настройки Nexus для control plane</p>
      </div>

      {error && <p className="text-red-400">{error}</p>}
      {message && <p className="text-emerald-400">{message}</p>}

      {can("publisher") ? (
        <form onSubmit={onSave} className="max-w-2xl space-y-6">
          <section className="space-y-4 rounded-lg border border-zinc-800 bg-zinc-900 p-6">
            <h2 className="font-medium">Nexus</h2>
            <label className="block">
              <span className="text-xs text-zinc-500">nexus.mavenBase</span>
              <input
                value={mavenBase}
                onChange={(e) => setMavenBase(e.target.value)}
                className={inputClass}
              />
            </label>
            <label className="block">
              <span className="text-xs text-zinc-500">nexus.credentialsId</span>
              <input
                value={credentialsId}
                onChange={(e) => setCredentialsId(e.target.value)}
                className={inputClass}
              />
            </label>
          </section>

          <button
            type="submit"
            disabled={saving}
            className="rounded bg-sky-600 px-4 py-2 text-sm hover:bg-sky-500 disabled:opacity-50"
          >
            {saving ? "Сохранение…" : "Сохранить"}
          </button>
        </form>
      ) : (
        settings && (
          <dl className="max-w-lg space-y-3 rounded-lg border border-zinc-800 bg-zinc-900 p-6 text-sm">
            <div>
              <dt className="text-zinc-500">nexus.mavenBase</dt>
              <dd className="font-mono">{settings.nexusMavenBase}</dd>
            </div>
            <div>
              <dt className="text-zinc-500">nexus.credentialsId</dt>
              <dd className="font-mono">{settings.nexusCredentialsId}</dd>
            </div>
            <div>
              <dt className="text-zinc-500">updatedAt</dt>
              <dd>{new Date(settings.updatedAt).toLocaleString()}</dd>
            </div>
          </dl>
        )
      )}
    </div>
  );
}
