import type { GpContentModel, PipelineStage } from "../../lib/gpContentYaml";

type Props = {
  model: GpContentModel;
  containerfile: string;
  onChange: (model: GpContentModel, containerfile: string) => void;
  disabled?: boolean;
};

export default function GpContentEditor({ model, containerfile, onChange, disabled }: Props) {
  function updateModel(patch: Partial<GpContentModel>) {
    onChange({ ...model, ...patch }, containerfile);
  }

  function updateStage(index: number, patch: Partial<PipelineStage>) {
    const stages = model.pipeline.stages.map((s, i) => (i === index ? { ...s, ...patch } : s));
    onChange({ ...model, pipeline: { stages } }, containerfile);
  }

  function addStage() {
    const stages = [...model.pipeline.stages, { id: "stage", name: "Stage" }];
    onChange({ ...model, pipeline: { stages } }, containerfile);
  }

  function removeStage(index: number) {
    const stages = model.pipeline.stages.filter((_, i) => i !== index);
    onChange({ ...model, pipeline: { stages } }, containerfile);
  }

  return (
    <div className="space-y-5 text-sm">
      <label className="block text-xs text-zinc-500">
        Build engine
        <select
          value={model.build.engine}
          disabled={disabled}
          onChange={(e) =>
            updateModel({
              build: {
                ...model.build,
                engine: e.target.value as GpContentModel["build"]["engine"],
              },
            })
          }
          className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 font-mono"
        >
          <option value="buildkit">buildkit</option>
          <option value="buildpack">buildpack</option>
          <option value="dockerfile">dockerfile</option>
        </select>
      </label>

      {model.build.engine === "buildkit" && model.build.buildkit && (
        <div className="space-y-3 rounded border border-zinc-800 p-3">
          <p className="text-xs text-zinc-500">BuildKit</p>
          <label className="block text-xs text-zinc-500">
            Dockerfile path
            <input
              value={model.build.buildkit.dockerfile}
              disabled={disabled}
              onChange={(e) =>
                updateModel({
                  build: {
                    ...model.build,
                    buildkit: { ...model.build.buildkit!, dockerfile: e.target.value },
                  },
                })
              }
              className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 font-mono text-xs"
            />
          </label>
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
        </div>
      )}

      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <p className="text-xs text-zinc-500">Pipeline stages</p>
          {!disabled && (
            <button
              type="button"
              onClick={addStage}
              className="text-xs text-sky-400 hover:underline"
            >
              + stage
            </button>
          )}
        </div>
        {model.pipeline.stages.map((st, i) => (
          <div key={`${st.id}-${i}`} className="grid grid-cols-3 gap-2 items-end">
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
            <div className="flex gap-2">
              <label className="block text-xs text-zinc-500 flex-1">
                when
                <input
                  value={st.when ?? ""}
                  disabled={disabled}
                  placeholder="tag"
                  onChange={(e) => updateStage(i, { when: e.target.value || undefined })}
                  className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 font-mono text-xs"
                />
              </label>
              {!disabled && model.pipeline.stages.length > 1 && (
                <button
                  type="button"
                  onClick={() => removeStage(i)}
                  className="text-xs text-red-400 hover:underline pb-1.5"
                >
                  ×
                </button>
              )}
            </div>
          </div>
        ))}
      </div>

      <label className="block text-xs text-zinc-500">
        Containerfile
        <textarea
          value={containerfile}
          disabled={disabled}
          rows={12}
          onChange={(e) => onChange(model, e.target.value)}
          className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 font-mono text-xs"
        />
      </label>
    </div>
  );
}
