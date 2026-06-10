export type DashboardStats = {
  projects: number;
  gpReleases: number;
  buildReports: number;
  goldenPaths: number;
};

export type Project = {
  name: string;
  goldenPath: string;
  version: string;
  canaryMode?: string;
  gitUrl?: string;
  lastSeenAt: string;
};

export type GPRelease = {
  name: string;
  version: string;
  status: string;
  manifestHash?: string;
  manifestUrl?: string;
  createdAt: string;
};

export type CompositionItem = {
  type: string;
  name: string;
  version: string;
};

export type GPReleaseDetail = GPRelease & {
  composition: CompositionItem[];
};

export type Component = {
  type: string;
  name: string;
  latestVersion: string;
  versionCount: number;
  latestCreatedAt: string;
};

export type BlastRadius = {
  goldenPath: string;
  version: string;
  onThisVersion: number;
  onOtherVersions: number;
  onOlderVersions: number;
  totalOnGP: number;
  byVersion: { version: string; count: number }[];
};

export type ListResponse<T> = { items: T[] };

export type GPProfileSlot = {
  key: string;
  type: string;
  name: string;
};

export type GPProfile = {
  name: string;
  slots: GPProfileSlot[];
};

export type ComponentVersion = {
  version: string;
  status: string;
  createdAt: string;
};

export type AuditLogEntry = {
  id: number;
  action: string;
  entityType: string;
  entityKey: string;
  actor?: string;
  payload: Record<string, unknown>;
  createdAt: string;
};

export type PublishGPResult = {
  id: number;
  name: string;
  version: string;
  status: string;
  manifestHash?: string;
  manifestUrl?: string;
  resolvedVersion?: string;
};

export type DraftGPResult = {
  id: number;
  name: string;
  version: string;
  status: string;
};

export type ArtifactMeta = {
  key: string;
  sha256: string;
  size: number;
};

export type ArtifactDetail = {
  key: string;
  sha256: string;
  body: string;
};

export type CatalogPolicy = {
  gpName: string;
  latest: string;
  latestCanary: string;
  minimum: string;
  deprecated: string[];
};

export type PointerStatus = {
  pin: string;
  resolvedVersion: string;
  manifestHash?: string;
};

export type CatalogOverview = {
  catalog: CatalogPolicy;
  pointers: PointerStatus[];
};

export type ResolvePreviewResult = {
  requestedPin: string;
  resolvedVersion: string;
  channel: string;
  manifest: Record<string, unknown>;
};

export type CanaryPolicy = {
  gpName: string;
  enabled: boolean;
  canaryPercent: number;
  degradedThresholdPct: number;
  criticalThresholdPct: number;
  criticalConsecutiveFailures: number;
};

export type CanaryOverview = {
  policy: CanaryPolicy;
  catalog: CatalogPolicy;
  inCanary: number;
  totalProjects: number;
};

export type FailureEntry = {
  project: string;
  failedStage?: string;
  buildUrl?: string;
  reportedAt: string;
};

export type HealthSummary = {
  gpName: string;
  version: string;
  channel: string;
  windowHours: number;
  successCount: number;
  failureCount: number;
  failureRate: number;
  health: "healthy" | "degraded" | "critical";
  recentFailures: FailureEntry[];
};

export type Role = "admin" | "publisher" | "reader";

export type MeResponse = {
  subject: string;
  email?: string;
  roles: Role[];
  authMethod: string;
};
