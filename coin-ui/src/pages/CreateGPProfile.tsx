import { FormEvent, useCallback, useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import type { Component } from "../api/types";
import { api, getActor, setActor } from "../lib/api";
import {
  CANONICAL_SLOT_ORDER,
  SLOT_LABELS,
  SLOT_PICKER_SPECS,
  slotsToProfile,
} from "../lib/gpSlots";

const inputClass =
  "w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm focus:border-sky-600 focus:outline-none";

type SlotRow = {
  key: string;
  type: string;
  componentName: string;
  version: string;
};

function componentRegistered(items: Component[], type: string, name: string): boolean {
  return items.some((c) => c.type === type && c.name === name);
}

async function publishedVersions(type: string, name: string): Promise<string[]> {
  const r = await api.componentVersionsOptional(type, name);
  return r.items.filter((v) => v.status === "published").map((v) => v.version);
}

function buildInitialSlots(
  namesByType: Record<string, string[]>,
): SlotRow[] {
  return SLOT_PICKER_SPECS.map((spec) => ({
    key: spec.key,
    type: spec.type,
    componentName: spec.fixedName ?? namesByType[spec.key]?.[0] ?? "",
    version: "",
  }));
}

export default function CreateGPProfile() {
  const navigate = useNavigate();
  const gpVersion = "0.0.1";
  const [name, setName] = useState("");
  const [slots, setSlots] = useState<SlotRow[]>(() => buildInitialSlots({}));
  const [componentNames, setComponentNames] = useState<Record<string, string[]>>({});
  const [versionOptions, setVersionOptions] = useState<Record<string, string[]>>({});
  const [registered, setRegistered] = useState<Component[]>([]);
  const [actor, setActorField] = useState(getActor());
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadVersionsForSlot = useCallback(async (type: string, componentName: string, key: string) => {
    if (!componentName) {
      setVersionOptions((prev) => ({ ...prev, [key]: [] }));
      return [] as string[];
    }
    const versions = await publishedVersions(type, componentName);
    setVersionOptions((prev) => ({ ...prev, [key]: versions }));
    return versions;
  }, []);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const { items } = await api.components();
        const namesByType: Record<string, string[]> = {};
        for (const spec of SLOT_PICKER_SPECS) {
          if (spec.fixedName) continue;
          namesByType[spec.key] = items
            .filter((c: Component) => c.type === spec.type)
            .map((c) => c.name)
            .filter((n) => !(spec.excludeNames ?? []).includes(n))
            .sort();
        }

        const initial = buildInitialSlots(namesByType);
        const versions: Record<string, string[]> = {};

        await Promise.all(
          SLOT_PICKER_SPECS.map(async (spec) => {
            const componentName = spec.fixedName ?? namesByType[spec.key]?.[0] ?? "";
            const vers = await publishedVersions(spec.type, componentName);
            versions[spec.key] = vers;
            const row = initial.find((s) => s.key === spec.key);
            if (row) {
              row.componentName = componentName;
              row.version = vers[0] ?? "";
            }
          }),
        );

        if (cancelled) return;
        setRegistered(items);
        setComponentNames(namesByType);
        setVersionOptions(versions);
        setSlots(initial);
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Ошибка загрузки компонентов");
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    void load();
    return () => {
      cancelled = true;
    };
  }, []);

  async function onComponentChange(key: string, type: string, componentName: string) {
    setError(null);
    const versions = await loadVersionsForSlot(type, componentName, key);
    setSlots((prev) =>
      prev.map((s) =>
        s.key === key ? { ...s, componentName, version: versions[0] ?? "" } : s,
      ),
    );
  }

  function onVersionChange(key: string, version: string) {
    setSlots((prev) => prev.map((s) => (s.key === key ? { ...s, version } : s)));
  }

  function slotHint(key: string, slot: SlotRow, spec: (typeof SLOT_PICKER_SPECS)[number]): string | null {
    const componentName = spec.fixedName ?? slot.componentName;
    if (!componentName) {
      return "Нет зарегистрированных компонентов этого типа";
    }
    if (!componentRegistered(registered, spec.type, componentName)) {
      return `Компонент ${spec.type}/${componentName} не в registry — опубликуйте platform job`;
    }
    if ((versionOptions[key] ?? []).length === 0) {
      return "Нет published версий";
    }
    return null;
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    const trimmed = name.trim();
    if (!trimmed) {
      setError("Имя GP обязательно");
      return;
    }
    for (const slot of slots) {
      if (!slot.componentName || !slot.version) {
        setError(`Выберите версию для слота ${slot.key}`);
        return;
      }
    }

    const composition = Object.fromEntries(slots.map((s) => [s.key, s.version]));
    const profileSlots = slotsToProfile(slots);

    setSubmitting(true);
    setActor(actor);
    try {
      await api.createGPProfile({
        name: trimmed,
        slots: profileSlots,
        actor: actor.trim() || undefined,
      });
      await api.publishGPRelease(trimmed, {
        version: gpVersion,
        composition,
        actor: actor.trim() || undefined,
      });
      navigate(`/releases/${encodeURIComponent(trimmed)}/${encodeURIComponent(gpVersion)}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка создания GP");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div>
        <Link to="/releases" className="text-sm text-sky-400 hover:underline">
          ← GP releases
        </Link>
        <h1 className="mt-2 text-2xl font-semibold">Новый Golden Path</h1>
        <p className="mt-1 text-sm text-zinc-400">
          Профиль и первый release: выберите зарегистрированные версии для 4 runtime-компонентов
          (jnlp, agent, executor, lib, gp-content).
        </p>
      </div>

      <form onSubmit={onSubmit} className="space-y-4 rounded-lg border border-zinc-800 bg-zinc-900/50 p-6">
        <div className="grid gap-4 sm:grid-cols-2">
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
            <span className="text-sm text-zinc-300">Версия release</span>
            <input
              className={`${inputClass} cursor-not-allowed opacity-70`}
              value={gpVersion}
              readOnly
              disabled
              aria-readonly="true"
            />
            <p className="text-xs text-zinc-500">Первый release нового GP всегда 0.0.1</p>
          </label>
        </div>

        <section className="rounded border border-zinc-800 bg-zinc-950/60 p-4">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <h2 className="text-sm font-medium text-zinc-200">Composition (4 slots)</h2>
            <Link to="/components" className="text-xs text-sky-400 hover:underline">
              Registry компонентов →
            </Link>
          </div>
          <p className="mt-1 text-xs text-zinc-500">
            Только опубликованные версии из registry платформы.
          </p>
          {loading ? (
            <p className="mt-4 text-sm text-zinc-500">Загрузка компонентов…</p>
          ) : (
            <div className="mt-4 space-y-3">
              {CANONICAL_SLOT_ORDER.map((key) => {
                const slot = slots.find((s) => s.key === key);
                const spec = SLOT_PICKER_SPECS.find((s) => s.key === key);
                if (!slot || !spec) return null;
                const vers = versionOptions[key] ?? [];
                const names = componentNames[key];
                const hint = slotHint(key, slot, spec);
                return (
                  <div
                    key={key}
                    className="rounded border border-zinc-800/80 bg-zinc-900/40 p-3"
                  >
                    <div className="grid gap-2 sm:grid-cols-[8rem_1fr_1fr] sm:items-center">
                      <div>
                        <span className="font-mono text-sm text-sky-400">{key}</span>
                        <p className="text-xs text-zinc-500">{SLOT_LABELS[key]}</p>
                      </div>
                      {spec.fixedName ? (
                        <span className="font-mono text-sm text-zinc-300">
                          {spec.type}/{spec.fixedName}
                        </span>
                      ) : (
                        <select
                          className={inputClass}
                          value={slot.componentName}
                          onChange={(e) => onComponentChange(key, spec.type, e.target.value)}
                          required
                        >
                          {(names ?? []).length === 0 ? (
                            <option value="">— нет компонентов —</option>
                          ) : (
                            (names ?? []).map((n) => (
                              <option key={n} value={n}>
                                {spec.type}/{n}
                              </option>
                            ))
                          )}
                        </select>
                      )}
                      <select
                        className={inputClass}
                        value={slot.version}
                        onChange={(e) => onVersionChange(key, e.target.value)}
                        required
                        disabled={vers.length === 0}
                      >
                        {vers.length === 0 ? (
                          <option value="">— выберите версию —</option>
                        ) : (
                          vers.map((v) => (
                            <option key={v} value={v}>
                              {v}
                            </option>
                          ))
                        )}
                      </select>
                    </div>
                    {hint && <p className="mt-2 text-xs text-amber-400/90">{hint}</p>}
                  </div>
                );
              })}
            </div>
          )}
        </section>

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
          disabled={submitting || loading}
          className="rounded bg-sky-600 px-4 py-2 text-sm font-medium text-white hover:bg-sky-500 disabled:opacity-50"
        >
          {submitting ? "Создание…" : "Создать GP и опубликовать release"}
        </button>
      </form>
    </div>
  );
}
