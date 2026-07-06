export type BuildReport = {
  id: number;
  project: string;
  goldenPath: string;
  version: string;
  resolvedVersion?: string;
  branch?: string;
  buildUrl?: string;
  result: string;
  channel?: string;
  failedStage?: string;
  reportedAt: string;
};

export type DashboardStats = {
  projects: number;
  staleProjects: number;
  gpReleases: number;
  buildReports: number;
  goldenPaths: number;
};

export type Project = {
  name: string;
  groupId?: string;
  artifactId?: string;
  gitRepoName?: string;
  gitRepoUrl?: string;
  goldenPath: string;
  version: string;
  canaryMode?: string;
  branch?: string;
  lastBuildAt?: string;
};

export type GPRelease = {
  name: string;
  version: string;
  status: string;
  destinations: GPDestinations;
  manifestHash?: string;
  manifestUrl?: string;
  createdAt: string;
};

export type GPDestinations = {
  imageRegistryPrefix: string;
  buildCacheEnabled: boolean;
  artifactRepositoryBase: string;
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

export type PaginatedListResponse<T> = {
  items: T[];
  total: number;
  limit: number;
  offset: number;
};

export type GPProfile = {
  name: string;
  description?: string;
  createdAt: string;
};

export type ComponentPin = {
  type: string;
  name: string;
  version: string;
};

export type ComponentVersion = {
  version: string;
  status: string;
  createdAt: string;
};

export type ComponentDetail = Component & {
  gpUsage: ComponentGPUsage[];
};

export type ComponentGPUsage = {
  gpName: string;
  version: string;
  status: string;
};

export type ComponentVersionDetail = ComponentVersion & {
  type: string;
  name: string;
  metadata: Record<string, unknown>;
  contentRef?: Record<string, unknown>;
};

export type ValidationIssue = {
  field: string;
  message: string;
};

export type ValidateComponentPackageResult = {
  valid: boolean;
  issues: ValidationIssue[];
};

export type RegisterComponentPackageResult = {
  type: string;
  name: string;
  version: string;
  packageUrl?: string;
  packageSha256?: string;
  filesUploaded: number;
  contentRef: Record<string, unknown>;
};

export type DraftComponentResult = {
  id: number;
  componentId: number;
  type: string;
  name: string;
  version: string;
  status: string;
};

export type CanaryContext = {
  project: string;
  gpName: string;
  canaryMode: string;
  rolloutEnabled: boolean;
  canaryPercent: number;
  projectBucket: number;
  useCanaryLine: boolean;
  stableVersion: string;
  canaryVersion: string;
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
  audience?: string;
  line?: string;
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
  canaryContext?: CanaryContext;
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
