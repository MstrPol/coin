import { Link } from "react-router-dom";
import { GP_DRAFT_SLOT_ORDER, SLOT_LABELS } from "../lib/gpSlots";
import { gpSlotEmptyVersionHint } from "../lib/gpCompositionVersions";
import { platformEditPath } from "../lib/platformComponentPaths";

const inputClass =
  "mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono";

function statusBadge(status: string | undefined) {
  if (status === "draft") {
    return <span className="ml-2 text-xs text-amber-400">draft</span>;
  }
  if (status === "published") {
    return <span className="ml-2 text-xs text-emerald-400">published</span>;
  }
  return null;
}

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
  versionStatuses?: Record<string, Record<string, string>>;
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
  versionStatuses = {},
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

  const draftPins = GP_DRAFT_SLOT_ORDER.flatMap((key) => {
    const ver = composition[key];
    if (!ver) return [];
    const status = versionStatuses[key]?.[ver];
    if (status !== "draft") return [];
    const compType = key === "gp-content" ? "gp-content" : key === "branching-model" ? "branching-model" : "agent";
    return [{ type: compType, name: componentName(key), version: ver }];
  });

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
          const selectedVersion = composition[key] ?? "";
          const selectedStatus = versionStatuses[key]?.[selectedVersion];

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
                  <div className={`${inputClass} text-zinc-300`}>
                    {selectedVersion || "—"}
                    {statusBadge(selectedStatus)}
                  </div>
                ) : (
                  <>
                    <select
                      value={selectedVersion}
                      onChange={(e) =>
                        onCompositionChange({ ...composition, [key]: e.target.value })
                      }
                      className={inputClass}
                      required
                      disabled={versions.length === 0}
                    >
                      {versions.length === 0 ? (
                        <option value="">{gpSlotEmptyVersionHint(key)}</option>
                      ) : (
                        versions.map((v) => {
                          const st = versionStatuses[key]?.[v];
                          const suffix = st === "draft" ? " (draft)" : "";
                          return (
                            <option key={v} value={v}>
                              {v}
                              {suffix}
                            </option>
                          );
                        })
                      )}
                    </select>
                    {selectedStatus === "draft" && (
                      <p className="mt-1 text-xs text-amber-400/90">
                        Pin — draft, может измениться до publish.
                      </p>
                    )}
                  </>
                )}
              </label>
            </div>
          );
        })}
      </div>

      {draftPins.length > 0 && !readOnly && (
        <div className="mt-4 rounded border border-amber-900/50 bg-amber-950/20 px-4 py-3 text-sm text-amber-200/90">
          <p className="font-medium text-amber-300">Draft pins блокируют promote GP</p>
          <ul className="mt-2 space-y-1">
            {draftPins.map((pin) => {
              const href = platformEditPath(pin.type, pin.name, pin.version);
              return (
                <li key={`${pin.type}/${pin.name}@${pin.version}`} className="font-mono text-xs">
                  {pin.type}/{pin.name}@{pin.version}
                  {href && (
                    <>
                      {" "}
                      <Link to={href} className="text-sky-400 hover:underline">
                        Publish →
                      </Link>
                    </>
                  )}
                </li>
              );
            })}
          </ul>
        </div>
      )}
    </div>
  );
}
