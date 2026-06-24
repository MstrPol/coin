export type BranchingModel = {
  schemaVersion: number;
  name: string;
  trunk: { branch: string };
  branchTypes: string[];
  versioning: {
    tagPrefix: string;
    qualifiers: {
      snapshot: { enabled: boolean };
      rc: { enabled: boolean; releaseBranchesOnly: boolean };
    };
  };
  publish: {
    when: "tag" | "branch" | "always" | "never";
    branch?: string;
  };
};

export function defaultBranchingModel(name: string): BranchingModel {
  return {
    schemaVersion: 1,
    name,
    trunk: { branch: "main" },
    branchTypes: ["feature", "bugfix", "release"],
    versioning: {
      tagPrefix: "v",
      qualifiers: {
        snapshot: { enabled: true },
        rc: { enabled: true, releaseBranchesOnly: true },
      },
    },
    publish: { when: "tag" },
  };
}

export function serializeBranchingModel(model: BranchingModel): string {
  const lines = [
    `schemaVersion: ${model.schemaVersion}`,
    `name: ${model.name}`,
    "trunk:",
    `  branch: ${model.trunk.branch}`,
    "branchTypes:",
    ...model.branchTypes.map((t) => `  - ${t}`),
    "versioning:",
    `  tagPrefix: ${model.versioning.tagPrefix}`,
    "  qualifiers:",
    "    snapshot:",
    `      enabled: ${model.versioning.qualifiers.snapshot.enabled}`,
    "    rc:",
    `      enabled: ${model.versioning.qualifiers.rc.enabled}`,
    `      releaseBranchesOnly: ${model.versioning.qualifiers.rc.releaseBranchesOnly}`,
    "publish:",
    `  when: ${model.publish.when}`,
  ];
  if (model.publish.when === "branch" && model.publish.branch) {
    lines.push(`  branch: ${model.publish.branch}`);
  }
  return `${lines.join("\n")}\n`;
}

/** Minimal YAML parser for Component Studio scaffold (known model.yaml shape). */
export function parseBranchingModelYaml(raw: string, fallbackName: string): BranchingModel {
  const model = defaultBranchingModel(fallbackName);
  const lines = raw.split(/\r?\n/);
  let inBranchTypes = false;
  const branchTypes: string[] = [];

  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;

    if (trimmed === "branchTypes:") {
      inBranchTypes = true;
      continue;
    }
    if (inBranchTypes) {
      if (trimmed.startsWith("- ")) {
        branchTypes.push(trimmed.slice(2).trim());
        continue;
      }
      if (!line.startsWith(" ")) {
        inBranchTypes = false;
      }
    }

    const kv = trimmed.match(/^([a-zA-Z]+):\s*(.+)$/);
    if (!kv) continue;
    const [, key, value] = kv;

    switch (key) {
      case "schemaVersion":
        model.schemaVersion = Number(value);
        break;
      case "name":
        model.name = value;
        break;
      case "branch":
        if (line.includes("trunk:") || lines[lines.indexOf(line) - 1]?.includes("trunk:")) {
          model.trunk.branch = value;
        } else if (model.publish.when === "branch") {
          model.publish.branch = value;
        }
        break;
      case "tagPrefix":
        model.versioning.tagPrefix = value;
        break;
      case "when":
        if (model.publish.when !== value) {
          model.publish.when = value as BranchingModel["publish"]["when"];
        }
        break;
      case "enabled":
        if (lines.slice(Math.max(0, lines.indexOf(line) - 3), lines.indexOf(line)).join("\n").includes("snapshot:")) {
          model.versioning.qualifiers.snapshot.enabled = value === "true";
        } else if (lines.slice(Math.max(0, lines.indexOf(line) - 3), lines.indexOf(line)).join("\n").includes("rc:")) {
          model.versioning.qualifiers.rc.enabled = value === "true";
        }
        break;
      case "releaseBranchesOnly":
        model.versioning.qualifiers.rc.releaseBranchesOnly = value === "true";
        break;
      default:
        break;
    }
  }

  if (branchTypes.length > 0) {
    model.branchTypes = branchTypes;
  }
  return model;
}

/** content_ref v2 manifest subset for resolve materializer. */
export function buildManifestSubset(
  type: string,
  model: BranchingModel,
): Record<string, unknown> {
  if (type === "branching-model") {
    return {
      branching: {
        name: model.name,
        trunk: model.trunk,
        branchTypes: model.branchTypes,
        versioning: model.versioning,
        publish: model.publish,
      },
    };
  }
  return {};
}

export function validateBranchingModelClient(model: BranchingModel, componentName: string): string[] {
  const issues: string[] = [];
  if (!model.trunk.branch.trim()) issues.push("trunk.branch обязателен");
  if (model.branchTypes.length === 0) issues.push("нужен хотя бы один branch type");
  if (!model.branchTypes.includes("release")) issues.push("branchTypes должен включать release");
  if (model.name !== componentName) {
    issues.push(`model.name (${model.name}) должен совпадать с именем компонента (${componentName})`);
  }
  if (model.publish.when === "branch" && !model.publish.branch?.trim()) {
    issues.push("publish.branch обязателен при publish.when=branch");
  }
  return issues;
}
