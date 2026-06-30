import { useCallback, useEffect, useState } from "react";
import type { Deliverable, GpContentModel, GpContentPreviewResult, PipelineStage } from "../../lib/gpContentYaml";
import { defaultGpContent, defaultGpContentDocker } from "../../lib/gpContentYaml";
import { api } from "../../lib/api";

type Props = {
  model: GpContentModel;
  containerfile: string;
  componentName: string;
  onChange: (model: GpContentModel, containerfile: string) => void;
  disabled?: boolean;
};

function toggleDeliverable(list: Deliverable[], item: Deliverable, on: boolean): Deliverable[] {
  if (on) return list.includes(item) ? list : [...list, item];
  return list.filter((d) => d !== item);
}

export default function GpContentEditor({ model, containerfile, componentName, onChange, disabled }: Props) {
  const [preview, setPreview] = useState<GpContentPreviewResult | null>(null);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [previewing, setPreviewing] = useState(false);

  function updateModel(patch: Partial<GpContentModel>) {
    onChange({ ...model, ...patch }, containerfile);
  }

  function setEngine(engine: GpContentModel["build"]["engine"]) {
    if (engine === model.build.engine) return;
    const next =
      engine === "dockerfile"
        ? defaultGpContentDocker(componentName || model.name)
        : defaultGpContent(componentName || model.name, "buildkit");
    onChange(next, engine === "buildkit" ? containerfile : "");
  }

  function updateStage(index: number, patch: Partial<PipelineStage>) {
    const stages = model.pipeline.stages.map((s, i) => (i === index ? { ...s, ...patch } : s));
    onChange({ ...model, pipeline: { stages } }, containerfile);
  }

  const runPreview = useCallback(async () => {
    setPreviewing(true);
    setPreviewError(null);
    try {
      const result = await api.gpContentPreview(model, componentName);
      setPreview(result);
    } catch (err) {
      setPreviewError(err instanceof Error ? err.message : "preview failed");
      setPreview(null);
    } finally {
      setPreviewing(false);
    }
  }, [model, componentName]);

  useEffect(() => {
    if (disabled) return;
    const timer = window.setTimeout(() => {
      void runPreview();
    }, 400);
    return () => window.clearTimeout(timer);
  }, [model, disabled, runPreview]);

  return (
    <div className="space-y-6">
      <section className="grid gap-4 md:grid-cols-2">
        <label className="block text-xs text-zinc-500">
          Build engine
          <select
            value={model.build.engine}
            disabled={disabled}
            onChange={(e) => setEngine(e.target.value as GpContentModel["build"]["engine"])}
            className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 font-mono"
          >
            <option value="buildkit">buildkit (managed Containerfile)</option>
            <option value="dockerfile">dockerfile (BYO)</option>
          </select>
        </label>
        <div className="text-xs text-zinc-500">
          <p className="mb-2">Deliverables</p>
          <div className="flex flex-wrap gap-3">
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={model.capabilities.deliverables.includes("image")}
                disabled={disabled}
                onChange={(e) =>
                  updateModel({
                    capabilities: {
                      deliverables: toggleDeliverable(model.capabilities.deliverables, "image", e.target.checked),
                    },
                  })
                }
              />
              image
            </label>
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={model.capabilities.deliverables.includes("artifact")}
                disabled={disabled || model.build.engine === "dockerfile"}
                onChange={(e) =>
                  updateModel({
                    capabilities: {
                      deliverables: toggleDeliverable(model.capabilities.deliverables, "artifact", e.target.checked),
                    },
                  })
                }
              />
              artifact (buildkit only)
            </label>
          </div>
        </div>
      </section>

      {model.build.engine === "buildkit" && model.build.buildkit && (
        <section className="space-y-3 rounded border border-zinc-800 p-4">
          <h3 className="text-sm font-medium text-zinc-300">BuildKit</h3>
          <label className="block text-xs text-zinc-500">
            Cache ref template
            <input
              value={model.build.buildkit.cacheRefTemplate}
              disabled={disabled}
              onChange={(e) =>
                updateModel({
                  build: {
                    ...model.build,
                    buildkit: { ...model.build.buildkit!, cacheRefTemplate: e.target.value },
                  },
                })
              }
              className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 font-mono text-xs"
            />
          </label>
          <label className="block text-xs text-zinc-500">
            Containerfile artifact
            <textarea
              value={containerfile}
              disabled={disabled}
              rows={10}
              onChange={(e) => onChange(model, e.target.value)}
              className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 font-mono text-xs"
            />
          </label>
        </section>
      )}

      {model.build.engine === "dockerfile" && model.build.dockerfile && (
        <section className="space-y-3 rounded border border-zinc-800 p-4">
          <h3 className="text-sm font-medium text-zinc-300">BYO Dockerfile</h3>
          <label className="block text-xs text-zinc-500">
            Path in workspace
            <input
              value={model.build.dockerfile.path}
              disabled={disabled}
              onChange={(e) =>
                updateModel({
                  build: {
                    ...model.build,
                    dockerfile: { ...model.build.dockerfile!, path: e.target.value },
                  },
                })
              }
              className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 font-mono text-xs"
            />
          </label>
          <div className="grid gap-3 md:grid-cols-2">
            <label className="block text-xs text-zinc-500">
              imageTarget
              <input
                value={model.build.dockerfile.imageTarget ?? ""}
                disabled={disabled}
                onChange={(e) =>
                  updateModel({
                    build: {
                      ...model.build,
                      dockerfile: { ...model.build.dockerfile!, imageTarget: e.target.value || undefined },
                    },
                  })
                }
                className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 font-mono text-xs"
              />
            </label>
            <label className="block text-xs text-zinc-500">
              testTarget
              <input
                value={model.build.dockerfile.testTarget ?? ""}
                disabled={disabled}
                onChange={(e) =>
                  updateModel({
                    build: {
                      ...model.build,
                      dockerfile: { ...model.build.dockerfile!, testTarget: e.target.value || undefined },
                    },
                  })
                }
                className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 font-mono text-xs"
              />
            </label>
          </div>
          <label className="block text-xs text-zinc-500">
            Cache ref template
            <input
              value={model.build.dockerfile.cacheRefTemplate}
              disabled={disabled}
              onChange={(e) =>
                updateModel({
                  build: {
                    ...model.build,
                    dockerfile: { ...model.build.dockerfile!, cacheRefTemplate: e.target.value },
                  },
                })
              }
              className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 font-mono text-xs"
            />
          </label>
        </section>
      )}

      <section className="space-y-3">
        <div className="flex items-center justify-between gap-2">
          <h3 className="text-sm font-medium text-zinc-300">Pipeline stages</h3>
          {!disabled && (
            <button
              type="button"
              className="rounded border border-zinc-700 px-2 py-1 text-xs hover:bg-zinc-800"
              onClick={() =>
                onChange(
                  { ...model, pipeline: { stages: [...model.pipeline.stages, { id: "stage", name: "Stage" }] } },
                  containerfile,
                )
              }
            >
              + stage
            </button>
          )}
        </div>
        {model.pipeline.stages.map((st, i) => (
          <div key={`${st.id}-${i}`} className="grid grid-cols-3 gap-2 items-end rounded border border-zinc-800 p-3">
            <label className="block text-xs text-zinc-500">
              id
              <input
                value={st.id}
                disabled={disabled}
                onChange={(e) => updateStage(i, { id: e.target.value })}
                className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 font-mono text-xs"
              />
            </label>
            <label className="block text-xs text-zinc-500">
              name
              <input
                value={st.name}
                disabled={disabled}
                onChange={(e) => updateStage(i, { name: e.target.value })}
                className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 font-mono text-xs"
              />
            </label>
            {!disabled && model.pipeline.stages.length > 1 && (
              <button
                type="button"
                onClick={() =>
                  onChange(
                    { ...model, pipeline: { stages: model.pipeline.stages.filter((_, j) => j !== i) } },
                    containerfile,
                  )
                }
                className="text-xs text-red-400 hover:text-red-300 pb-1.5"
              >
                удалить
              </button>
            )}
          </div>
        ))}
      </section>

      <section className="space-y-2 rounded border border-zinc-800 p-4">
        <div className="flex items-center justify-between gap-2">
          <h3 className="text-sm font-medium text-zinc-300">Manifest preview</h3>
          {previewing && <span className="text-xs text-zinc-500">обновление…</span>}
        </div>
        {previewError && <p className="text-xs text-red-400">{previewError}</p>}
        {preview && !preview.valid && (
          <ul className="text-xs text-amber-400 space-y-1">
            {preview.issues.map((iss) => (
              <li key={`${iss.field}-${iss.message}`}>
                {iss.field}: {iss.message}
              </li>
            ))}
          </ul>
        )}
        {preview?.warnings?.map((w) => (
          <p key={w} className="text-xs text-zinc-500">
            {w}
          </p>
        ))}
        {preview?.build && (
          <pre className="overflow-x-auto rounded bg-zinc-950 p-3 text-xs text-zinc-400">
            {JSON.stringify({ build: preview.build, pipeline: preview.pipeline, capabilities: preview.capabilities }, null, 2)}
          </pre>
        )}
      </section>
    </div>
  );
}
