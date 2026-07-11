import yaml from "js-yaml";

export type BuildEngine = "buildkit" | "dockerfile";

export type ParameterType = "string" | "boolean" | "number" | "enum";

export type BuildStackParameter = {
  name: string;
  type: ParameterType;
  default?: string | boolean | number;
  required?: boolean;
  allowedValues?: string[];
};

export type PipelineAction = "run" | "build" | "publish";

export type InlineContainerfile = {
  body?: string;
};

export type InlineDockerfile = {
  path: string;
};

export type InlineRun = {
  engine: BuildEngine;
  output?: string;
  containerfile?: InlineContainerfile;
  dockerfile?: InlineDockerfile;
};

export type DeliverableType = "image" | "liquibase-image" | "artifact";

export type InlineBuild = {
  id: string;
  type: DeliverableType;
  engine: BuildEngine;
  containerfile?: InlineContainerfile;
  dockerfile?: InlineDockerfile;
  image?: { repositorySuffix?: string };
  artifact?: { format?: "zip" | "tar"; paths?: string[] };
};

export type InlinePublish = {
  buildStepId: string;
};

export type PipelineStep = {
  action: PipelineAction;
  run?: InlineRun;
  build?: InlineBuild;
  publish?: InlinePublish;
};

export type PipelineStage = {
  id: string;
  name: string;
  steps: PipelineStep[];
};

export type GpContentModel = {
  schemaVersion: 3;
  name: string;
  kind: "golden-path" | "gp-content";
  parameters: BuildStackParameter[];
  validateSchema: string;
  pipeline: {
    stages: PipelineStage[];
  };
};

export type GpContentPreviewResult = {
  valid: boolean;
  issues: Array<{ field: string; message: string }>;
  warnings?: string[];
  parameters?: BuildStackParameter[];
  pipeline?: Record<string, unknown>;
  artifacts?: Record<string, unknown>;
};


export const GP_CONTENT_ARTIFACT = "content.yaml";
export const GP_SCHEMA_ARTIFACT = "schemas/config.v2.schema.json";

/** Short hash id: 5–6 lowercase alphanumeric chars (a-z, 0-9). */
export const SHORT_ID_PATTERN = /^[a-z0-9]{5,6}$/;

const SHORT_ID_ALPHABET = "abcdefghijklmnopqrstuvwxyz0123456789";

export function isShortId(value: string): boolean {
  return SHORT_ID_PATTERN.test(value.trim());
}

export function generateShortId(length = 6): string {
  const size = Math.min(6, Math.max(5, length));
  const bytes = new Uint8Array(size);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (b) => SHORT_ID_ALPHABET[b % SHORT_ID_ALPHABET.length]).join("");
}

export function generateUniqueShortId(existing: Iterable<string>, length = 6): string {
  const taken = new Set(Array.from(existing, (id) => id.trim()).filter(Boolean));
  for (let attempt = 0; attempt < 64; attempt++) {
    const id = generateShortId(length);
    if (!taken.has(id)) return id;
  }
  throw new Error("failed to generate unique short id");
}

export function collectStageIds(model: GpContentModel): string[] {
  return (model.pipeline?.stages ?? []).map((stage) => stage.id);
}

