export type PipelineStage = {
  id: string;
  name: string;
  when?: string;
};

export type GpContentModel = {
  name: string;
  kind: "gp-content";
  build: {
    engine: "buildkit" | "buildpack" | "dockerfile";
    buildkit?: {
      dockerfile: string;
      targets: Record<string, string>;
      cacheRefTemplate: string;
    };
  };
  pipeline: {
    stages: PipelineStage[];
  };
  validateSchema: {
    artifactKey: string;
  };
  containerfile: {
    artifactKey: string;
  };
};

export const GP_CONTENT_ARTIFACT = "content.yaml";
export const GP_CONTAINERFILE_ARTIFACT = "dockerfiles/Containerfile";
export const GP_SCHEMA_ARTIFACT = "schemas/config.v2.schema.json";

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

export function defaultGpContent(name: string): GpContentModel {
  return {
    name,
    kind: "gp-content",
    build: {
      engine: "buildkit",
      buildkit: {
        dockerfile: GP_CONTAINERFILE_ARTIFACT,
        targets: {
          validate: "validate",
          test: "test",
          image: "runtime",
          artifact: "artifact",
        },
        cacheRefTemplate: "{{registryHost}}/coin-cache/{{project}}:buildkit",
      },
    },
    pipeline: {
      stages: [
        { id: "validate", name: "Validate" },
        { id: "test", name: "Test" },
        { id: "build", name: "Build" },
        { id: "publish", name: "Publish", when: "tag" },
      ],
    },
    validateSchema: { artifactKey: GP_SCHEMA_ARTIFACT },
    containerfile: { artifactKey: GP_CONTAINERFILE_ARTIFACT },
  };
}

export function serializeGpContent(model: GpContentModel): string {
  const lines = [
    `name: ${model.name}`,
    `kind: ${model.kind}`,
    "build:",
    `  engine: ${model.build.engine}`,
  ];
  if (model.build.engine === "buildkit" && model.build.buildkit) {
    const bk = model.build.buildkit;
    lines.push("  buildkit:");
    lines.push(`    dockerfile: ${bk.dockerfile}`);
    lines.push("    targets:");
    for (const [k, v] of Object.entries(bk.targets)) {
      lines.push(`      ${k}: ${v}`);
    }
    lines.push(`    cacheRefTemplate: ${bk.cacheRefTemplate}`);
  }
  lines.push("pipeline:");
  lines.push("  stages:");
  for (const st of model.pipeline.stages) {
    lines.push(`    - id: ${st.id}`);
    lines.push(`      name: ${st.name}`);
    if (st.when) {
      lines.push(`      when: ${st.when}`);
    }
  }
  lines.push("validateSchema:");
  lines.push(`  artifactKey: ${model.validateSchema.artifactKey}`);
  lines.push("containerfile:");
  lines.push(`  artifactKey: ${model.containerfile.artifactKey}`);
  return `${lines.join("\n")}\n`;
}

/** Minimal YAML parser for Component Studio gp-content scaffold. */
export function parseGpContentYaml(raw: string, fallbackName: string): GpContentModel {
  const model = defaultGpContent(fallbackName);
  const lines = raw.split(/\r?\n/);
  let inStages = false;
  let inTargets = false;
  let currentStage: PipelineStage | null = null;
  const stages: PipelineStage[] = [];
  const targets: Record<string, string> = {};

  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;

    if (trimmed === "stages:") {
      inStages = true;
      continue;
    }
    if (inStages && trimmed.startsWith("- id:")) {
      if (currentStage) stages.push(currentStage);
      currentStage = { id: trimmed.slice(6).trim(), name: "" };
      continue;
    }
    if (inStages && currentStage && trimmed.startsWith("name:")) {
      currentStage.name = trimmed.slice(5).trim();
      continue;
    }
    if (inStages && currentStage && trimmed.startsWith("when:")) {
      currentStage.when = trimmed.slice(5).trim();
      continue;
    }
    if (inStages && !line.startsWith(" ")) {
      if (currentStage) {
        stages.push(currentStage);
        currentStage = null;
      }
      inStages = false;
    }

    if (trimmed === "targets:") {
      inTargets = true;
      continue;
    }
    if (inTargets) {
      const targetKv = trimmed.match(/^([a-zA-Z]+):\s*(.+)$/);
      if (targetKv) {
        targets[targetKv[1]] = targetKv[2];
        continue;
      }
      if (!line.startsWith(" ")) {
        inTargets = false;
      }
    }

    const kv = trimmed.match(/^([a-zA-Z]+):\s*(.+)$/);
    if (!kv) continue;
    const [, key, value] = kv;

    switch (key) {
      case "name":
        if (!inStages) model.name = value;
        break;
      case "kind":
        model.kind = value as GpContentModel["kind"];
        break;
      case "engine":
        model.build.engine = value as GpContentModel["build"]["engine"];
        break;
      case "dockerfile":
        if (model.build.buildkit) model.build.buildkit.dockerfile = value;
        break;
      case "cacheRefTemplate":
        if (model.build.buildkit) model.build.buildkit.cacheRefTemplate = value;
        break;
      case "artifactKey":
        if (lines.slice(Math.max(0, lines.indexOf(line) - 2), lines.indexOf(line)).join("\n").includes("validateSchema:")) {
          model.validateSchema.artifactKey = value;
        } else {
          model.containerfile.artifactKey = value;
        }
        break;
      default:
        break;
    }
  }
  if (currentStage) stages.push(currentStage);
  if (stages.length > 0) model.pipeline.stages = stages;
  if (Object.keys(targets).length > 0 && model.build.buildkit) {
    model.build.buildkit.targets = targets;
  }
  return model;
}

export function buildGpContentManifestSubset(model: GpContentModel): Record<string, unknown> {
  const subset: Record<string, unknown> = {
    build: model.build,
    pipeline: model.pipeline,
    validateSchema: { artifactKey: model.validateSchema.artifactKey },
    containerfile: { artifactKey: model.containerfile.artifactKey },
  };
  return subset;
}

export function validateGpContentClient(model: GpContentModel, componentName: string): string[] {
  const issues: string[] = [];
  if (model.name !== componentName) {
    issues.push(`content.yaml name (${model.name}) должен совпадать с именем компонента (${componentName})`);
  }
  if (model.kind !== "gp-content") {
    issues.push("kind должен быть gp-content");
  }
  if (!model.build.engine) {
    issues.push("build.engine обязателен");
  }
  if (model.build.engine === "buildkit") {
    if (!model.build.buildkit?.dockerfile.trim()) {
      issues.push("build.buildkit.dockerfile обязателен для buildkit");
    }
    if (!model.build.buildkit?.cacheRefTemplate.trim()) {
      issues.push("build.buildkit.cacheRefTemplate обязателен");
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
  if (!model.validateSchema.artifactKey.trim()) {
    issues.push("validateSchema.artifactKey обязателен");
  }
  if (!model.containerfile.artifactKey.trim()) {
    issues.push("containerfile.artifactKey обязателен");
  }
  return issues;
}
