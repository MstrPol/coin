import type { BranchingModel, BranchRule } from "../../lib/branchingModelYaml";

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

    </div>
  );
}
