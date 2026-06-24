import { FormEvent, useEffect, useState, type ReactNode } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import type { GPProfileSlot, GPRelease } from "../api/types";
import { api, getActor, setActor } from "../lib/api";
import { isCanonicalProfile, SLOT_LABELS, sortProfileSlots } from "../lib/gpSlots";

type SlotVersions = Record<string, string[]>;
type Tab = "draft" | "publish" | "promote";

type PublishWizardProps = {
  scopedGpName?: string;
  lockedTab?: "draft" | "publish";
};

function releasePath(gp: string, ver: string) {
  return `/gp/${encodeURIComponent(gp)}/releases/${encodeURIComponent(ver)}`;
}

function stripSnapshot(version: string): string {
  const idx = version.indexOf("-snapshot.");
  return idx >= 0 ? version.slice(0, idx) : version;
}

function nextSnapshotVersion(drafts: GPRelease[], base: string): string {
  const cleanBase = stripSnapshot(base);
  const prefix = `${cleanBase}-snapshot.`;
  let maxN = 0;
  for (const d of drafts) {
    if (d.version.startsWith(prefix)) {
      const n = parseInt(d.version.slice(prefix.length), 10);
      if (!Number.isNaN(n) && n > maxN) maxN = n;
    }
  }
  return `${cleanBase}-snapshot.${maxN + 1}`;
}

