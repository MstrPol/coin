import { useEffect, useRef, useState } from "react";
import type {
  BuildStackParameter,
  GpContentModel,
  GpContentPreviewResult,
  InlineBuild,
  InlineRun,
  PipelineAction,
  PipelineStage,
  PipelineStep,
} from "../../lib/gpContentYaml";
import { collectStageIds, generateUniqueShortId } from "../../lib/gpContentYaml";

type Props = {
  model: GpContentModel;
  onChange: (model: GpContentModel) => void;
  disabled?: boolean;
};

const inputClass = "mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 font-mono text-xs";
const cardClass = "space-y-3 rounded border border-zinc-800 p-4";

export default function GpContentEditor({ model, onChange, disabled }: Props) {
  return (
    <div className="space-y-6">
      <EditorHeader />
      <ParametersCard model={model} disabled={disabled} onChange={onChange} />
      <PipelineCard model={model} disabled={disabled} onChange={onChange} />
    </div>
  );
}

function EditorHeader() {
  return (
    <p className="text-xs text-zinc-500">
      Build Stack v3: Parameters → Pipeline stages. Containerfile задаётся inline в buildkit steps.
    </p>
  );
}

function ParametersCard({
  model,
  disabled,
  onChange,
}: {
  model: GpContentModel;
  disabled?: boolean;
  onChange: (model: GpContentModel) => void;
}) {
  const updateParam = (index: number, patch: Partial<BuildStackParameter>) => {
    const parameters = model.parameters.map((param, i) => (i === index ? { ...param, ...patch } : param));
    onChange({ ...model, parameters });
  };
  return (
    <section className={cardClass}>
      <CardTitle
        title="Parameters"
        disabled={disabled}
        onAdd={() => onChange({ ...model, parameters: [...model.parameters, { name: "PARAM", type: "string" }] })}
      />
      {model.parameters.map((param, index) => (
        <div
          key={index}
          className={`grid gap-2 items-end ${param.type === "enum" ? "md:grid-cols-6" : "md:grid-cols-5"}`}
        >
          <TextInput label="name" value={param.name} disabled={disabled} onChange={(name) => updateParam(index, { name })} />
          <label className="block text-xs text-zinc-500">
            type
            <select
              value={param.type}
              disabled={disabled}
              onChange={(e) => {
                const type = e.target.value as BuildStackParameter["type"];
                const patch: Partial<BuildStackParameter> = { type };
                if (type !== "enum") patch.allowedValues = undefined;
                updateParam(index, patch);
              }}
              className={inputClass}
            >
              <option value="string">string</option>
              <option value="boolean">boolean</option>
              <option value="number">number</option>
              <option value="enum">enum</option>
            </select>
          </label>
          <TextInput
            label="default"
            value={param.default == null ? "" : String(param.default)}
            disabled={disabled}
            onChange={(value) => updateParam(index, { default: value || undefined })}
          />
          {param.type === "enum" && (
            <AllowedValuesInput
              values={param.allowedValues}
              disabled={disabled}
              onChange={(allowedValues) => updateParam(index, { allowedValues })}
            />
          )}
          <label className="flex items-center gap-2 self-end pb-2 text-xs text-zinc-500">
            <input
              type="checkbox"
              checked={param.required ?? false}
              disabled={disabled}
              onChange={(e) => updateParam(index, { required: e.target.checked || undefined })}
              className="rounded border-zinc-600 bg-zinc-950"
            />
            required
          </label>
          <RemoveButton disabled={disabled} onClick={() => onChange({ ...model, parameters: model.parameters.filter((_, i) => i !== index) })} />
        </div>
      ))}
    </section>
  );
}

