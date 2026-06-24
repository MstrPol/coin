import { FormEvent, useCallback, useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import BranchingModelEditor from "../components/studio/BranchingModelEditor";
import GpContentEditor from "../components/studio/GpContentEditor";
import PilotPromotePanel from "../components/studio/PilotPromotePanel";
import type { ComponentVersionDetail, ValidateComponentPackageResult } from "../api/types";
import { useAuth } from "../context/AuthContext";
import { api, getActor } from "../lib/api";
import {
  buildManifestSubset,
  defaultBranchingModel,
  parseBranchingModelYaml,
  serializeBranchingModel,
  validateBranchingModelClient,
  type BranchingModel,
} from "../lib/branchingModelYaml";
import {
  buildGpContentManifestSubset,
  defaultContainerfile,
  defaultGpContent,
  GP_CONTAINERFILE_ARTIFACT,
  GP_CONTENT_ARTIFACT,
  parseGpContentYaml,
  serializeGpContent,
  validateGpContentClient,
  type GpContentModel,
} from "../lib/gpContentYaml";
import {
  isStudioType,
  STUDIO_COMPONENT_TYPES,
  studioTypeConfig,
  usesPGOnlyCanaryRegistry,
} from "../lib/componentStudio";

function statusBadge(status: string): string {
  switch (status) {
    case "draft":
      return "text-amber-400";
    case "canary":
      return "text-sky-400";
    case "published":
      return "text-emerald-400";
    default:
      return "text-zinc-400";
  }
}

export default function ComponentStudio() {
  const { type: typeParam, name: nameParam, version: versionParam } = useParams();
  const navigate = useNavigate();
  const { can } = useAuth();

  if (typeParam && nameParam && versionParam) {
    return (
      <StudioEditor
        type={typeParam}
        name={nameParam}
        version={versionParam}
        canEdit={can("publisher")}
      />
    );
  }

  return <StudioHome canEdit={can("publisher")} onOpen={(t, n, v) => navigate(`/studio/${t}/${n}/${encodeURIComponent(v)}`)} />;
}

function StudioHome({
  canEdit,
  onOpen,
}: {
  canEdit: boolean;
  onOpen: (type: string, name: string, version: string) => void;
}) {
  const [compType, setCompType] = useState(STUDIO_COMPONENT_TYPES[0]?.type ?? "branching-model");
  const [compName, setCompName] = useState("");
  const [version, setVersion] = useState("1.0.0");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onCreate(e: FormEvent) {
    e.preventDefault();
    if (!canEdit) return;
    const name = compName.trim();
    if (!name) return;
    setBusy(true);
    setError(null);
    try {
      await api.createComponent({
        type: compType,
        name,
        actor: getActor() || undefined,
      }).catch((err: Error) => {
        if (!err.message.includes("409") && !err.message.toLowerCase().includes("already exists")) {
          throw err;
        }
      });
      await api.createDraftComponentVersion(compType, name, {
        version: version.trim(),
        actor: getActor() || undefined,
      });
      onOpen(compType, name, version.trim());
    } catch (err) {
      setError(err instanceof Error ? err.message : "create failed");
    } finally {
      setBusy(false);
    }
  }

  const selected = studioTypeConfig(compType);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-semibold">Component Studio</h1>
        <p className="mt-1 text-zinc-400">
          UI-first authoring platform components — draft в PostgreSQL, без git и shell-скриптов
        </p>
      </div>

      {error && <p className="text-red-400">{error}</p>}

      <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-6 max-w-xl">
        <h2 className="font-medium">Новый draft</h2>
        {canEdit ? (
          <form onSubmit={onCreate} className="mt-4 space-y-4">
            <label className="block text-xs text-zinc-500">
              Component type
              <select
                value={compType}
                onChange={(e) => setCompType(e.target.value)}
                className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
              >
                {STUDIO_COMPONENT_TYPES.map((t) => (
                  <option key={t.type} value={t.type}>
                    {t.label}
                  </option>
                ))}
              </select>
            </label>
            {selected && <p className="text-xs text-zinc-500">{selected.description}</p>}
            <label className="block text-xs text-zinc-500">
              Name
              <input
                value={compName}
                onChange={(e) => setCompName(e.target.value)}
                placeholder="trunk-based"
                required
                className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
              />
            </label>
            <label className="block text-xs text-zinc-500">
              Version
              <input
                value={version}
                onChange={(e) => setVersion(e.target.value)}
                required
                className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono"
              />
            </label>
            <button
              type="submit"
              disabled={busy}
              className="rounded bg-sky-600 px-4 py-2 text-sm hover:bg-sky-500 disabled:opacity-50"
            >
              {busy ? "Создание…" : "Создать draft и открыть редактор"}
            </button>
          </form>
        ) : (
          <p className="mt-4 text-sm text-zinc-500">Требуется роль publisher</p>
        )}
      </section>

      <section>
        <h2 className="font-medium mb-3">Поддерживаемые типы</h2>
        <ul className="space-y-2 text-sm text-zinc-400">
          {STUDIO_COMPONENT_TYPES.map((t) => (
            <li key={t.type}>
              <span className="font-mono text-zinc-200">{t.type}</span> — {t.description}
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}

function StudioEditor({
  type,
  name,
  version,
  canEdit,
}: {
  type: string;
  name: string;
  version: string;
  canEdit: boolean;
}) {
  const config = studioTypeConfig(type);
  const [detail, setDetail] = useState<ComponentVersionDetail | null>(null);
  const [branchingModel, setBranchingModel] = useState<BranchingModel | null>(null);
  const [gpContent, setGpContent] = useState<GpContentModel | null>(null);
  const [containerfile, setContainerfile] = useState("");
  const [yamlPreview, setYamlPreview] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [artifacts, setArtifacts] = useState<string[]>([]);
  const [dirty, setDirty] = useState(false);
  const [validation, setValidation] = useState<ValidateComponentPackageResult | null>(null);
  const [validating, setValidating] = useState(false);
  const [publishing, setPublishing] = useState(false);

  const isDraft = detail?.status === "draft";
  const readOnly = !canEdit || !isDraft;
  const hasContentRef = !!detail?.contentRef;

  const load = useCallback(async () => {
    setError(null);
    const d = await api.componentVersionDetail(type, name, version);
    setDetail(d);
    if (!isStudioType(type) || !config) return;

    try {
      const list = await api.listComponentArtifacts(type, name, version);
      setArtifacts(list.items.map((a) => a.key));
      if (config.editor === "branching-model") {
        const art = await api.getComponentArtifact(type, name, version, config.primaryArtifact);
        const model = parseBranchingModelYaml(art.body, name);
        setBranchingModel(model);
        setYamlPreview(art.body);
        setGpContent(null);
      } else if (config.editor === "gp-content") {
        const art = await api.getComponentArtifact(type, name, version, GP_CONTENT_ARTIFACT);
        const model = parseGpContentYaml(art.body, name);
        setGpContent(model);
        setYamlPreview(art.body);
        setBranchingModel(null);
        try {
          const cf = await api.getComponentArtifact(type, name, version, GP_CONTAINERFILE_ARTIFACT);
          setContainerfile(cf.body);
        } catch {
          setContainerfile(defaultContainerfile());
        }
      }
    } catch {
      if (config.editor === "branching-model") {
        const model = defaultBranchingModel(name);
        setBranchingModel(model);
        setYamlPreview(serializeBranchingModel(model));
        setGpContent(null);
      } else if (config.editor === "gp-content") {
        const model = defaultGpContent(name);
        setGpContent(model);
        setContainerfile(defaultContainerfile());
        setYamlPreview(serializeGpContent(model));
        setBranchingModel(null);
      }
      setArtifacts([]);
    }
    setValidation(null);
    setDirty(false);
  }, [type, name, version, config]);

  useEffect(() => {
    load().catch((err: Error) => setError(err.message));
  }, [load]);

  useEffect(() => {
    if (branchingModel && config?.editor === "branching-model") {
      setYamlPreview(serializeBranchingModel(branchingModel));
    }
    if (gpContent && config?.editor === "gp-content") {
      setYamlPreview(serializeGpContent(gpContent));
    }
  }, [branchingModel, gpContent, config]);

  async function persistDraft(): Promise<boolean> {
    if (!config || readOnly) return false;
    if (config.editor === "branching-model" && branchingModel) {
      const body = serializeBranchingModel(branchingModel);
      await api.saveComponentArtifact(type, name, version, config.primaryArtifact, body);
    } else if (config.editor === "gp-content" && gpContent) {
      await api.saveComponentArtifact(type, name, version, GP_CONTENT_ARTIFACT, serializeGpContent(gpContent));
      await api.saveComponentArtifact(type, name, version, GP_CONTAINERFILE_ARTIFACT, containerfile);
    } else {
      return false;
    }
    const list = await api.listComponentArtifacts(type, name, version);
    setArtifacts(list.items.map((a) => a.key));
    setDirty(false);
    return true;
  }

  async function saveDraft() {
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      await persistDraft();
      setMessage(`Сохранено: ${config?.primaryArtifact}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "save failed");
    } finally {
      setSaving(false);
    }
  }

  async function runValidate(): Promise<ValidateComponentPackageResult | null> {
    if (!config || readOnly) return null;
    if (config.editor === "branching-model" && !branchingModel) return null;
    if (config.editor === "gp-content" && !gpContent) return null;
    setValidating(true);
    setError(null);
    setMessage(null);
    try {
      const clientIssues =
        config.editor === "branching-model" && branchingModel
          ? validateBranchingModelClient(branchingModel, name)
          : config.editor === "gp-content" && gpContent
            ? validateGpContentClient(gpContent, name)
            : [];
      if (clientIssues.length > 0) {
        const result: ValidateComponentPackageResult = {
          valid: false,
          issues: clientIssues.map((message) => ({ field: "client", message })),
        };
        setValidation(result);
        return result;
      }
      if (dirty) {
        await persistDraft();
      }
      const result = await api.validateComponentPackage(type, name, version);
      setValidation(result);
      if (result.valid) {
        setMessage("Валидация пройдена");
      }
      return result;
    } catch (err) {
      setError(err instanceof Error ? err.message : "validate failed");
      return null;
    } finally {
      setValidating(false);
    }
  }

  async function publishToCanary() {
    if (!config || readOnly) return;
    if (config.editor === "branching-model" && !branchingModel) return;
    if (config.editor === "gp-content" && !gpContent) return;
    setPublishing(true);
    setError(null);
    setMessage(null);
    try {
      const v = validation?.valid ? validation : await runValidate();
      if (!v?.valid) {
        setError("Исправьте ошибки валидации перед публикацией");
        return;
      }
      const manifest =
        config.editor === "branching-model" && branchingModel
          ? buildManifestSubset(type, branchingModel)
          : config.editor === "gp-content" && gpContent
            ? buildGpContentManifestSubset(gpContent)
            : {};
      await api.registerComponentPackage(type, name, version, {
        manifest,
        actor: getActor() || undefined,
      });
      await api.publishComponentToCanary(type, name, version, getActor() || undefined);
      setMessage(
        usesPGOnlyCanaryRegistry(type)
          ? `Опубликовано в canary (PG): ${type}/${name}@${version}`
          : `Опубликовано в canary: ${type}/${name}@${version}`,
      );
      await load();
      setValidation(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "publish failed");
    } finally {
      setPublishing(false);
    }
  }

  if (!config) {
    return (
      <div className="space-y-4">
        <StudioBackLink />
        <p className="text-red-400">Тип {type} пока не поддерживается в Component Studio</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <StudioBackLink />
          <h1 className="mt-2 text-2xl font-semibold font-mono">
            {type}/{name}@{version}
          </h1>
          <p className="mt-1 text-zinc-400">
            Component Studio ·{" "}
            {detail && (
              <span className={statusBadge(detail.status)}>{detail.status}</span>
            )}
          </p>
        </div>
        {canEdit && isDraft && (
          <button
            type="button"
            onClick={saveDraft}
            disabled={saving}
            className="rounded bg-emerald-700 px-4 py-2 text-sm hover:bg-emerald-600 disabled:opacity-50"
          >
            {saving ? "Сохранение…" : "Сохранить draft"}
          </button>
        )}
      </div>

      {error && <p className="text-red-400">{error}</p>}
      {message && <p className="text-emerald-400">{message}</p>}

      {!isDraft && detail && (
        <p className="rounded border border-amber-900/50 bg-amber-950/30 px-4 py-3 text-sm text-amber-200">
          Редактирование доступно только для draft.{" "}
          <Link to={`/components/${type}/${name}/${encodeURIComponent(version)}`} className="text-sky-400 hover:underline">
            Открыть в registry →
          </Link>
        </p>
      )}

      <div className="grid gap-6 lg:grid-cols-2">
        <section className="rounded-lg border border-zinc-800 bg-zinc-900 p-5">
          <h2 className="mb-4 font-medium">{config.label}</h2>
          {config.editor === "branching-model" && branchingModel && (
            <BranchingModelEditor
              model={branchingModel}
              onChange={(m) => {
                setBranchingModel(m);
                setDirty(true);
                setValidation(null);
              }}
              disabled={readOnly}
            />
          )}
          {config.editor === "gp-content" && gpContent && (
            <GpContentEditor
              model={gpContent}
              containerfile={containerfile}
              onChange={(m, cf) => {
                setGpContent(m);
                setContainerfile(cf);
                setDirty(true);
                setValidation(null);
              }}
              disabled={readOnly}
            />
          )}
        </section>

        <section className="space-y-4">
          <div className="rounded-lg border border-zinc-800 bg-zinc-950 p-4">
            <h2 className="mb-2 text-sm font-medium text-zinc-400">{config.primaryArtifact}</h2>
            <pre className="overflow-x-auto text-xs font-mono text-zinc-300 whitespace-pre-wrap">
              {yamlPreview}
            </pre>
          </div>

          {artifacts.length > 0 && (
            <div className="text-xs text-zinc-500">
              Draft artifacts: {artifacts.join(", ")}
            </div>
          )}

          <LifecyclePanel
            type={type}
            status={detail?.status ?? "draft"}
            hasContentRef={hasContentRef}
            validation={validation}
            canEdit={canEdit && isDraft}
            validating={validating}
            publishing={publishing}
            onValidate={() => void runValidate()}
            onPublishCanary={() => void publishToCanary()}
          />
        </section>
      </div>

      {detail?.status === "canary" && (
        <PilotPromotePanel
          type={type}
          name={name}
          version={version}
          canEdit={canEdit}
          onPromoted={() => void load()}
        />
      )}
    </div>
  );
}

function LifecyclePanel({
  type,
  status,
  hasContentRef,
  validation,
  canEdit,
  validating,
  publishing,
  onValidate,
  onPublishCanary,
}: {
  type: string;
  status: string;
  hasContentRef: boolean;
  validation: ValidateComponentPackageResult | null;
  canEdit: boolean;
  validating: boolean;
  publishing: boolean;
  onValidate: () => void;
  onPublishCanary: () => void;
}) {
  const validated = validation?.valid === true;
  const pgOnly = usesPGOnlyCanaryRegistry(type);
  const steps = [
    { id: "draft", label: "Draft (PG)", done: true },
    { id: "validate", label: "Validate", done: validated },
    {
      id: "register",
      label: pgOnly ? "Register (PG)" : "Nexus register",
      done: hasContentRef,
    },
    { id: "canary", label: "Publish canary", done: status === "canary" || status === "published" },
    {
      id: "stable",
      label: pgOnly ? "Promote (Nexus)" : "Promote stable",
      done: status === "published",
    },
  ];

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900 p-4 space-y-4">
      <h2 className="text-sm font-medium">Lifecycle</h2>
      <ol className="space-y-2 text-sm">
        {steps.map((s) => (
          <li key={s.id} className="flex items-center gap-2">
            <span className={s.done ? "text-emerald-400" : "text-zinc-600"}>
              {s.done ? "✓" : "○"}
            </span>
            <span className={s.done ? "text-zinc-200" : "text-zinc-500"}>{s.label}</span>
          </li>
        ))}
      </ol>

      {validation && !validation.valid && validation.issues.length > 0 && (
        <div className="rounded border border-red-900/50 bg-red-950/30 p-3 text-sm">
          <p className="font-medium text-red-300 mb-2">Ошибки валидации</p>
          <ul className="space-y-1 text-red-200/90">
            {validation.issues.map((issue, i) => (
              <li key={`${issue.field}-${i}`}>
                <span className="font-mono text-xs text-red-400">{issue.field}</span>: {issue.message}
              </li>
            ))}
          </ul>
        </div>
      )}

      {canEdit && status === "draft" && (
        <div className="flex flex-wrap gap-2 pt-1">
          <button
            type="button"
            onClick={onValidate}
            disabled={validating || publishing}
            className="rounded border border-zinc-600 px-3 py-1.5 text-sm hover:bg-zinc-800 disabled:opacity-50"
          >
            {validating ? "Проверка…" : "Validate"}
          </button>
          <button
            type="button"
            onClick={onPublishCanary}
            disabled={publishing || validating}
            className="rounded bg-sky-600 px-3 py-1.5 text-sm hover:bg-sky-500 disabled:opacity-50"
          >
            {publishing ? "Публикация…" : "Publish to canary"}
          </button>
        </div>
      )}

      {status === "canary" && (
        <p className="text-xs text-sky-400">
          Версия в canary — назначьте pilot projects и promote stable ниже.
          {pgOnly && " Promote загрузит immutable package в Nexus."}
        </p>
      )}
    </div>
  );
}

function StudioBackLink() {
  return (
    <Link to="/studio" className="text-sm text-sky-400 hover:underline">
      ← Component Studio
    </Link>
  );
}
