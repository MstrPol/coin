export type PipelineStage = {
  id: string;
  name: string;
};

export type Deliverable = "image" | "artifact";

export type GpContentModel = {
  schemaVersion: number;
  name: string;
  kind: "gp-content";
  capabilities: {
    deliverables: Deliverable[];
  };
  build: {
    engine: "buildkit" | "dockerfile";
    buildkit?: {
      targets: Record<string, string>;
      cacheRefTemplate: string;
    };
    dockerfile?: {
      path: string;
      imageTarget?: string;
      testTarget?: string;
      cacheRefTemplate: string;
    };
  };
  pipeline: {
    stages: PipelineStage[];
  };
  artifacts: {
    validateSchema: string;
    containerfile?: string;
  };
};

export type GpContentPreviewResult = {
  valid: boolean;
  issues: Array<{ field: string; message: string }>;
  warnings?: string[];
  build?: Record<string, unknown>;
  pipeline?: Record<string, unknown>;
  capabilities?: Record<string, unknown>;
};

export const GP_CONTENT_ARTIFACT = "content.yaml";
export const GP_CONTAINERFILE_ARTIFACT = "dockerfiles/Containerfile";
export const GP_SCHEMA_ARTIFACT = "schemas/config.v2.schema.json";

const BUILDKIT_TARGETS: Record<string, string> = {
  validate: "validate",
  test: "test",
  image: "runtime",
  artifact: "artifact",
};

export function defaultContainerfile(): string {
  return `# Coin managed Containerfile: Go build targets for BuildKit engine.
# Не копируйте в репозитории сервисов.

FROM --platform=$TARGETPLATFORM golang:1.22-bookworm AS base
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .

FROM base AS validate
RUN test -f go.mod && test -f main.go

FROM base AS test
RUN go test ./...

FROM base AS artifact
RUN mkdir -p /out && CGO_ENABLED=0 go build -buildvcs=false -trimpath -o /out/app .

FROM --platform=$TARGETPLATFORM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=artifact /out/app /app/app
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
`;
}

export function defaultGpContent(name: string, preset: "buildkit" | "dockerfile" = "buildkit"): GpContentModel {
  if (preset === "dockerfile") {
    return defaultGpContentDocker(name);
  }
  return {
    schemaVersion: 2,
    name,
    kind: "gp-content",
    capabilities: { deliverables: ["image", "artifact"] },
    build: {
      engine: "buildkit",
      buildkit: {
        targets: { ...BUILDKIT_TARGETS },
        cacheRefTemplate: "{{registryHost}}/coin-cache/{{project}}:buildkit",
      },
    },
    pipeline: {
      stages: [
        { id: "validate", name: "Validate" },
        { id: "test", name: "Test" },
        { id: "build", name: "Build" },
        { id: "publish", name: "Publish" },
      ],
    },
    artifacts: {
      validateSchema: GP_SCHEMA_ARTIFACT,
      containerfile: GP_CONTAINERFILE_ARTIFACT,
    },
  };
}

export function defaultGpContentDocker(name: string): GpContentModel {
  return {
    schemaVersion: 2,
    name,
    kind: "gp-content",
    capabilities: { deliverables: ["image"] },
    build: {
      engine: "dockerfile",
      dockerfile: {
        path: "Dockerfile",
        imageTarget: "runtime",
        testTarget: "test",
        cacheRefTemplate: "{{registryHost}}/coin-cache/{{project}}:dockerfile",
      },
    },
    pipeline: {
      stages: [
        { id: "validate", name: "Validate" },
        { id: "test", name: "Test" },
        { id: "build", name: "Build" },
        { id: "publish", name: "Publish" },
      ],
    },
    artifacts: {
      validateSchema: GP_SCHEMA_ARTIFACT,
    },
  };
}

