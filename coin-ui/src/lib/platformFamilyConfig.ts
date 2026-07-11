export type PlatformFamilyId = "runtime" | "build-stacks" | "branching-models";

export type PlatformFamilyConfig = {
  id: PlatformFamilyId;
  segment: string;
  compType: string;
  catalogTitle: string;
  profileLabel: string;
  catalogDescription: string;
  hint?: string;
  runbookHref?: string;
};

export const PLATFORM_FAMILIES: Record<PlatformFamilyId, PlatformFamilyConfig> = {
  runtime: {
    id: "runtime",
    segment: "runtime",
    compType: "agent",
    catalogTitle: "Runtime",
    profileLabel: "Agent stack",
    catalogDescription: "Agent stack profiles — CI pod images registered after Nexus push",
    hint: "Agent pin в GP draft composition — только published versions.",
    runbookHref: "/docs/agent-build-model.md",
  },
  "build-stacks": {
    id: "build-stacks",
    segment: "build-stacks",
    compType: "gp-content",
    catalogTitle: "Build stacks",
    profileLabel: "Build stack",
    catalogDescription: "gp-content — Dockerfile, scripts и schema для каждого GP profile",
  },
  "branching-models": {
    id: "branching-models",
    segment: "branching-models",
    compType: "branching-model",
    catalogTitle: "Branching models",
    profileLabel: "Branching model",
    catalogDescription: "Правила ветвления и публикации для GP (model.yaml)",
  },
};

export function familyByCompType(compType: string): PlatformFamilyConfig | undefined {
  return Object.values(PLATFORM_FAMILIES).find((f) => f.compType === compType);
}

export function familyCatalogPath(id: PlatformFamilyId): string {
  return `/platform/${PLATFORM_FAMILIES[id].segment}`;
}

export function familyHubPath(id: PlatformFamilyId, name: string): string {
  return `${familyCatalogPath(id)}/${encodeURIComponent(name)}`;
}

export function familyNewProfilePath(id: PlatformFamilyId): string {
  return `${familyCatalogPath(id)}/new`;
}

export function familyReleaseDetailPath(id: PlatformFamilyId, name: string, version: string): string {
  return `${familyHubPath(id, name)}/releases/${encodeURIComponent(version)}`;
}

export function familyNewDraftPath(id: PlatformFamilyId, name: string): string {
  return `${familyHubPath(id, name)}/releases/new-draft`;
}
