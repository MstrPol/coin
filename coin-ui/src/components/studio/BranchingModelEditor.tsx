import { useCallback, useEffect, useState } from "react";
import type { BranchingModel, BranchingPreviewResult, BranchRule, PreviewScenario } from "../../lib/branchingModelYaml";
import { DEFAULT_PREVIEW_SCENARIOS } from "../../lib/branchingModelYaml";
import { api } from "../../lib/api";

type Props = {
  model: BranchingModel;
  onChange: (model: BranchingModel) => void;
  disabled?: boolean;
};

function emptyRule(): BranchRule {
  return {
    name: "rule",
    pattern: "^feature/.+$",
    versioning: { template: "v{base}-snapshot-{n}" },
    publish: false,
  };
}

export default function BranchingModelEditor({ model, onChange, disabled }: Props) {
  const [testBranch, setTestBranch] = useState("feature/PROJ-101");
  const [requestPublish, setRequestPublish] = useState(true);
  const [preview, setPreview] = useState<BranchingPreviewResult | null>(null);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [previewing, setPreviewing] = useState(false);

  function patchRule(index: number, partial: Partial<BranchRule>) {
    const branches = model.branches.map((br, i) => (i === index ? { ...br, ...partial } : br));
    onChange({ ...model, branches });
  }

  function moveRule(index: number, delta: number) {
    const next = index + delta;
    if (next < 0 || next >= model.branches.length) return;
    const branches = [...model.branches];
    const [item] = branches.splice(index, 1);
    branches.splice(next, 0, item);
    onChange({ ...model, branches });
  }

  const runPreview = useCallback(
    async (scenarios: PreviewScenario[]) => {
      setPreviewing(true);
      setPreviewError(null);
      try {
        const result = await api.branchingModelPreview(model, scenarios);
        setPreview(result);
      } catch (err) {
        setPreviewError(err instanceof Error ? err.message : "preview failed");
        setPreview(null);
      } finally {
        setPreviewing(false);
      }
    },
    [model],
  );

  useEffect(() => {
    if (disabled) return;
    const timer = window.setTimeout(() => {
      void runPreview(DEFAULT_PREVIEW_SCENARIOS);
    }, 400);
    return () => window.clearTimeout(timer);
  }, [model, disabled, runPreview]);

  return (
    <div className="space-y-6">
      <section className="space-y-3">
        <div className="flex items-center justify-between gap-2">
          <h3 className="text-sm font-medium text-zinc-300">Branch rules (first match wins)</h3>
          {!disabled && (
            <button
              type="button"
              className="rounded border border-zinc-700 px-2 py-1 text-xs hover:bg-zinc-800"
              onClick={() => onChange({ ...model, branches: [...model.branches, emptyRule()] })}
            >
              + правило
            </button>
          )}
        </div>
        <div className="space-y-4">
          {model.branches.map((br, index) => (
            <div key={`${br.name}-${index}`} className="rounded border border-zinc-800 bg-zinc-950 p-4 space-y-3">
              <div className="flex flex-wrap items-center gap-2">
                <span className="text-xs text-zinc-500">#{index + 1}</span>
                {!disabled && (
                  <>
                    <button type="button" className="text-xs text-zinc-500 hover:text-zinc-300" onClick={() => moveRule(index, -1)}>
                      ↑
                    </button>
                    <button type="button" className="text-xs text-zinc-500 hover:text-zinc-300" onClick={() => moveRule(index, 1)}>
                      ↓
                    </button>
                    <button
                      type="button"
                      className="text-xs text-red-400 hover:text-red-300"
                      onClick={() => onChange({ ...model, branches: model.branches.filter((_, i) => i !== index) })}
                    >
                      удалить
                    </button>
                  </>
                )}
              </div>
              <label className="block text-xs text-zinc-500">
                name
                <input
                  value={br.name}
                  disabled={disabled}
                  onChange={(e) => patchRule(index, { name: e.target.value })}
                  className="mt-1 w-full rounded border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm font-mono disabled:opacity-50"
                />
              </label>
              <label className="block text-xs text-zinc-500">
                pattern (Go RE2, named groups (?P&lt;jira&gt;...))
                <input
                  value={br.pattern}
                  disabled={disabled}
                  onChange={(e) => patchRule(index, { pattern: e.target.value })}
                  className="mt-1 w-full rounded border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm font-mono disabled:opacity-50"
                />
              </label>
              <label className="block text-xs text-zinc-500">
                versioning.template
                <input
                  value={br.versioning.template}
                  disabled={disabled}
                  onChange={(e) =>
                    patchRule(index, { versioning: { template: e.target.value } })
                  }
                  className="mt-1 w-full rounded border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm font-mono disabled:opacity-50"
                />
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={br.publish}
                  disabled={disabled}
                  onChange={(e) => patchRule(index, { publish: e.target.checked })}
                />
                publish allowed
              </label>
            </div>
          ))}
        </div>
      </section>

      <section className="space-y-3 rounded border border-zinc-800 bg-zinc-950 p-4">
        <h3 className="text-sm font-medium text-zinc-300">Preview</h3>
        <p className="text-xs text-zinc-500">
          {preview?.patternHint ?? "Go RE2 named groups; first matching rule wins"}
        </p>
        <div className="flex flex-wrap items-end gap-3">
          <label className="block text-xs text-zinc-500">
            test branch
            <input
              value={testBranch}
              disabled={disabled}
              onChange={(e) => setTestBranch(e.target.value)}
              className="mt-1 w-56 rounded border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm font-mono disabled:opacity-50"
            />
          </label>
          <label className="flex items-center gap-2 text-sm pb-2">
            <input
              type="checkbox"
              checked={requestPublish}
              disabled={disabled}
              onChange={(e) => setRequestPublish(e.target.checked)}
            />
            requestPublish
          </label>
          <button
            type="button"
            disabled={disabled || previewing}
            onClick={() =>
              void runPreview([{ id: "custom", branch: testBranch, requestPublish }])
            }
            className="rounded bg-zinc-800 px-3 py-2 text-sm hover:bg-zinc-700 disabled:opacity-50"
          >
            {previewing ? "…" : "Проверить ветку"}
          </button>
        </div>
        {previewError && <p className="text-sm text-red-400">{previewError}</p>}
        {preview && (
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead>
                <tr className="text-left text-zinc-500">
                  <th className="py-1 pr-2">scenario</th>
                  <th className="py-1 pr-2">rule</th>
                  <th className="py-1 pr-2">version</th>
                  <th className="py-1">publish</th>
                </tr>
              </thead>
              <tbody>
                {preview.results.map((row) => (
                  <tr key={row.id} className="border-t border-zinc-800">
                    <td className="py-2 pr-2 font-mono">{row.id}</td>
                    <td className="py-2 pr-2">{row.matchedRule ?? row.branchError ?? "—"}</td>
                    <td className="py-2 pr-2 font-mono">{row.coinVersion ?? row.versionError ?? "—"}</td>
                    <td className="py-2">{row.publishOutcome}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  );
}
