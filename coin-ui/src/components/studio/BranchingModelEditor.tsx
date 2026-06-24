import type { BranchingModel } from "../../lib/branchingModelYaml";

type Props = {
  model: BranchingModel;
  onChange: (model: BranchingModel) => void;
  disabled?: boolean;
};

const PUBLISH_OPTIONS = [
  { value: "tag", label: "tag — только RC-теги" },
  { value: "branch", label: "branch — snapshot с ветки" },
  { value: "always", label: "always" },
  { value: "never", label: "never" },
] as const;

export default function BranchingModelEditor({ model, onChange, disabled }: Props) {
  function patch(partial: Partial<BranchingModel>) {
    onChange({ ...model, ...partial });
  }

  return (
    <div className="space-y-5">
      <section className="space-y-3">
        <h3 className="text-sm font-medium text-zinc-300">Trunk</h3>
        <label className="block text-xs text-zinc-500">
          Основная ветка
          <input
            value={model.trunk.branch}
            disabled={disabled}
            onChange={(e) => patch({ trunk: { branch: e.target.value } })}
            className="mt-1 w-full max-w-xs rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono disabled:opacity-50"
          />
        </label>
      </section>

      <section className="space-y-3">
        <h3 className="text-sm font-medium text-zinc-300">Типы веток</h3>
        <p className="text-xs text-zinc-500">Через запятую: feature, bugfix, release</p>
        <input
          value={model.branchTypes.join(", ")}
          disabled={disabled}
          onChange={(e) =>
            patch({
              branchTypes: e.target.value
                .split(",")
                .map((s) => s.trim())
                .filter(Boolean),
            })
          }
          className="w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono disabled:opacity-50"
        />
      </section>

      <section className="space-y-3">
        <h3 className="text-sm font-medium text-zinc-300">Versioning</h3>
        <label className="block text-xs text-zinc-500">
          Префикс git-тега
          <input
            value={model.versioning.tagPrefix}
            disabled={disabled}
            onChange={(e) =>
              patch({
                versioning: { ...model.versioning, tagPrefix: e.target.value },
              })
            }
            className="mt-1 w-24 rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono disabled:opacity-50"
          />
        </label>
        <div className="flex flex-wrap gap-4 text-sm">
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={model.versioning.qualifiers.snapshot.enabled}
              disabled={disabled}
              onChange={(e) =>
                patch({
                  versioning: {
                    ...model.versioning,
                    qualifiers: {
                      ...model.versioning.qualifiers,
                      snapshot: { enabled: e.target.checked },
                    },
                  },
                })
              }
            />
            snapshot
          </label>
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={model.versioning.qualifiers.rc.enabled}
              disabled={disabled}
              onChange={(e) =>
                patch({
                  versioning: {
                    ...model.versioning,
                    qualifiers: {
                      ...model.versioning.qualifiers,
                      rc: { ...model.versioning.qualifiers.rc, enabled: e.target.checked },
                    },
                  },
                })
              }
            />
            rc
          </label>
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={model.versioning.qualifiers.rc.releaseBranchesOnly}
              disabled={disabled}
              onChange={(e) =>
                patch({
                  versioning: {
                    ...model.versioning,
                    qualifiers: {
                      ...model.versioning.qualifiers,
                      rc: {
                        ...model.versioning.qualifiers.rc,
                        releaseBranchesOnly: e.target.checked,
                      },
                    },
                  },
                })
              }
            />
            rc только с release/*
          </label>
        </div>
      </section>

      <section className="space-y-3">
        <h3 className="text-sm font-medium text-zinc-300">Publish policy</h3>
        <select
          value={model.publish.when}
          disabled={disabled}
          onChange={(e) =>
            patch({
              publish: {
                ...model.publish,
                when: e.target.value as BranchingModel["publish"]["when"],
              },
            })
          }
          className="rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm disabled:opacity-50"
        >
          {PUBLISH_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
        {model.publish.when === "branch" && (
          <label className="block text-xs text-zinc-500">
            Ветка для snapshot publish
            <input
              value={model.publish.branch ?? "main"}
              disabled={disabled}
              onChange={(e) =>
                patch({ publish: { ...model.publish, branch: e.target.value } })
              }
              className="mt-1 w-full max-w-xs rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm font-mono disabled:opacity-50"
            />
          </label>
        )}
      </section>
    </div>
  );
}