function yamlQuote(value: string): string {
  if (/[:#{}[\],&*!|>'"%@`]/.test(value) || value.includes(" ")) {
    return `"${value.replace(/\\/g, "\\\\").replace(/"/g, '\\"')}"`;
  }
  return value;
}

export function serializeGpContent(model: GpContentModel): string {
  const lines = [
    `schemaVersion: ${model.schemaVersion}`,
    `name: ${model.name}`,
    `kind: ${model.kind}`,
    "",
    "capabilities:",
    "  deliverables:",
  ];
  for (const d of model.capabilities.deliverables) {
    lines.push(`    - ${d}`);
  }
  lines.push("build:", `  engine: ${model.build.engine}`);
  if (model.build.engine === "buildkit" && model.build.buildkit) {
    const bk = model.build.buildkit;
    lines.push("  buildkit:", "    targets:");
    for (const [k, v] of Object.entries(bk.targets)) {
      lines.push(`      ${k}: ${v}`);
    }
    lines.push(`    cacheRefTemplate: ${yamlQuote(bk.cacheRefTemplate)}`);
  }
  if (model.build.engine === "dockerfile" && model.build.dockerfile) {
    const df = model.build.dockerfile;
    lines.push("  dockerfile:");
    lines.push(`    path: ${df.path}`);
    if (df.imageTarget) lines.push(`    imageTarget: ${df.imageTarget}`);
    if (df.testTarget) lines.push(`    testTarget: ${df.testTarget}`);
    lines.push(`    cacheRefTemplate: ${yamlQuote(df.cacheRefTemplate)}`);
  }
  lines.push("pipeline:", "  stages:");
  for (const st of model.pipeline.stages) {
    lines.push(`    - id: ${st.id}`, `      name: ${st.name}`);
  }
  lines.push("artifacts:");
  lines.push(`  validateSchema: ${model.artifacts.validateSchema}`);
  if (model.artifacts.containerfile) {
    lines.push(`  containerfile: ${model.artifacts.containerfile}`);
  }
  return `${lines.join("\n")}\n`;
}

function parseStagesBlock(lines: string[], start: number): { stages: PipelineStage[]; next: number } {
  const stages: PipelineStage[] = [];
  let i = start;
  while (i < lines.length) {
    const line = lines[i];
    if (line.trim() === "" || line.trim().startsWith("#")) {
      i++;
      continue;
    }
    if (!line.startsWith("    -")) break;
    const stage: PipelineStage = { id: "", name: "" };
    const idMatch = line.match(/id:\s*(.+)$/);
    if (idMatch) stage.id = idMatch[1].trim();
    i++;
    while (i < lines.length && lines[i].startsWith("      ")) {
      const inner = lines[i].trim();
      if (inner.startsWith("name:")) stage.name = inner.slice(5).trim();
      i++;
    }
    if (stage.id && stage.name) stages.push(stage);
  }
  return { stages, next: i };
}

function parseTargetsBlock(lines: string[], start: number): { targets: Record<string, string>; next: number } {
  const targets: Record<string, string> = {};
  let i = start;
  while (i < lines.length) {
    const line = lines[i];
    if (!line.startsWith("      ")) break;
    const kv = line.trim().match(/^([a-zA-Z]+):\s*(.+)$/);
    if (kv) targets[kv[1]] = kv[2];
    i++;
  }
  return { targets, next: i };
}

export function parseGpContentYaml(raw: string, fallbackName: string): GpContentModel {
  const model = defaultGpContent(fallbackName);
  const lines = raw.split(/\r?\n/);
  let i = 0;
  const deliverables: Deliverable[] = [];
  let inDeliverables = false;

  while (i < lines.length) {
    const line = lines[i];
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) {
      i++;
      continue;
    }

    if (trimmed === "deliverables:") {
      inDeliverables = true;
      i++;
      continue;
    }
    if (inDeliverables) {
      if (trimmed.startsWith("- ")) {
        const d = trimmed.slice(2).trim() as Deliverable;
        if (d === "image" || d === "artifact") deliverables.push(d);
        i++;
        continue;
      }
      if (!line.startsWith("    ")) inDeliverables = false;
    }

    if (trimmed === "stages:") {
      const parsed = parseStagesBlock(lines, i + 1);
      if (parsed.stages.length > 0) model.pipeline.stages = parsed.stages;
      i = parsed.next;
      continue;
    }

    if (trimmed === "targets:") {
      const parsed = parseTargetsBlock(lines, i + 1);
      if (model.build.buildkit && Object.keys(parsed.targets).length > 0) {
        model.build.buildkit.targets = parsed.targets;
      }
      i = parsed.next;
      continue;
    }

    const kv = trimmed.match(/^([a-zA-Z]+):\s*(.+)$/);
    if (kv) {
      const [, key, value] = kv;
      switch (key) {
        case "schemaVersion":
          model.schemaVersion = Number(value);
          break;
        case "name":
          model.name = value;
          break;
        case "kind":
          model.kind = value as GpContentModel["kind"];
          break;
        case "engine":
          model.build.engine = value as GpContentModel["build"]["engine"];
          break;
        case "path":
          if (!model.build.dockerfile) {
            model.build.dockerfile = {
              path: value,
              cacheRefTemplate: "{{registryHost}}/coin-cache/{{project}}:dockerfile",
            };
          } else {
            model.build.dockerfile.path = value;
          }
          break;
        case "imageTarget":
          if (model.build.dockerfile) model.build.dockerfile.imageTarget = value;
          break;
        case "testTarget":
          if (model.build.dockerfile) model.build.dockerfile.testTarget = value;
          break;
        case "cacheRefTemplate":
          if (model.build.engine === "dockerfile" && model.build.dockerfile) {
            model.build.dockerfile.cacheRefTemplate = value.replace(/^"|"$/g, "");
          } else if (model.build.buildkit) {
            model.build.buildkit.cacheRefTemplate = value.replace(/^"|"$/g, "");
          }
          break;
        case "validateSchema":
          model.artifacts.validateSchema = value;
          break;
        case "containerfile":
          model.artifacts.containerfile = value;
          break;
        default:
          break;
      }
    }
    i++;
  }

  if (deliverables.length > 0) {
    model.capabilities.deliverables = deliverables;
  }
  if (model.build.engine === "dockerfile") {
    delete model.build.buildkit;
    if (!model.build.dockerfile) {
      model.build.dockerfile = {
        path: "Dockerfile",
        cacheRefTemplate: "{{registryHost}}/coin-cache/{{project}}:dockerfile",
      };
    }
    delete model.artifacts.containerfile;
  }
  if (model.build.engine === "buildkit" && !model.build.buildkit) {
    model.build.buildkit = {
      targets: { ...BUILDKIT_TARGETS },
      cacheRefTemplate: "{{registryHost}}/coin-cache/{{project}}:buildkit",
    };
    delete model.build.dockerfile;
  }
  return model;
}