export default function PublishWizard({ scopedGpName, lockedTab }: PublishWizardProps = {}) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const isScoped = Boolean(scopedGpName);
  const [tab, setTab] = useState<Tab>(lockedTab ?? "draft");
  const [gpNames, setGpNames] = useState<string[]>([]);
  const [gpName, setGpName] = useState(scopedGpName ?? searchParams.get("name") ?? "");
  const [slots, setSlots] = useState<GPProfileSlot[]>([]);
  const [composition, setComposition] = useState<Record<string, string>>({});
  const [versionOptions, setVersionOptions] = useState<SlotVersions>({});
  const [version, setVersion] = useState("");
  const [baseVersion, setBaseVersion] = useState("1.0.0");
  const [drafts, setDrafts] = useState<GPRelease[]>([]);
  const [promoteVersion, setPromoteVersion] = useState("");
  const [actor, setActorField] = useState(getActor());
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    if (scopedGpName) {
      setGpName(scopedGpName);
      return;
    }
    api
      .gpNames()
      .then((r) => {
        setGpNames(r.items);
        if (r.items.length === 0) {
          setGpName("");
        } else {
          const fromUrl = searchParams.get("name");
          setGpName((prev) => {
            if (fromUrl && r.items.includes(fromUrl)) return fromUrl;
            if (prev && r.items.includes(prev)) return prev;
            return r.items[0];
          });
        }
      })
      .catch((err: Error) => setError(err.message));
  }, [scopedGpName]);

  useEffect(() => {
    if (lockedTab) setTab(lockedTab);
  }, [lockedTab]);

  useEffect(() => {
    if (!gpName) {
      setLoading(false);
      setSlots([]);
      setComposition({});
      setVersionOptions({});
      return;
    }
    setLoading(true);
    setError(null);
    setSuccess(null);

    Promise.all([api.gpProfile(gpName), api.gpReleases(gpName, true)])
      .then(async ([profile, releases]) => {
        setSlots(sortProfileSlots(profile.slots));
        setDrafts(releases.items.filter((r) => r.status === "draft"));

        const published = releases.items.filter((r) => r.status === "published");
        const latest = published[0];
        let defaults: Record<string, string> = {};
        if (latest) {
          const detail = await api.gpRelease(gpName, latest.version);
          for (const c of detail.composition) {
            const slot = profile.slots.find((s) => s.type === c.type && s.name === c.name);
            if (slot) defaults[slot.key] = c.version;
          }
          if (!baseVersion || baseVersion === "1.0.0") {
            setBaseVersion(latest.version);
          }
        }

        const versions: SlotVersions = {};
        await Promise.all(
          profile.slots.map(async (slot) => {
            const r = await api.componentVersions(slot.type, slot.name);
            versions[slot.key] = r.items
              .filter((v) => v.status === "published")
              .map((v) => v.version);
            if (!defaults[slot.key] && r.items.length > 0) {
              defaults[slot.key] = r.items[0].version;
            }
          }),
        );

        setVersionOptions(versions);
        setComposition(defaults);

        const draftItems = releases.items.filter((r) => r.status === "draft");
        setDrafts(draftItems);
        if (tab === "draft") {
          setVersion(nextSnapshotVersion(draftItems, baseVersion || "1.0.0"));
        }
        if (draftItems.length > 0 && !promoteVersion) {
          setPromoteVersion(draftItems[0].version);
        }
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [gpName]);

  useEffect(() => {
    if (tab === "draft" && baseVersion) {
      setVersion(nextSnapshotVersion(drafts, baseVersion));
    }
  }, [tab, baseVersion, drafts]);

  async function onSubmitDraft(e: FormEvent) {
    e.preventDefault();
    if (!version.trim()) {
      setError("Version обязательна");
      return;
    }
    setSubmitting(true);
    setError(null);
    setSuccess(null);
    setActor(actor);

    try {
      const result = await api.createDraftGPRelease(gpName, {
        version: version.trim(),
        composition,
        actor: actor.trim() || undefined,
      });
      setSuccess(`Draft ${result.name}@${result.version} создан`);
      setTimeout(() => {
        navigate(releasePath(result.name, result.version));
      }, 1200);
    } catch (err) {
      setError(err instanceof Error ? err.message : "create draft failed");
    } finally {
      setSubmitting(false);
    }
  }

  async function onSubmitPublish(e: FormEvent) {
    e.preventDefault();
    if (!version.trim()) {
      setError("Version обязательна");
      return;
    }
    setSubmitting(true);
    setError(null);
    setSuccess(null);
    setActor(actor);

    try {
      const result = await api.publishGPRelease(gpName, {
        version: version.trim(),
        composition,
        actor: actor.trim() || undefined,
      });
      setSuccess(`${result.name}@${result.version} опубликован`);
      setTimeout(() => {
        navigate(releasePath(result.name, result.version));
      }, 1200);
    } catch (err) {
      setError(err instanceof Error ? err.message : "publish failed");
    } finally {
      setSubmitting(false);
    }
  }

  async function onPromote() {
    if (!promoteVersion) {
      setError("Выберите draft");
      return;
    }
    setSubmitting(true);
    setError(null);
    setSuccess(null);
    setActor(actor);

    try {
      const result = await api.promoteDraftGPRelease(
        gpName,
        promoteVersion,
        actor.trim() || undefined,
      );
      setSuccess(`${result.name}@${result.version} promoted → published`);
      setTimeout(() => {
        navigate(releasePath(result.name, result.version));
      }, 1200);
    } catch (err) {
      setError(err instanceof Error ? err.message : "promote failed");
    } finally {
      setSubmitting(false);
    }
  }

  const backTo = isScoped
    ? `/gp/${encodeURIComponent(gpName)}/releases`
    : "/gp";

  return (
    <div className="space-y-6">
      <div>
        <Link to={backTo} className="text-sm text-sky-400 hover:underline">
          ← {isScoped ? "Releases" : "GP Profiles"}
        </Link>
        <h1 className="mt-2 text-2xl font-semibold">
          {lockedTab === "draft"
            ? "New draft snapshot"
            : lockedTab === "publish"
              ? "New release"
              : "Publish GP release"}
        </h1>
        <p className="mt-1 text-zinc-400">
          {isScoped
            ? lockedTab === "draft"
              ? `Draft snapshot для ${gpName}`
              : `Stable release для ${gpName}`
            : "Draft snapshots, direct publish или promote draft"}
        </p>
      </div>

      {!isScoped && (
      <div className="flex gap-2 border-b border-zinc-800">
        {(
          [
            ["draft", "Create draft"],
            ["publish", "Publish stable"],
            ["promote", "Promote draft"],
          ] as const
        ).map(([id, label]) => (
          <button
            key={id}
            type="button"
            onClick={() => {
              setTab(id);
              setError(null);
              setSuccess(null);
              if (id === "publish") setVersion("");
            }}
            className={`border-b-2 px-4 py-2 text-sm ${
              tab === id
                ? "border-sky-500 text-sky-400"
                : "border-transparent text-zinc-500 hover:text-zinc-300"
            }`}
          >
            {label}
          </button>
        ))}
      </div>
      )}

      {error && (
        <p className="rounded border border-red-900/50 bg-red-950/30 px-4 py-3 text-red-400">
          {error}
        </p>
      )}
      {success && (
        <p className="rounded border border-emerald-900/50 bg-emerald-950/30 px-4 py-3 text-emerald-400">
          {success}
        </p>
      )}

      {gpNames.length === 0 && (
        <div className="rounded-lg border border-amber-900/40 bg-amber-950/20 px-4 py-4 text-sm text-amber-200">
          <p>Нет Golden Path на платформе.</p>
          <p className="mt-2">
            <Link to="/gp/new" className="font-medium text-sky-400 hover:underline">
              Создайте GP profile
            </Link>{" "}
            — затем вернитесь к publish.
          </p>
        </div>
      )}

      {loading ? (
        <p className="text-zinc-500">Загрузка profile…</p>
      ) : gpName ? (
        <>
          {!isScoped && <GpSelector gpNames={gpNames} gpName={gpName} onChange={setGpName} />}
          {!isCanonicalProfile(slots) && (
            <p className="rounded border border-amber-500/40 bg-amber-950/30 px-4 py-3 text-sm text-amber-200">
              Профиль использует устаревшие slots (не 4-component model). Создайте новый GP через{" "}
              <Link to="/gp/new" className="text-sky-400 hover:underline">
                Новый GP
              </Link>
              .
            </p>
          )}
          <CompositionSection
            slots={slots}
            composition={composition}
            versionOptions={versionOptions}
            onChange={setComposition}
          />

          {tab === "draft" && (
            <form onSubmit={onSubmitDraft} className="space-y-6">
              <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
                <h2 className="font-medium">Draft snapshot</h2>
                <p className="mt-1 text-sm text-zinc-500">
                  Версия вида <span className="font-mono">1.0.0-snapshot.N</span> — редактируемый
                  draft с копией artifacts из latest published
                </p>
                <div className="mt-4 grid gap-4 sm:grid-cols-2">
                  <Field label="Base version">
                    <input
                      value={baseVersion}
                      onChange={(e) => setBaseVersion(e.target.value)}
                      placeholder="1.0.0"
                      className={inputClass}
                    />
                  </Field>
                  <Field label="Snapshot version *">
                    <input
                      value={version}
                      onChange={(e) => setVersion(e.target.value)}
                      className={inputClass}
                      required
                    />
                  </Field>
                  <Field label="Actor">
                    <input
                      value={actor}
                      onChange={(e) => setActorField(e.target.value)}
                      placeholder="platform-team"
                      className={inputClass}
                    />
                  </Field>
                </div>
              </section>
              <SubmitRow submitting={submitting} label="Create draft" />
            </form>
          )}

          {tab === "publish" && (
            <form onSubmit={onSubmitPublish} className="space-y-6">
              <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
                <h2 className="font-medium">Stable release</h2>
                <p className="mt-1 text-sm text-zinc-500">
                  Append-only publish — сразу в Nexus + обновление pointers
                </p>
                <div className="mt-4 grid gap-4 sm:grid-cols-2">
                  <Field label="Version *">
                    <input
                      value={version}
                      onChange={(e) => setVersion(e.target.value)}
                      placeholder="1.0.1"
                      className={inputClass}
                      required
                    />
                  </Field>
                  <Field label="Actor">
                    <input
                      value={actor}
                      onChange={(e) => setActorField(e.target.value)}
                      placeholder="platform-team"
                      className={inputClass}
                    />
                  </Field>
                </div>
              </section>
              <SubmitRow submitting={submitting} label="Publish" />
            </form>
          )}

          {tab === "promote" && (
            <section className="space-y-6 rounded-lg border border-zinc-800 bg-zinc-900 p-6">
              <h2 className="font-medium">Promote draft → published</h2>
              <p className="text-sm text-zinc-500">
                Draft публикуется в Nexus; wildcards (~, ^, *) обновляются
              </p>
              {drafts.length === 0 ? (
                <p className="text-zinc-500">Нет draft releases для {gpName}</p>
              ) : (
                <div className="grid gap-4 sm:grid-cols-2">
                  <Field label="Draft">
                    <select
                      value={promoteVersion}
                      onChange={(e) => setPromoteVersion(e.target.value)}
                      className={inputClass}
                    >
                      {drafts.map((d) => (
                        <option key={d.version} value={d.version}>
                          {d.version}
                        </option>
                      ))}
                    </select>
                  </Field>
                  <Field label="Actor">
                    <input
                      value={actor}
                      onChange={(e) => setActorField(e.target.value)}
                      placeholder="platform-team"
                      className={inputClass}
                    />
                  </Field>
                </div>
              )}
              <div className="flex gap-3">
                <button
                  type="button"
                  disabled={submitting || drafts.length === 0}
                  onClick={onPromote}
                  className="rounded bg-sky-600 px-5 py-2.5 text-sm font-medium hover:bg-sky-500 disabled:opacity-50"
                >
                  {submitting ? "Promoting…" : "Promote"}
                </button>
                {promoteVersion && (
                  <Link
                    to={releasePath(gpName, promoteVersion)}
                    className="rounded border border-zinc-700 px-5 py-2.5 text-sm text-zinc-300 hover:bg-zinc-800"
                  >
                    View draft
                  </Link>
                )}
              </div>
            </section>
          )}
        </>
      ) : gpNames.length > 0 ? (
        <p className="text-zinc-500">Выберите Golden path.</p>
      ) : null}
    </div>
  );
}

function GpSelector({
  gpNames,
  gpName,
  onChange,
}: {
  gpNames: string[];
  gpName: string;
  onChange: (v: string) => void;
}) {
  return (
    <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <Field label="Golden path">
          <select value={gpName} onChange={(e) => onChange(e.target.value)} className={inputClass}>
            {gpNames.map((n) => (
              <option key={n} value={n}>
                {n}
              </option>
            ))}
          </select>
        </Field>
        <Link
          to="/gp/new"
          className="rounded-lg border border-sky-500/70 bg-sky-950/50 px-4 py-2 text-sm font-semibold text-sky-300 hover:border-sky-400 hover:bg-sky-900/60"
        >
          + Новый GP
        </Link>
      </div>
    </section>
  );
}

function CompositionSection({
  slots,
  composition,
  versionOptions,
  onChange,
}: {
  slots: GPProfileSlot[];
  composition: Record<string, string>;
  versionOptions: SlotVersions;
  onChange: (v: Record<string, string>) => void;
}) {
  return (
    <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
      <h2 className="font-medium">Composition (4 runtime pins)</h2>
      <p className="mt-1 text-sm text-zinc-500">
        jnlp · agent · executor · lib · gp-content. Prefill из latest published.
      </p>
      <div className="mt-4 space-y-4">
        {slots.map((slot) => (
          <div
            key={slot.key}
            className="grid gap-2 sm:grid-cols-[9rem_1fr_1fr] sm:items-center"
          >
            <span className="font-mono text-sm text-sky-400">{slot.key}</span>
            <span className="text-sm text-zinc-500">
              {SLOT_LABELS[slot.key] ?? `${slot.type}/${slot.name}`}
              <span className="mt-0.5 block font-mono text-xs text-zinc-600">
                {slot.type}/{slot.name}
              </span>
            </span>
            <select
              value={composition[slot.key] ?? ""}
              onChange={(e) => onChange({ ...composition, [slot.key]: e.target.value })}
              className={inputClass}
              required
            >
              {(versionOptions[slot.key] ?? []).map((v) => (
                <option key={v} value={v}>
                  {v}
                </option>
              ))}
            </select>
          </div>
        ))}
      </div>
    </section>
  );
}

function SubmitRow({ submitting, label }: { submitting: boolean; label: string }) {
  return (
    <div className="flex gap-3">
      <button
        type="submit"
        disabled={submitting}
        className="rounded bg-sky-600 px-5 py-2.5 text-sm font-medium hover:bg-sky-500 disabled:opacity-50"
      >
        {submitting ? "…" : label}
      </button>
      <Link
        to="/audit"
        className="rounded border border-zinc-700 px-5 py-2.5 text-sm text-zinc-300 hover:bg-zinc-800"
      >
        Audit log
      </Link>
    </div>
  );
}

function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <label className="block">
      <span className="text-xs text-zinc-500">{label}</span>
      <div className="mt-1">{children}</div>
    </label>
  );
}

const inputClass =
  "w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm focus:border-sky-600 focus:outline-none";