function PipelineCard({
  model,
  disabled,
  onChange,
}: {
  model: GpContentModel;
  disabled?: boolean;
  onChange: (model: GpContentModel) => void;
}) {
  const stageKeysRef = useRef<string[]>([]);
  const stepKeysRef = useRef<string[][]>([]);

  const stages = model.pipeline?.stages ?? [];
  const buildIds = stages.flatMap((s) =>
    s.steps.filter((st) => st.action === "build" && st.build?.id).map((st) => st.build!.id),
  );

  const syncStageKeys = () => {
    const keys = stageKeysRef.current;
    while (keys.length < stages.length) {
      keys.push(newListKey("stage"));
    }
    if (keys.length > stages.length) {
      keys.length = stages.length;
    }
    const stepKeys = stepKeysRef.current;
    while (stepKeys.length < stages.length) {
      stepKeys.push([]);
    }
    if (stepKeys.length > stages.length) {
      stepKeys.length = stages.length;
    }
    stages.forEach((stage, stageIndex) => {
      const keysForStage = stepKeys[stageIndex] ?? [];
      while (keysForStage.length < stage.steps.length) {
        keysForStage.push(newListKey("step"));
      }
      if (keysForStage.length > stage.steps.length) {
        keysForStage.length = stage.steps.length;
      }
      stepKeys[stageIndex] = keysForStage;
    });
  };
  syncStageKeys();

  const updateStages = (stages: PipelineStage[]) => onChange({ ...model, pipeline: { stages } });

  const updateStage = (index: number, patch: Partial<PipelineStage>) => {
    updateStages(stages.map((stage, i) => (i === index ? { ...stage, ...patch } : stage)));
  };

  const moveStage = (from: number, to: number) => {
    const nextStages = [...stages];
    if (to < 0 || to >= nextStages.length || from === to) return;
    const keys = [...stageKeysRef.current];
    const stepKeys = [...stepKeysRef.current];
    const [item] = nextStages.splice(from, 1);
    const [key] = keys.splice(from, 1);
    const [stageStepKeys] = stepKeys.splice(from, 1);
    nextStages.splice(to, 0, item);
    keys.splice(to, 0, key);
    stepKeys.splice(to, 0, stageStepKeys);
    stageKeysRef.current = keys;
    stepKeysRef.current = stepKeys;
    updateStages(nextStages);
  };

  const removeStage = (index: number) => {
    if (stages.length <= 1) return;
    stageKeysRef.current = stageKeysRef.current.filter((_, i) => i !== index);
    stepKeysRef.current = stepKeysRef.current.filter((_, i) => i !== index);
    updateStages(stages.filter((_, i) => i !== index));
  };

  const addStage = () => {
    const id = generateUniqueShortId(collectStageIds(model));
    stageKeysRef.current.push(newListKey("stage"));
    stepKeysRef.current.push([]);
    updateStages([...stages, { id, name: "Stage", steps: [] }]);
  };

  const addStep = (stageIndex: number, action: PipelineAction) => {
    const stepKeys = stepKeysRef.current[stageIndex] ?? [];
    stepKeys.push(newListKey("step"));
    stepKeysRef.current[stageIndex] = stepKeys;
    const stage = stages[stageIndex];
    updateStage(stageIndex, { steps: [...stage.steps, newStep(action)] });
  };

  const removeStep = (stageIndex: number, stepIndex: number) => {
    const stepKeys = stepKeysRef.current[stageIndex] ?? [];
    stepKeysRef.current[stageIndex] = stepKeys.filter((_, i) => i !== stepIndex);
    const stage = stages[stageIndex];
    updateStage(stageIndex, { steps: stage.steps.filter((_, i) => i !== stepIndex) });
  };

  const newStep = (action: PipelineAction): PipelineStep => {
    if (action === "run") return { action, run: { engine: "buildkit", output: "validate", containerfile: { body: "" } } };
    if (action === "build") {
      const id = generateUniqueShortId(buildIds);
      return { action, build: { id, type: "image", engine: "buildkit", containerfile: { body: "" } } };
    }
    return { action, publish: { buildStepId: buildIds[0] ?? "" } };
  };

  return (
    <section className={cardClass}>
      <CardTitle title="Pipeline stages" disabled={disabled} onAdd={addStage} />
      {stages.map((stage, stageIndex) => (
        <div key={stageKeysRef.current[stageIndex]} className="space-y-3 rounded border border-zinc-800 p-3">
          <div className="grid gap-2 md:grid-cols-[1fr_1fr_auto]">
            <TextInput label="id" value={stage.id} disabled={disabled} onChange={(id) => updateStage(stageIndex, { id })} />
            <TextInput label="name" value={stage.name} disabled={disabled} onChange={(name) => updateStage(stageIndex, { name })} />
            <div className="flex flex-wrap items-end gap-1 self-end pb-0.5">
              <MoveButton
                label="↑"
                title="Переместить выше"
                disabled={disabled || stageIndex === 0}
                onClick={() => moveStage(stageIndex, stageIndex - 1)}
              />
              <MoveButton
                label="↓"
                title="Переместить ниже"
                disabled={disabled || stageIndex === stages.length - 1}
                onClick={() => moveStage(stageIndex, stageIndex + 1)}
              />
              <RemoveButton
                disabled={disabled || stages.length <= 1}
                onClick={() => removeStage(stageIndex)}
              />
            </div>
          </div>
          <div className="flex flex-wrap gap-1">
            {(["run", "build", "publish"] as PipelineAction[]).map((action) => (
              <button
                key={action}
                type="button"
                disabled={disabled}
                className="rounded border border-zinc-700 px-2 py-1 text-xs disabled:opacity-40"
                onClick={() => addStep(stageIndex, action)}
              >
                + {action}
              </button>
            ))}
          </div>
          {stage.steps.map((step, stepIndex) => (
            <StepCard
              key={stepKeysRef.current[stageIndex]?.[stepIndex] ?? stepIndex}
              step={step}
              buildIds={buildIds}
              disabled={disabled}
              onChange={(next) => {
                const steps = stage.steps.map((s, i) => (i === stepIndex ? next : s));
                updateStage(stageIndex, { steps });
              }}
              onRemove={() => removeStep(stageIndex, stepIndex)}
            />
          ))}
        </div>
      ))}
    </section>
  );
}