export function defaultContainerfileBody(): string {
  return `# Coin managed Containerfile: Go build targets for BuildKit engine.
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

function buildkitRunStep(output: string, body: string): PipelineStep {
  return {
    action: "run",
    run: { engine: "buildkit", output, containerfile: { body } },
  };
}

function buildkitBuildStep(id: string, type: DeliverableType, body: string, artifact?: InlineBuild["artifact"]): PipelineStep {
  return {
    action: "build",
    build: { id, type, engine: "buildkit", containerfile: { body }, artifact },
  };
}

export function defaultGpContent(name: string, preset: "buildkit" | "dockerfile" = "buildkit"): GpContentModel {
  if (preset === "dockerfile") {
    return defaultGpContentDocker(name);
  }
  const body = defaultContainerfileBody();
  const taken = new Set<string>();
  const stage = (label: string, steps: PipelineStep[]): PipelineStage => ({
    id: generateUniqueShortId(taken),
    name: label,
    steps,
  });
  const imageBuildId = generateUniqueShortId(taken);
  const zipBuildId = generateUniqueShortId(taken);
  return {
    schemaVersion: 3,
    name,
    kind: "gp-content",
    parameters: [{ name: "GO_VERSION", type: "string", default: "1.22", required: true }],
    validateSchema: GP_SCHEMA_ARTIFACT,
    pipeline: {
      stages: [
        stage("Validate", [buildkitRunStep("validate", body)]),
        stage("Test", [buildkitRunStep("test", body)]),
        stage("Build", [
          buildkitBuildStep(imageBuildId, "image", body),
          buildkitBuildStep(zipBuildId, "artifact", body, { format: "zip", paths: ["/out/app"] }),
        ]),
        stage("Publish", [
          { action: "publish", publish: { buildStepId: imageBuildId } },
          { action: "publish", publish: { buildStepId: zipBuildId } },
        ]),
      ],
    },
  };
}

export function defaultGpContentDocker(name: string): GpContentModel {
  const taken = new Set<string>();
  const stage = (label: string, steps: PipelineStep[]): PipelineStage => ({
    id: generateUniqueShortId(taken),
    name: label,
    steps,
  });
  const appBuildId = generateUniqueShortId(taken);
  return {
    schemaVersion: 3,
    name,
    kind: "gp-content",
    parameters: [],
    validateSchema: GP_SCHEMA_ARTIFACT,
    pipeline: {
      stages: [
        stage("Validate", [
          {
            action: "run",
            run: { engine: "dockerfile", dockerfile: { path: "Dockerfile" }, output: "validate" },
          },
        ]),
        stage("Test", [
          {
            action: "run",
            run: { engine: "dockerfile", dockerfile: { path: "Dockerfile" }, output: "test" },
          },
        ]),
        stage("Build", [
          {
            action: "build",
            build: {
              id: appBuildId,
              type: "image",
              engine: "dockerfile",
              dockerfile: { path: "Dockerfile" },
            },
          },
        ]),
        stage("Publish", [{ action: "publish", publish: { buildStepId: appBuildId } }]),
      ],
    },
  };
}

function yamlScalar(value: string): string {
  if (/^[a-zA-Z0-9._-]+$/.test(value)) return value;
  return JSON.stringify(value);
}

export function serializeGpContent(model: GpContentModel): string {
  const lines = [
    `schemaVersion: ${model.schemaVersion}`,
    `name: ${model.name}`,
    `kind: ${model.kind}`,
    `validateSchema: ${model.validateSchema}`,
    "",
    "parameters:",
  ];
  for (const param of model.parameters) {
    lines.push(`  - name: ${param.name}`, `    type: ${param.type}`);
    if (param.default !== undefined) {
      const def = param.default;
      lines.push(`    default: ${typeof def === "string" ? yamlScalar(def) : String(def)}`);
    }
    if (param.required) lines.push("    required: true");
    if (param.allowedValues?.length) {
      lines.push("    allowedValues:");
      for (const value of param.allowedValues) lines.push(`      - ${yamlScalar(value)}`);
    }
  }
  lines.push("pipeline:", "  stages:");
  for (const stage of model.pipeline.stages) {
    lines.push(`    - id: ${stage.id}`, `      name: ${yamlScalar(stage.name)}`, "      steps:");
    for (const step of stage.steps) {
      lines.push(`        - action: ${step.action}`);
      if (step.run) {
        lines.push("          run:", `            engine: ${step.run.engine}`);
        if (step.run.output) lines.push(`            output: ${step.run.output}`);
        if (step.run.containerfile?.body) {
          lines.push("            containerfile:", "              body: |");
          for (const line of step.run.containerfile.body.split("\n")) lines.push(`                ${line}`);
        }
        if (step.run.dockerfile) {
          lines.push("            dockerfile:", `              path: ${step.run.dockerfile.path}`);
        }
      }
      if (step.build) {
        lines.push("          build:", `            id: ${step.build.id}`, `            type: ${step.build.type}`, `            engine: ${step.build.engine}`);
        if (step.build.containerfile?.body) {
          lines.push("            containerfile:", "              body: |");
          for (const line of step.build.containerfile.body.split("\n")) lines.push(`                ${line}`);
        }
        if (step.build.dockerfile) {
          lines.push("            dockerfile:", `              path: ${step.build.dockerfile.path}`);
        }
        if (step.build.image?.repositorySuffix) {
          lines.push("            image:", `              repositorySuffix: ${yamlScalar(step.build.image.repositorySuffix)}`);
        }
        if (step.build.artifact) {
          lines.push("            artifact:");
          if (step.build.artifact.format) lines.push(`              format: ${step.build.artifact.format}`);
          if (step.build.artifact.paths?.length) {
            lines.push("              paths:");
            for (const p of step.build.artifact.paths) lines.push(`                - ${p}`);
          }
        }
      }
      if (step.publish) {
        lines.push("          publish:", `            buildStepId: ${step.publish.buildStepId}`);
      }
    }
  }
  return `${lines.join("\n")}\n`;
}

export function parseGpContentYaml(raw: string, fallbackName: string): GpContentModel {
  try {
    const doc = yaml.load(raw) as Record<string, unknown> | null;
    if (!doc || typeof doc !== "object") return defaultGpContent(fallbackName);
    if (doc.schemaVersion !== 3) return defaultGpContent(fallbackName);

    const parameters = Array.isArray(doc.parameters)
      ? (doc.parameters as Record<string, unknown>[]).map((p) => ({
          name: String(p.name ?? ""),
          type: String(p.type ?? "string") as ParameterType,
          default: p.default as string | number | boolean | undefined,
          required: Boolean(p.required),
          allowedValues: Array.isArray(p.allowedValues) ? (p.allowedValues as string[]) : undefined,
        }))
      : [];

    const pipelineDoc = doc.pipeline as { stages?: unknown[] } | undefined;
    const stages = Array.isArray(pipelineDoc?.stages)
      ? (pipelineDoc!.stages as Record<string, unknown>[]).map(parseStage)
      : [];

    return {
      schemaVersion: 3,
      name: String(doc.name ?? fallbackName),
      kind: "gp-content",
      validateSchema: String(doc.validateSchema ?? GP_SCHEMA_ARTIFACT),
      parameters,
      pipeline: { stages },
    };
  } catch {
    return defaultGpContent(fallbackName);
  }
}

function parseContainerfile(block: unknown): InlineContainerfile | undefined {
  if (!block || typeof block !== "object") return undefined;
  const b = block as Record<string, unknown>;
  if (typeof b.body === "string") return { body: b.body };
  return undefined;
}

function parseDockerfile(block: unknown): InlineDockerfile | undefined {
  if (!block || typeof block !== "object") return undefined;
  const b = block as Record<string, unknown>;
  const path = typeof b.path === "string" ? b.path : undefined;
  if (!path) return undefined;
  return { path };
}

function parseRun(block: unknown): InlineRun | undefined {
  if (!block || typeof block !== "object") return undefined;
  const b = block as Record<string, unknown>;
  return {
    engine: String(b.engine ?? "buildkit") as BuildEngine,
    output: typeof b.output === "string" ? b.output : undefined,
    containerfile: parseContainerfile(b.containerfile),
    dockerfile: parseDockerfile(b.dockerfile),
  };
}

function parseBuild(block: unknown): InlineBuild | undefined {
  if (!block || typeof block !== "object") return undefined;
  const b = block as Record<string, unknown>;
  const artifactBlock = b.artifact as Record<string, unknown> | undefined;
  const imageBlock = b.image as Record<string, unknown> | undefined;
  return {
    id: String(b.id ?? ""),
    type: String(b.type ?? "image") as DeliverableType,
    engine: String(b.engine ?? "buildkit") as BuildEngine,
    containerfile: parseContainerfile(b.containerfile),
    dockerfile: parseDockerfile(b.dockerfile),
    artifact: artifactBlock
      ? {
          format:
            artifactBlock.format === "zip" || artifactBlock.format === "tar"
              ? artifactBlock.format
              : undefined,
          paths: Array.isArray(artifactBlock.paths) ? (artifactBlock.paths as string[]) : undefined,
        }
      : undefined,
    image: imageBlock?.repositorySuffix
      ? { repositorySuffix: String(imageBlock.repositorySuffix) }
      : undefined,
  };
}

function parsePublish(block: unknown): InlinePublish | undefined {
  if (!block || typeof block !== "object") return undefined;
  const b = block as Record<string, unknown>;
  const buildStepId = typeof b.buildStepId === "string" ? b.buildStepId : "";
  if (!buildStepId) return undefined;
  return { buildStepId };
}

function parseStep(block: unknown): PipelineStep {
  const b = (block ?? {}) as Record<string, unknown>;
  const action = String(b.action ?? "run") as PipelineStep["action"];
  return {
    action,
    run: b.run ? parseRun(b.run) : undefined,
    build: b.build ? parseBuild(b.build) : undefined,
    publish: b.publish ? parsePublish(b.publish) : undefined,
  };
}

function parseStage(block: Record<string, unknown>): PipelineStage {
  return {
    id: String(block.id ?? ""),
    name: String(block.name ?? ""),
    steps: Array.isArray(block.steps) ? (block.steps as unknown[]).map(parseStep) : [],
  };
}

export function buildGpContentManifestSubset(model: GpContentModel): Record<string, unknown> {
  return {
    schemaVersion: model.schemaVersion,
    parameters: model.parameters,
    validateSchema: { artifactKey: model.validateSchema },
    pipeline: { stages: model.pipeline.stages },
  };
}

export function collectBuildIds(model: GpContentModel): string[] {
  const ids: string[] = [];
  for (const stage of model.pipeline.stages) {
    for (const step of stage.steps) {
      if (step.action === "build" && step.build?.id) ids.push(step.build.id);
    }
  }
  return ids;
}

export function validateGpContentClient(model: GpContentModel, componentName: string): string[] {
  const issues: string[] = [];
  if (model.schemaVersion !== 3) issues.push("schemaVersion должен быть 3");
  if (model.name !== componentName) issues.push(`name (${model.name}) должен совпадать с ${componentName}`);
  if (model.kind !== "gp-content") issues.push("kind должен быть gp-content");
  if (!model.validateSchema.trim()) issues.push("validateSchema обязателен");
  if (model.pipeline.stages.length === 0) issues.push("pipeline.stages не может быть пустым");

  const buildIds = new Set<string>();
  for (const parameter of model.parameters) {
    if (/(SECRET|PASSWORD|PASSWD|TOKEN|CREDENTIAL)/i.test(parameter.name)) {
      issues.push(`parameter ${parameter.name} похож на secret/credential`);
    }
    if (parameter.type === "enum" && !parameter.allowedValues?.length) {
      issues.push(`parameter ${parameter.name} требует allowedValues`);
    }
  }

  for (const stage of model.pipeline.stages) {
    if (!stage.id.trim() || !stage.name.trim()) issues.push("каждый stage должен иметь id и name");
    else if (!isShortId(stage.id)) issues.push(`stage.id ${stage.id} должен быть short hash (5–6 символов a-z0-9)`);
    if (stage.steps.length === 0) issues.push(`stage ${stage.id} должен иметь steps`);
    for (const step of stage.steps) {
      if (step.action === "run" && step.run?.engine === "buildkit" && !step.run.containerfile?.body?.trim()) {
        issues.push(`stage ${stage.id}: run step требует containerfile.body`);
      }
      if (step.action === "build") {
        if (!step.build?.id.trim()) issues.push(`stage ${stage.id}: build.id обязателен`);
        else if (!isShortId(step.build.id)) issues.push(`build.id ${step.build.id} должен быть short hash (5–6 символов a-z0-9)`);
        else if (buildIds.has(step.build.id)) issues.push(`дубликат build.id ${step.build.id}`);
        else buildIds.add(step.build.id);
        if (step.build?.engine === "buildkit" && !step.build.containerfile?.body?.trim()) {
          issues.push(`stage ${stage.id}: build step требует containerfile.body`);
        }
      }
      if (step.action === "publish") {
        if (!step.publish?.buildStepId || !buildIds.has(step.publish.buildStepId)) {
          issues.push(`stage ${stage.id}: publish ссылается на отсутствующий buildStepId`);
        }
      }
    }
  }
  return issues;
}

export function emptyGpContentModel(name: string): GpContentModel {
  return {
    schemaVersion: 3,
    name,
    kind: "golden-path",
    parameters: [],
    validateSchema: GP_SCHEMA_ARTIFACT,
    pipeline: { stages: [] },
  };
}

export function pipelineBodyToModel(
  body: { schemaVersion?: number; body?: unknown } | Record<string, unknown>,
  gpName: string,
): GpContentModel {
  const raw = ("body" in body ? body.body : body) as Record<string, unknown>;
  const parameters = normalizeParameters(raw.parameters);
  const validateSchema = String(raw.validateSchema ?? GP_SCHEMA_ARTIFACT);
  const pipelineRaw = (raw.pipeline ?? {}) as Record<string, unknown>;
  const stages = normalizePipelineStages(pipelineRaw.stages ?? pipelineRaw.Stages);
  return {
    schemaVersion: 3,
    name: gpName,
    kind: "golden-path",
    parameters,
    validateSchema,
    pipeline: { stages },
  };
}

function normalizeParameters(raw: unknown): BuildStackParameter[] {
  if (!Array.isArray(raw)) return [];
  return raw.map((item) => {
    const p = item as Record<string, unknown>;
    const defaultVal = p.default ?? p.Default;
    return {
      name: String(p.name ?? p.Name ?? ""),
      type: (p.type ?? p.Type ?? "string") as ParameterType,
      default:
        typeof defaultVal === "string" ||
        typeof defaultVal === "number" ||
        typeof defaultVal === "boolean"
          ? defaultVal
          : undefined,
      required: Boolean(p.required ?? p.Required),
      allowedValues: (p.allowedValues ?? p.AllowedValues) as string[] | undefined,
    };
  });
}

function normalizePipelineStages(raw: unknown): PipelineStage[] {
  if (!Array.isArray(raw)) return [];
  return raw.map((item) => {
    const stage = item as Record<string, unknown>;
    const stepsRaw = stage.steps ?? stage.Steps;
    const steps = Array.isArray(stepsRaw) ? stepsRaw.map(normalizePipelineStep) : [];
    return {
      id: String(stage.id ?? stage.ID ?? ""),
      name: String(stage.name ?? stage.Name ?? ""),
      steps,
    };
  });
}

function normalizePipelineStep(item: unknown): PipelineStep {
  const step = item as Record<string, unknown>;
  const action = String(step.action ?? step.Action ?? "run") as PipelineAction;
  const run = (step.run ?? step.Run) as InlineRun | undefined;
  const build = (step.build ?? step.Build) as InlineBuild | undefined;
  const publish = (step.publish ?? step.Publish) as InlinePublish | undefined;
  return { action, run, build, publish };
}
