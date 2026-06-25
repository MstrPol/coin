import { FormEvent, useEffect, useState, type ReactNode } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import type { GPRelease } from "../api/types";
import { api, getActor, setActor } from "../lib/api";
import {
  defaultAgentStackName,
  defaultBranchingModelForGP,
  GP_DRAFT_SLOT_ORDER,
} from "../lib/gpSlots";
import GpCompositionForm from "../components/GpCompositionForm";

type SlotVersions = Record<string, string[]>;
type Tab = "draft" | "promote";

type PublishWizardProps = {
  scopedGpName?: string;
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

function publishedVersions(
  items: { version: string; status: string }[],
): string[] {
  return items.filter((v) => v.status === "published").map((v) => v.version);
}

export default function PublishWizard({ scopedGpName }: PublishWizardProps = {}) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const isScoped = Boolean(scopedGpName);
  const [tab, setTab] = useState<Tab>("draft");
  const [gpNames, setGpNames] = useState<string[]>([]);
  const [gpName, setGpName] = useState(scopedGpName ?? searchParams.get("name") ?? "");
  const [branchingModelName, setBranchingModelName] = useState("");
  const [branchingModelOptions, setBranchingModelOptions] = useState<string[]>([]);
  const [gpContentName, setGpContentName] = useState("");
  const [gpContentOptions, setGpContentOptions] = useState<string[]>([]);
  const [agentStackName, setAgentStackName] = useState(defaultAgentStackName());
  const [agentStackOptions, setAgentStackOptions] = useState<string[]>([]);
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
  const [catalogEmpty, setCatalogEmpty] = useState(false);

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
  }, [scopedGpName, searchParams]);

  useEffect(() => {
    if (!gpName) {
      setLoading(false);
      setComposition({});
      setVersionOptions({});
      setBranchingModelName("");
      setBranchingModelOptions([]);
      setGpContentName("");
      setGpContentOptions([]);
      setAgentStackName(defaultAgentStackName());
      setAgentStackOptions([]);
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(null);

    const defaultBm = defaultBranchingModelForGP(gpName);

    Promise.all([api.components(), api.gpReleases(gpName, true)])
      .then(async ([components, releases]) => {
        const bmNames = components.items
          .filter((c) => c.type === "branching-model")
          .map((c) => c.name)
          .sort();
        setBranchingModelOptions(bmNames);

        const gcNames = components.items
          .filter((c) => c.type === "gp-content")
          .map((c) => c.name)
          .sort();
        setGpContentOptions(gcNames);
        setCatalogEmpty(gcNames.length === 0);

        const agentNames = components.items
          .filter((c) => c.type === "agent")
          .map((c) => c.name)
          .sort();
        setAgentStackOptions(agentNames);

        const published = releases.items.filter((r) => r.status === "published");
        const draftItems = releases.items.filter((r) => r.status === "draft");
        setDrafts(draftItems);

        let bmName = defaultBm;
        let gcName = "";
        let agentName = defaultAgentStackName();
        let defaults: Record<string, string> = {};

        const latest = published[0] ?? draftItems[0];
        if (latest) {
          const detail = await api.gpRelease(gpName, latest.version);
          for (const c of detail.composition) {
            if (c.type === "agent") {
              defaults.agent = c.version;
              agentName = c.name;
            }
            if (c.type === "gp-content") {
              defaults["gp-content"] = c.version;
              gcName = c.name;
            }
            if (c.type === "branching-model") {
              defaults["branching-model"] = c.version;
              bmName = c.name;
            }
          }
          if (published[0] && (!baseVersion || baseVersion === "1.0.0")) {
            setBaseVersion(published[0].version);
          }
        }

        if (!gcName) {
          if (gcNames.includes(gpName)) gcName = gpName;
          else if (gcNames.length > 0) gcName = gcNames[0];
        }
        if (!bmNames.includes(bmName) && bmNames.length > 0) {
          bmName = bmNames.includes(defaultBm) ? defaultBm : bmNames[0];
        }
        if (!agentNames.includes(agentName) && agentNames.length > 0) {
          agentName = agentNames.includes(defaultAgentStackName())
            ? defaultAgentStackName()
            : agentNames[0];
        }
        setGpContentName(gcName);
        setBranchingModelName(bmName);
        setAgentStackName(agentName);

        const versions: SlotVersions = {};
        if (agentName) {
          const agentR = await api.componentVersions("agent", agentName);
          versions.agent = publishedVersions(agentR.items);
          if (!defaults.agent && versions.agent.length > 0) {
            defaults.agent = versions.agent[0];
          }
        } else {
          versions.agent = [];
        }
        if (gcName) {
          const gcR = await api.componentVersionsOptional("gp-content", gcName);
          versions["gp-content"] = publishedVersions(gcR?.items ?? []);
          if (!defaults["gp-content"] && versions["gp-content"].length > 0) {
            defaults["gp-content"] = versions["gp-content"][0];
          }
        } else {
          versions["gp-content"] = [];
        }
        if (bmName) {
          const bmR = await api.componentVersions("branching-model", bmName);
          versions["branching-model"] = publishedVersions(bmR.items);
          if (!defaults["branching-model"] && versions["branching-model"].length > 0) {
            defaults["branching-model"] = versions["branching-model"][0];
          }
        }

        setVersionOptions(versions);
        setComposition(defaults);
        setVersion(nextSnapshotVersion(draftItems, baseVersion || "1.0.0"));
        if (draftItems.length > 0 && !promoteVersion) {
          setPromoteVersion(draftItems[0].version);
        }
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [gpName]);

  useEffect(() => {
    if (!gpName || !branchingModelName) return;
    api
      .componentVersions("branching-model", branchingModelName)
      .then((r) => {
        const vers = r.items.filter((v) => v.status === "published").map((v) => v.version);
        setVersionOptions((prev) => ({ ...prev, "branching-model": vers }));
        setComposition((prev) => {
          if (prev["branching-model"] && vers.includes(prev["branching-model"])) {
            return prev;
          }
          return { ...prev, "branching-model": vers[0] ?? "" };
        });
      })
      .catch(() => {});
  }, [gpName, branchingModelName]);

  useEffect(() => {
    if (!gpContentName) return;
    api
      .componentVersionsOptional("gp-content", gpContentName)
      .then((r) => {
        const vers = publishedVersions(r?.items ?? []);
        setVersionOptions((prev) => ({ ...prev, "gp-content": vers }));
        setComposition((prev) => {
          if (prev["gp-content"] && vers.includes(prev["gp-content"])) return prev;
          return { ...prev, "gp-content": vers[0] ?? "" };
        });
      })
      .catch(() => {});
  }, [gpContentName]);

  useEffect(() => {
    if (!agentStackName) return;
    api
      .componentVersions("agent", agentStackName)
      .then((r) => {
        const vers = publishedVersions(r.items);
        setVersionOptions((prev) => ({ ...prev, agent: vers }));
        setComposition((prev) => {
          if (prev.agent && vers.includes(prev.agent)) return prev;
          return { ...prev, agent: vers[0] ?? "" };
        });
      })
      .catch(() => {});
  }, [agentStackName]);

  useEffect(() => {
    if (tab === "draft" && baseVersion) {
      setVersion(nextSnapshotVersion(drafts, baseVersion));
    }
  }, [tab, baseVersion, drafts]);

  const draftBlocked =
    catalogEmpty ||
    !agentStackName ||
    !gpContentName ||
    !branchingModelName ||
    GP_DRAFT_SLOT_ORDER.some((key) => !(versionOptions[key] ?? []).length);

  async function onSubmitDraft(e: FormEvent) {
    e.preventDefault();
    if (!version.trim()) {
      setError("Version обязательна");
      return;
    }
    if (catalogEmpty) {
      setError("Нет gp-content в registry — опубликуйте build stack");
      return;
    }
    if (!gpContentName) {
      setError("Выберите gp-content");
      return;
    }
    if (!branchingModelName) {
      setError("Выберите branching model");
      return;
    }
    if (!agentStackName) {
      setError("Выберите agent stack");
      return;
    }
    for (const key of GP_DRAFT_SLOT_ORDER) {
      if (!composition[key]) {
        setError(`Выберите версию для ${key}`);
        return;
      }
    }

    setSubmitting(true);
    setError(null);
    setSuccess(null);
    setActor(actor);

    try {
      const result = await api.createDraftGPRelease(gpName, {
        version: version.trim(),
        composition,
        agentStackName,
        gpContentName,
        branchingModelName,
        actor: actor.trim() || undefined,
      });
      setSuccess(`Draft ${result.name}@${result.version} создан`);
      setTimeout(() => {
        navigate(releasePath(result.name, result.version));
      }, 1200);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "create draft failed";
      if (msg.includes("gp-content") || msg.includes("gpContentName")) {
        setError("Выберите опубликованный gp-content и версию из каталога");
      } else if (msg.includes("component")) {
        setError(msg);
      } else {
        setError(msg);
      }
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

  const noGpAvailable = !isScoped && gpNames.length === 0;

  return (
    <div className="space-y-6">
      <div>
        <Link to={backTo} className="text-sm text-sky-400 hover:underline">
          ← {isScoped ? "Releases" : "GP Profiles"}
        </Link>
        <h1 className="mt-2 text-2xl font-semibold">
          {isScoped ? "New draft snapshot" : "GP draft & promote"}
        </h1>
        <p className="mt-1 text-zinc-400">
          {isScoped
            ? `Draft snapshot для ${gpName} — agent + gp-content + branching-model`
            : "Создайте draft snapshot или promote draft → published"}
        </p>
      </div>

      {!isScoped && (
        <div className="flex gap-2 border-b border-zinc-800">
          {(
            [
              ["draft", "Create draft"],
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

      {noGpAvailable && (
        <div className="rounded-lg border border-amber-900/40 bg-amber-950/20 px-4 py-4 text-sm text-amber-200">
          <p>Нет Golden Path на платформе.</p>
          <p className="mt-2">
            <Link to="/gp/new" className="font-medium text-sky-400 hover:underline">
              Создайте GP profile
            </Link>{" "}
            — затем вернитесь к draft.
          </p>
        </div>
      )}

      {loading ? (
        <p className="text-zinc-500">Загрузка…</p>
      ) : gpName ? (
        <>
          {!isScoped && <GpSelector gpNames={gpNames} gpName={gpName} onChange={setGpName} />}

          {catalogEmpty && (
            <div className="rounded-lg border border-amber-900/40 bg-amber-950/20 px-4 py-4 text-sm text-amber-200">
              <p>Нет gp-content в platform registry.</p>
              <p className="mt-2 text-zinc-400">
                Platform team публикует build stacks до создания GP draft.
              </p>
              <p className="mt-2">
                <Link to="/platform/build-stacks" className="text-sky-400 hover:underline">
                  Build stacks →
                </Link>
              </p>
            </div>
          )}

          <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
            <h2 className="font-medium">GP composition (3 pins)</h2>
            <p className="mt-1 text-sm text-zinc-500">
              Agent stack, gp-content и branching-model — per GP version.
            </p>
            <GpCompositionForm
              agentStackName={agentStackName}
              agentStackOptions={agentStackOptions}
              onAgentStackChange={setAgentStackName}
              gpContentName={gpContentName}
              gpContentOptions={gpContentOptions}
              onGpContentChange={setGpContentName}
              branchingModelName={branchingModelName}
              branchingModelOptions={branchingModelOptions}
              onBranchingModelChange={setBranchingModelName}
              composition={composition}
              versionOptions={versionOptions}
              onCompositionChange={setComposition}
            />
          </section>

          {(isScoped || tab === "draft") && (
            <form onSubmit={onSubmitDraft} className="space-y-6">
              <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6">
                <h2 className="font-medium">Draft snapshot</h2>
                <p className="mt-1 text-sm text-zinc-500">
                  Версия вида <span className="font-mono">1.0.0-snapshot.N</span> — редактируемый
                  draft; stable release только через promote
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
              <SubmitRow submitting={submitting} label="Create draft" disabled={draftBlocked} />
            </form>
          )}

          {!isScoped && tab === "promote" && (
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
      ) : !noGpAvailable ? (
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

function SubmitRow({
  submitting,
  label,
  disabled,
}: {
  submitting: boolean;
  label: string;
  disabled?: boolean;
}) {
  return (
    <div className="flex gap-3">
      <button
        type="submit"
        disabled={submitting || disabled}
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
