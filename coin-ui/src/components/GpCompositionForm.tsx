import { GP_DRAFT_SLOT_ORDER, SLOT_LABELS } from "../lib/gpSlots";

const inputClass =
  "mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono";

export type GpCompositionFormProps = {
  agentStackName: string;
  agentStackOptions: string[];
  onAgentStackChange: (v: string) => void;
  gpContentName: string;
  gpContentOptions: string[];
  onGpContentChange: (v: string) => void;
  branchingModelName: string;
  branchingModelOptions: string[];
  onBranchingModelChange: (v: string) => void;
  composition: Record<string, string>;
  versionOptions: Record<string, string[]>;
  onCompositionChange: (v: Record<string, string>) => void;
  readOnly?: boolean;
};

export default function GpCompositionForm({
  agentStackName,
  agentStackOptions,
  onAgentStackChange,
  gpContentName,
  gpContentOptions,
  onGpContentChange,
  branchingModelName,
  branchingModelOptions,
  onBranchingModelChange,
  composition,
  versionOptions,
  onCompositionChange,
  readOnly = false,
}: GpCompositionFormProps) {
  function componentName(key: string): string {
    if (key === "agent") return agentStackName;
    if (key === "gp-content") return gpContentName;
    return branchingModelName;
  }

  function componentPrefix(key: string): string {
    if (key === "agent") return "agent";
    if (key === "gp-content") return "gp-content";
    return "branching-model";
  }

  function onNameChange(key: string, value: string) {
    if (readOnly) return;
    if (key === "agent") onAgentStackChange(value);
    else if (key === "gp-content") onGpContentChange(value);
    else onBranchingModelChange(value);
  }

  return (
    <div className="mt-4 overflow-x-auto">
      <div className="hidden min-w-[32rem] grid-cols-[6.5rem_1fr_7.5rem] gap-x-4 gap-y-1 px-1 text-xs font-medium text-zinc-500 sm:grid">
        <span>Slot</span>
        <span>Component</span>
        <span>Version</span>
      </div>
      <div className="space-y-4 sm:space-y-3">
        {GP_DRAFT_SLOT_ORDER.map((key) => {
          const nameOptions =
            key === "agent"
              ? agentStackOptions
              : key === "gp-content"
                ? gpContentOptions
                : branchingModelOptions;
          const versions = versionOptions[key] ?? [];
          const prefix = componentPrefix(key);

          return (
            <div
              key={key}
              className="min-w-[32rem] grid gap-3 sm:grid-cols-[6.5rem_1fr_7.5rem] sm:items-end"
            >
              <div className="font-mono text-sm text-sky-400 sm:pb-2">{key}</div>

              <label className="block text-sm">
                <span className="text-xs text-zinc-500">{SLOT_LABELS[key]}</span>
                {readOnly ? (
                  <div className={`${inputClass} text-zinc-300`}>
                    {prefix}/{componentName(key)}
                  </div>
                ) : (
                  <select
                    value={componentName(key)}
                    onChange={(e) => onNameChange(key, e.target.value)}
                    className={inputClass}
                    required
                  >
                    {nameOptions.map((n) => (
                      <option key={n} value={n}>
                        {prefix}/{n}
                      </option>
                    ))}
                  </select>
                )}
              </label>

              <label className="block text-sm">
                <span className="text-xs text-zinc-500">Version</span>
                {readOnly ? (
                  <div className={`${inputClass} text-zinc-300`}>{composition[key] ?? "—"}</div>
                ) : (
                  <select
                    value={composition[key] ?? ""}
                    onChange={(e) =>
                      onCompositionChange({ ...composition, [key]: e.target.value })
                    }
                    className={inputClass}
                    required
                    disabled={versions.length === 0}
                  >
                    {versions.length === 0 ? (
                      <option value="">— нет published —</option>
                    ) : (
                      versions.map((v) => (
                        <option key={v} value={v}>
                          {v}
                        </option>
                      ))
                    )}
                  </select>
                )}
              </label>
            </div>
          );
        })}
      </div>
    </div>
  );
}
