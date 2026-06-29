export type BranchRule = {
  name: string;
  pattern: string;
  versioning: { template: string };
  publish: boolean;
};

export type BranchingModel = {
  schemaVersion: number;
  name: string;
  branches: BranchRule[];
};

export type PreviewScenario = {
  id: string;
  branch: string;
  tagName?: string;
  tags?: string[];
  requestPublish?: boolean;
};

export type BranchingPreviewResult = {
  patternHint: string;
  results: Array<{
    id: string;
    matchedRule?: string;
    branchValid: boolean;
    branchError?: string;
    coinVersion?: string;
    versionError?: string;
    publishOutcome: string;
    publishReason?: string;
  }>;
};

const TRUNK_BASED_RULES: BranchRule[] = [
  {
    name: "main",
    pattern: "^main$|^master$",
    versioning: { template: "v{base}-main-snapshot-{n}" },
    publish: false,
  },
  {
    name: "feature",
    pattern: "^feature/(?P<jira>[A-Z][A-Z0-9]*-\\d+)(?:-.+)?$",
    versioning: { template: "v{base}-{jira}-snapshot-{n}" },
    publish: false,
  },
  {
    name: "bugfix",
    pattern: "^bugfix/(?P<jira>[A-Z][A-Z0-9]*-\\d+)(?:-.+)?$",
    versioning: { template: "v{base}-{jira}-snapshot-{n}" },
    publish: false,
  },
  {
    name: "release",
    pattern: "^release/(?P<jira>[A-Z][A-Z0-9]*-\\d+)(?:-.+)?$",
    versioning: { template: "v{base}-{jira}-rc-{n}" },
    publish: true,
  },
];

export const DEFAULT_PREVIEW_SCENARIOS: PreviewScenario[] = [
  { id: "main", branch: "main", requestPublish: true },
  { id: "feature", branch: "feature/PROJ-101", requestPublish: true },
  { id: "release", branch: "release/PROJ-404", requestPublish: true },
];

export function defaultBranchingModel(name: string): BranchingModel {
  return {
    schemaVersion: 2,
    name,
    branches: TRUNK_BASED_RULES.map((r) => ({ ...r, versioning: { ...r.versioning } })),
  };
}

function yamlQuote(value: string): string {
  if (/[:#{}[\],&*!|>'"%@`]/.test(value) || value.includes(" ")) {
    return `"${value.replace(/\\/g, "\\\\").replace(/"/g, '\\"')}"`;
  }
  return value;
}

export function serializeBranchingModel(model: BranchingModel): string {
  const lines = [`schemaVersion: ${model.schemaVersion}`, `name: ${model.name}`, "branches:"];
  for (const br of model.branches) {
    lines.push(`  - name: ${br.name}`);
    lines.push(`    pattern: ${yamlQuote(br.pattern)}`);
    lines.push("    versioning:");
    lines.push(`      template: ${yamlQuote(br.versioning.template)}`);
    lines.push(`    publish: ${br.publish}`);
  }
  return `${lines.join("\n")}\n`;
}

function parseBranchesBlock(lines: string[], start: number): { branches: BranchRule[]; next: number } {
  const branches: BranchRule[] = [];
  let i = start;
  while (i < lines.length) {
    const line = lines[i];
    if (line.trim() === "" || line.trim().startsWith("#")) {
      i++;
      continue;
    }
    if (!line.startsWith("  -")) {
      break;
    }
    const rule: BranchRule = {
      name: "",
      pattern: "",
      versioning: { template: "" },
      publish: false,
    };
    i++;
    while (i < lines.length && (lines[i].startsWith("    ") || lines[i].trim() === "")) {
      const inner = lines[i].trim();
      if (inner.startsWith("name:")) rule.name = inner.slice(5).trim();
      if (inner.startsWith("pattern:")) rule.pattern = inner.slice(8).trim().replace(/^"|"$/g, "");
      if (inner.startsWith("template:")) rule.versioning.template = inner.slice(9).trim().replace(/^"|"$/g, "");
      if (inner.startsWith("publish:")) rule.publish = inner.slice(8).trim() === "true";
      i++;
      if (i < lines.length && lines[i].startsWith("  -")) break;
    }
    if (rule.name && rule.pattern) {
      branches.push(rule);
    }
  }
  return { branches, next: i };
}

export function parseBranchingModelYaml(raw: string, fallbackName: string): BranchingModel {
  const model = defaultBranchingModel(fallbackName);
  const lines = raw.split(/\r?\n/);
  for (let i = 0; i < lines.length; i++) {
    const trimmed = lines[i].trim();
    if (!trimmed || trimmed.startsWith("#")) continue;
    if (trimmed.startsWith("schemaVersion:")) {
      model.schemaVersion = Number(trimmed.split(":")[1].trim());
    } else if (trimmed.startsWith("name:")) {
      model.name = trimmed.slice(5).trim();
    } else if (trimmed === "branches:") {
      const parsed = parseBranchesBlock(lines, i + 1);
      if (parsed.branches.length > 0) {
        model.branches = parsed.branches;
      }
      i = parsed.next - 1;
    }
  }
  return model;
}

export function buildManifestSubset(
  type: string,
  model: BranchingModel,
): Record<string, unknown> {
  if (type === "branching-model") {
    return {
      branching: {
        name: model.name,
        branches: model.branches,
      },
    };
  }
  return {};
}

export function validateBranchingModelClient(model: BranchingModel, componentName: string): string[] {
  const issues: string[] = [];
  if (model.schemaVersion !== 2) issues.push("schemaVersion должен быть 2");
  if (model.name !== componentName) {
    issues.push(`model.name (${model.name}) должен совпадать с именем компонента (${componentName})`);
  }
  if (model.branches.length === 0) issues.push("нужно хотя бы одно правило branches[]");
  model.branches.forEach((br, i) => {
    if (!br.name.trim()) issues.push(`branches[${i}].name обязателен`);
    if (!br.pattern.trim()) issues.push(`branches[${i}].pattern обязателен`);
    if (!br.versioning.template.trim()) issues.push(`branches[${i}].versioning.template обязателен`);
  });
  return issues;
}

export function previewModelPayload(model: BranchingModel) {
  return {
    name: model.name,
    branches: model.branches,
  };
}
