import { useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import BranchingModelEditor from "../studio/BranchingModelEditor";
import GpContentEditor from "../studio/GpContentEditor";
import type { ComponentVersionDetail, ValidateComponentPackageResult } from "../../api/types";
import { api, getActor } from "../../lib/api";
import {
  buildManifestSubset,
  defaultBranchingModel,
  parseBranchingModelYaml,
  serializeBranchingModel,
  validateBranchingModelClient,
  type BranchingModel,
} from "../../lib/branchingModelYaml";
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
} from "../../lib/gpContentYaml";
import { platformCatalogPath } from "../../lib/platformComponentPaths";
import { platformTypeConfig } from "../../lib/platformComponentTypes";

function statusClass(status: string): string {
  switch (status) {
    case "draft":
      return "text-amber-400";
    case "published":
      return "text-emerald-400";
    default:
      return "text-zinc-400";
  }
}

export default function PlatformComponentEditor({
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
  const config = platformTypeConfig(type);
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
    if (!config) return;

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

  async function publishToStable() {
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
      await api.promoteComponentVersion(type, name, version, getActor() || undefined);
      setMessage(`Опубликовано: ${type}/${name}@${version}`);
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
        <PlatformBackLink type={type} />
        <p className="text-red-400">Тип {type} не поддерживается</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <PlatformBackLink type={type} />
          <h1 className="mt-2 text-2xl font-semibold font-mono">
            {type}/{name}@{version}
          </h1>
          <p className="mt-1 text-zinc-400">
            Platform ·{" "}
            {detail && <span className={statusClass(detail.status)}>{detail.status}</span>}
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
          <Link
            to={`/components/${type}/${name}/${encodeURIComponent(version)}`}
            className="text-sky-400 hover:underline"
          >
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
              componentName={name}
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
            <div className="text-xs text-zinc-500">Draft artifacts: {artifacts.join(", ")}</div>
          )}

          <LifecyclePanel
            status={detail?.status ?? "draft"}
            hasContentRef={hasContentRef}
            validation={validation}
            canEdit={canEdit && isDraft}
            validating={validating}
            publishing={publishing}
            onValidate={() => void runValidate()}
            onPublish={() => void publishToStable()}
          />
        </section>
      </div>
    </div>
  );
}

function LifecyclePanel({
  status,
  hasContentRef,
  validation,
  canEdit,
  validating,
  publishing,
  onValidate,
  onPublish,
}: {
  status: string;
  hasContentRef: boolean;
  validation: ValidateComponentPackageResult | null;
  canEdit: boolean;
  validating: boolean;
  publishing: boolean;
  onValidate: () => void;
  onPublish: () => void;
}) {
  const validated = validation?.valid === true;
  const steps = [
    { id: "draft", label: "Draft", done: true },
    { id: "validate", label: "Validate", done: validated },
    { id: "register", label: "Register package", done: hasContentRef },
    { id: "published", label: "Publish", done: status === "published" },
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
            onClick={onPublish}
            disabled={publishing || validating}
            className="rounded bg-sky-600 px-3 py-1.5 text-sm hover:bg-sky-500 disabled:opacity-50"
          >
            {publishing ? "Публикация…" : "Publish"}
          </button>
        </div>
      )}
    </div>
  );
}

function PlatformBackLink({ type }: { type: string }) {
  const to = platformCatalogPath(type) ?? "/platform/build-stacks";
  return (
    <Link to={to} className="text-sm text-sky-400 hover:underline">
      ← Platform catalog
    </Link>
  );
}