function StepCard({
  step,
  buildIds,
  disabled,
  onChange,
  onRemove,
}: {
  step: PipelineStep;
  buildIds: string[];
  disabled?: boolean;
  onChange: (step: PipelineStep) => void;
  onRemove: () => void;
}) {
  return (
    <div className="space-y-2 rounded border border-zinc-700/60 p-3">
      <div className="flex items-center justify-between gap-2">
        <span className="text-xs font-medium text-zinc-400">action: {step.action}</span>
        <RemoveButton disabled={disabled} onClick={onRemove} />
      </div>
      {step.action === "run" && step.run && (
        <RunBuildFields
          label="run"
          engine={step.run.engine}
          output={step.run.output ?? ""}
          containerfileBody={step.run.containerfile?.body ?? ""}
          dockerfilePath={step.run.dockerfile?.path ?? ""}
          disabled={disabled}
          onChange={(patch) => onChange({ ...step, run: { ...step.run!, ...patch } })}
        />
      )}
      {step.action === "build" && step.build && (
        <>
          <div className="grid gap-2 md:grid-cols-3">
            <TextInput label="build.id" value={step.build.id} disabled={disabled} onChange={(id) => onChange({ ...step, build: { ...step.build!, id } })} />
            <label className="block text-xs text-zinc-500">
              type
              <select
                value={step.build.type}
                disabled={disabled}
                onChange={(e) => onChange({ ...step, build: { ...step.build!, type: e.target.value as InlineBuild["type"] } })}
                className={inputClass}
              >
                <option value="image">image</option>
                <option value="liquibase-image">liquibase-image</option>
                <option value="artifact">artifact</option>
              </select>
            </label>
          </div>
          <RunBuildFields
            label="build"
            engine={step.build.engine}
            containerfileBody={step.build.containerfile?.body ?? ""}
            dockerfilePath={step.build.dockerfile?.path ?? ""}
            disabled={disabled}
            onChange={(patch) => onChange({ ...step, build: { ...step.build!, ...patch } })}
          />
        </>
      )}
      {step.action === "publish" && (
        <label className="block text-xs text-zinc-500">
          buildStepId
          <select
            value={step.publish?.buildStepId ?? ""}
            disabled={disabled}
            onChange={(e) => onChange({ ...step, publish: { buildStepId: e.target.value } })}
            className={inputClass}
          >
            <option value="">—</option>
            {buildIds.map((id) => (
              <option key={id} value={id}>
                {id}
              </option>
            ))}
          </select>
        </label>
      )}
    </div>
  );
}

function RunBuildFields({
  label,
  engine,
  output,
  containerfileBody,
  dockerfilePath,
  disabled,
  onChange,
}: {
  label: string;
  engine: "buildkit" | "dockerfile";
  output?: string;
  containerfileBody: string;
  dockerfilePath: string;
  disabled?: boolean;
  onChange: (patch: Partial<InlineRun & InlineBuild>) => void;
}) {
  return (
    <div className="space-y-2">
      <div className="grid gap-2 md:grid-cols-2">
        <label className="block text-xs text-zinc-500">
          engine
          <select
            value={engine}
            disabled={disabled}
            onChange={(e) => onChange({ engine: e.target.value as "buildkit" | "dockerfile" })}
            className={inputClass}
          >
            <option value="buildkit">buildkit</option>
            <option value="dockerfile">dockerfile</option>
          </select>
        </label>
        {label === "run" && (
          <label className="block text-xs text-zinc-500">
            output
            <select
              value={output ?? ""}
              disabled={disabled}
              onChange={(e) => onChange({ output: e.target.value || undefined })}
              className={inputClass}
            >
              <option value="validate">validate</option>
              <option value="test">test</option>
              <option value="image">image</option>
              <option value="artifact">artifact</option>
            </select>
          </label>
        )}
      </div>
      {engine === "buildkit" ? (
        <label className="block text-xs text-zinc-500">
          containerfile.body
          <textarea
            value={containerfileBody}
            disabled={disabled}
            rows={8}
            onChange={(e) => onChange({ containerfile: { body: e.target.value } })}
            className="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 font-mono text-xs"
          />
        </label>
      ) : (
        <TextInput label="dockerfile.path" value={dockerfilePath} disabled={disabled} onChange={(path) => onChange({ dockerfile: { path } })} />
      )}
    </div>
  );
}