export function buildGpContentManifestSubset(model: GpContentModel): Record<string, unknown> {
  const subset: Record<string, unknown> = {
    capabilities: { deliverables: [...model.capabilities.deliverables] },
    build: { engine: model.build.engine },
    pipeline: { stages: model.pipeline.stages.map((s) => ({ id: s.id, name: s.name })) },
    validateSchema: { artifactKey: model.artifacts.validateSchema },
  };
  if (model.build.engine === "buildkit" && model.build.buildkit) {
    subset.build = {
      engine: model.build.engine,
      buildkit: {
        targets: model.build.buildkit.targets,
        cacheRefTemplate: model.build.buildkit.cacheRefTemplate,
      },
    };
    if (model.artifacts.containerfile) {
      subset.containerfile = { artifactKey: model.artifacts.containerfile };
    }
  }
  if (model.build.engine === "dockerfile" && model.build.dockerfile) {
    subset.build = {
      engine: model.build.engine,
      dockerfile: {
        path: model.build.dockerfile.path,
        imageTarget: model.build.dockerfile.imageTarget,
        testTarget: model.build.dockerfile.testTarget,
        cacheRefTemplate: model.build.dockerfile.cacheRefTemplate,
      },
    };
  }
  return subset;
}

export function validateGpContentClient(model: GpContentModel, componentName: string): string[] {
  const issues: string[] = [];
  if (model.schemaVersion !== 2) {
    issues.push("schemaVersion должен быть 2");
  }
  if (model.name !== componentName) {
    issues.push(`content.yaml name (${model.name}) должен совпадать с именем компонента (${componentName})`);
  }
  if (model.kind !== "gp-content") {
    issues.push("kind должен быть gp-content");
  }
  if (!model.build.engine) {
    issues.push("build.engine обязателен");
  }
  if (model.capabilities.deliverables.length === 0) {
    issues.push("capabilities.deliverables не может быть пустым");
  }
  const hasArtifact = model.capabilities.deliverables.includes("artifact");
  if (model.build.engine === "buildkit") {
    if (!model.build.buildkit?.cacheRefTemplate.trim()) {
      issues.push("build.buildkit.cacheRefTemplate обязателен");
    }
    if (!model.artifacts.containerfile?.trim()) {
      issues.push("artifacts.containerfile обязателен для buildkit");
    }
  }
  if (model.build.engine === "dockerfile") {
    if (!model.build.dockerfile?.path.trim()) {
      issues.push("build.dockerfile.path обязателен");
    }
    if (!model.build.dockerfile?.cacheRefTemplate.trim()) {
      issues.push("build.dockerfile.cacheRefTemplate обязателен");
    }
    if (hasArtifact) {
      issues.push("artifact deliverable не поддерживается для dockerfile engine");
    }
    if (model.artifacts.containerfile) {
      issues.push("artifacts.containerfile не должен быть задан для BYO dockerfile");
    }
  }
  if (model.pipeline.stages.length === 0) {
    issues.push("pipeline.stages не может быть пустым");
  }
  for (const st of model.pipeline.stages) {
    if (!st.id.trim() || !st.name.trim()) {
      issues.push("каждый stage должен иметь id и name");
      break;
    }
  }
  if (!model.artifacts.validateSchema.trim()) {
    issues.push("artifacts.validateSchema обязателен");
  }
  return issues;
}