export function ResolvedManifestPreview({
  preview,
  previewing,
  previewError,
}: {
  preview: GpContentPreviewResult | null;
  previewing: boolean;
  previewError: string | null;
}) {
  const manifestPreview =
    preview && preview.pipeline
      ? {
          parameters: preview.parameters,
          pipeline: preview.pipeline,
          artifacts: preview.artifacts,
        }
      : null;
  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900 p-4 space-y-3">
      <div className="flex items-center justify-between gap-2">
        <h2 className="text-sm font-medium">Resolved manifest preview</h2>
        {previewing && <span className="text-xs text-zinc-500">обновление…</span>}
      </div>
      {previewError && <p className="text-xs text-red-400">{previewError}</p>}
      {preview && !preview.valid && (
        <ul className="space-y-1 text-xs text-amber-400">
          {preview.issues.map((iss) => (
            <li key={`${iss.field}-${iss.message}`}>
              {iss.field}: {iss.message}
            </li>
          ))}
        </ul>
      )}
      {preview?.warnings?.map((warning) => (
        <p key={warning} className="text-xs text-zinc-500">
          {warning}
        </p>
      ))}
      {manifestPreview && (
        <pre className="max-h-[32rem] overflow-auto rounded bg-zinc-950 p-3 text-xs text-zinc-400">
          {JSON.stringify(manifestPreview, null, 2)}
        </pre>
      )}
    </div>
  );
}

function AllowedValuesInput({
  values,
  disabled,
  onChange,
}: {
  values?: string[];
  disabled?: boolean;
  onChange: (values: string[] | undefined) => void;
}) {
  const focusedRef = useRef(false);
  const [draft, setDraft] = useState(() => values?.join(", ") ?? "");
  const valuesKey = values?.join("\0") ?? "";

  useEffect(() => {
    if (!focusedRef.current) setDraft(values?.join(", ") ?? "");
  }, [valuesKey, values]);

  function commit(nextDraft: string) {
    const parsed = nextDraft
      .split(",")
      .map((value) => value.trim())
      .filter(Boolean);
    onChange(parsed.length > 0 ? parsed : undefined);
    setDraft(parsed.join(", "));
  }

  return (
    <label className="block text-xs text-zinc-500">
      allowedValues
      <input
        value={draft}
        disabled={disabled}
        placeholder="value1, value2"
        onFocus={() => {
          focusedRef.current = true;
        }}
        onBlur={() => {
          focusedRef.current = false;
          commit(draft);
        }}
        onChange={(e) => setDraft(e.target.value)}
        className={inputClass}
      />
    </label>
  );
}

function CardTitle({ title, disabled, onAdd }: { title: string; disabled?: boolean; onAdd: () => void }) {
  return (
    <div className="flex items-center justify-between gap-2">
      <h3 className="text-sm font-medium text-zinc-300">{title}</h3>
      {!disabled && (
        <button type="button" className="rounded border border-zinc-700 px-2 py-1 text-xs hover:bg-zinc-800" onClick={onAdd}>
          + add
        </button>
      )}
    </div>
  );
}

function TextInput({
  label,
  value,
  disabled,
  onChange,
}: {
  label: string;
  value: string;
  disabled?: boolean;
  onChange: (value: string) => void;
}) {
  return (
    <label className="block text-xs text-zinc-500">
      {label}
      <input value={value} disabled={disabled} onChange={(e) => onChange(e.target.value)} className={inputClass} />
    </label>
  );
}

function newListKey(prefix: string): string {
  return `${prefix}-${Math.random().toString(36).slice(2, 9)}`;
}

function RemoveButton({ disabled, onClick }: { disabled?: boolean; onClick: () => void }) {
  return (
    <button type="button" disabled={disabled} className="self-end text-xs text-red-400 disabled:opacity-40" onClick={onClick}>
      удалить
    </button>
  );
}

function MoveButton({
  label,
  title,
  disabled,
  onClick,
}: {
  label: string;
  title: string;
  disabled?: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      title={title}
      disabled={disabled}
      className="rounded border border-zinc-700 px-2 py-1 text-xs text-zinc-400 hover:bg-zinc-800 disabled:opacity-40"
      onClick={onClick}
    >
      {label}
    </button>
  );
}
