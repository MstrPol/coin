import type {
  ArtifactDetail,
  ArtifactMeta,
  AuditLogEntry,
  BlastRadius,
  BuildReport,
  CanaryContext,
  CanaryOverview,
  CatalogOverview,
  Component,
  ComponentDetail,
  ComponentVersion,
  ComponentVersionDetail,
  DraftComponentResult,
  DashboardStats,
  DraftGPResult,
  GPProfile,
  GPRelease,
  GPReleaseDetail,
  HealthSummary,
  ListResponse,
  MeResponse,
  PaginatedListResponse,
  Project,
  PublishGPResult,
  RegisterComponentPackageResult,
  ResolvePreviewResult,
  ValidateComponentPackageResult,
} from "../api/types";

export type CompositionPinBlocker = {
  type: string;
  name: string;
  version: string;
  status: string;
};

export class PromoteBlockedError extends Error {
  blockingPins: CompositionPinBlocker[];

  constructor(message: string, blockingPins: CompositionPinBlocker[]) {
    super(message);
    this.name = "PromoteBlockedError";
    this.blockingPins = blockingPins;
  }
}

const base = import.meta.env.VITE_API_BASE ?? "/api";
const KEY_STORAGE = "coin-admin-api-key";
const TOKEN_STORAGE = "coin-access-token";
const ACTOR_STORAGE = "coin-actor";

export function getApiKey(): string | null {
  return localStorage.getItem(KEY_STORAGE);
}

export function setApiKey(key: string): void {
  localStorage.setItem(KEY_STORAGE, key);
  localStorage.removeItem(TOKEN_STORAGE);
}

export function clearApiKey(): void {
  localStorage.removeItem(KEY_STORAGE);
}

export function getAccessToken(): string | null {
  return localStorage.getItem(TOKEN_STORAGE);
}

export function setAccessToken(token: string): void {
  localStorage.setItem(TOKEN_STORAGE, token);
  localStorage.removeItem(KEY_STORAGE);
}

export function clearAccessToken(): void {
  localStorage.removeItem(TOKEN_STORAGE);
}

export function clearAuthCredentials(): void {
  clearApiKey();
  clearAccessToken();
}

export function getActor(): string {
  return localStorage.getItem(ACTOR_STORAGE) ?? "";
}

export function setActor(actor: string): void {
  localStorage.setItem(ACTOR_STORAGE, actor);
}

function headers(): HeadersInit {
  const h: Record<string, string> = { Accept: "application/json" };
  const token = getAccessToken();
  const key = getApiKey();
  if (token) {
    h.Authorization = `Bearer ${token}`;
  } else if (key) {
    h["X-API-Key"] = key;
  }
  return h;
}

async function parseError(res: Response, path: string): Promise<string> {
  try {
    const data = (await res.json()) as { error?: string };
    if (typeof data.error === "string") return data.error;
  } catch {
    /* ignore */
  }
  return `${res.status} ${path}`;
}

async function apiGet<T>(path: string): Promise<T> {
  const res = await fetch(`${base}${path}`, { headers: headers() });
  if (!res.ok) {
    throw new Error(await parseError(res, path));
  }
  return res.json() as Promise<T>;
}

async function apiList<T>(path: string): Promise<ListResponse<T>> {
  const data = await apiGet<ListResponse<T>>(path);
  return { items: data.items ?? [] };
}

async function apiPaginatedList<T>(path: string): Promise<PaginatedListResponse<T>> {
  const data = await apiGet<PaginatedListResponse<T>>(path);
  return {
    items: data.items ?? [],
    total: data.total ?? 0,
    limit: data.limit ?? 50,
    offset: data.offset ?? 0,
  };
}

export async function downloadCsv(path: string, filename: string): Promise<void> {
  const res = await fetch(`${base}${path}`, {
    headers: {
      ...headers(),
      Accept: "text/csv",
    },
  });
  if (!res.ok) {
    throw new Error(await parseError(res, path));
  }
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

/** Like apiList but returns empty items on 404 (unknown component). */
async function apiListOptional<T>(path: string): Promise<ListResponse<T> | null> {
  const res = await fetch(`${base}${path}`, { headers: headers() });
  if (res.status === 404) {
    return null;
  }
  if (!res.ok) {
    throw new Error(await parseError(res, path));
  }
  const data = (await res.json()) as ListResponse<T>;
  return { items: data.items ?? [] };
}

async function apiPost<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${base}${path}`, {
    method: "POST",
    headers: { ...headers(), "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    throw new Error(await parseError(res, path));
  }
  return res.json() as Promise<T>;
}

async function apiDelete(path: string): Promise<void> {
  const res = await fetch(`${base}${path}`, {
    method: "DELETE",
    headers: headers(),
  });
  if (!res.ok) {
    throw new Error(await parseError(res, path));
  }
}

async function apiPatch<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${base}${path}`, {
    method: "PATCH",
    headers: { ...headers(), "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    throw new Error(await parseError(res, path));
  }
  return res.json() as Promise<T>;
}

async function apiPut<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${base}${path}`, {
    method: "PUT",
    headers: { ...headers(), "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    throw new Error(await parseError(res, path));
  }
  return res.json() as Promise<T>;
}

export const api = {
  ready: () => apiGet<{ status: string; version?: string }>("/ready"),
  me: () => apiGet<MeResponse>("/v1/admin/me"),
  stats: () => apiGet<DashboardStats>("/v1/admin/stats"),
  buildReports: (params?: {
    project?: string;
    goldenPath?: string;
    result?: string;
    reportedAfter?: string;
    reportedBefore?: string;
    limit?: number;
    offset?: number;
  }) => {
    const q = new URLSearchParams();
    if (params?.project) q.set("project", params.project);
    if (params?.goldenPath) q.set("goldenPath", params.goldenPath);
    if (params?.result) q.set("result", params.result);
    if (params?.reportedAfter) q.set("reportedAfter", params.reportedAfter);
    if (params?.reportedBefore) q.set("reportedBefore", params.reportedBefore);
    if (params?.limit != null) q.set("limit", String(params.limit));
    if (params?.offset != null) q.set("offset", String(params.offset));
    const qs = q.toString();
    return apiPaginatedList<BuildReport>(`/v1/admin/build-reports${qs ? `?${qs}` : ""}`);
  },
  buildReportsExportPath: (params?: {
    project?: string;
    goldenPath?: string;
    result?: string;
    reportedAfter?: string;
    reportedBefore?: string;
  }) => {
    const q = new URLSearchParams();
    if (params?.project) q.set("project", params.project);
    if (params?.goldenPath) q.set("goldenPath", params.goldenPath);
    if (params?.result) q.set("result", params.result);
    if (params?.reportedAfter) q.set("reportedAfter", params.reportedAfter);
    if (params?.reportedBefore) q.set("reportedBefore", params.reportedBefore);
    const qs = q.toString();
    return `/v1/admin/build-reports/export${qs ? `?${qs}` : ""}`;
  },
  projects: (params?: {
    goldenPath?: string;
    version?: string;
    stale?: boolean;
    limit?: number;
    offset?: number;
  }) => {
    const q = new URLSearchParams();
    if (params?.goldenPath) q.set("goldenPath", params.goldenPath);
    if (params?.version) q.set("version", params.version);
    if (params?.stale) q.set("stale", "true");
    if (params?.limit != null) q.set("limit", String(params.limit));
    if (params?.offset != null) q.set("offset", String(params.offset));
    const qs = q.toString();
    return apiPaginatedList<Project>(`/v1/admin/projects${qs ? `?${qs}` : ""}`);
  },
  projectsExportPath: (params?: { goldenPath?: string; version?: string; stale?: boolean }) => {
    const q = new URLSearchParams();
    if (params?.goldenPath) q.set("goldenPath", params.goldenPath);
    if (params?.version) q.set("version", params.version);
    if (params?.stale) q.set("stale", "true");
    const qs = q.toString();
    return `/v1/admin/projects/export${qs ? `?${qs}` : ""}`;
  },
  gpNames: () => apiList<string>("/v1/admin/golden-paths/names"),
  gpProfile: (name: string) => apiGet<GPProfile>(`/v1/admin/golden-paths/${name}/profile`),
  createGPProfile: (body: {
    name: string;
    description?: string;
    actor?: string;
  }) => apiPost<{ status: string; name: string }>("/v1/admin/golden-paths/profiles", body),
  gpReleases: (name?: string, includeDrafts = false) => {
    const q = new URLSearchParams();
    if (name) q.set("name", name);
    if (includeDrafts) q.set("includeDrafts", "true");
    const qs = q.toString();
    return apiList<GPRelease>(`/v1/admin/golden-paths${qs ? `?${qs}` : ""}`);
  },
  gpRelease: (name: string, version: string) =>
    apiGet<GPReleaseDetail>(`/v1/admin/golden-paths/${name}/versions/${version}`),
  createDraftGPRelease: (
    name: string,
    body: {
      version: string;
      composition: Record<string, string>;
      agentStackName: string;
      gpContentName: string;
      branchingModelName: string;
      actor?: string;
    },
  ) => apiPost<DraftGPResult>(`/v1/admin/golden-paths/${name}/drafts`, body),
  deleteGPReleaseDraft: (name: string, version: string, actor?: string) => {
    const q = actor ? `?actor=${encodeURIComponent(actor)}` : "";
    return apiDelete(`/v1/admin/golden-paths/${name}/versions/${encodeURIComponent(version)}${q}`);
  },
  updateGPReleaseDraft: (
    name: string,
    version: string,
    body: {
      composition: Record<string, string>;
      agentStackName: string;
      gpContentName: string;
      branchingModelName: string;
      actor?: string;
    },
  ) =>
    apiPatch<{ name: string; version: string; status: string }>(
      `/v1/admin/golden-paths/${name}/versions/${encodeURIComponent(version)}`,
      body,
    ),
  publishGPRelease: (
    name: string,
    body: {
      version: string;
      composition: Record<string, string>;
      agentStackName?: string;
      gpContentName?: string;
      branchingModelName?: string;
      actor?: string;
    },
  ) => apiPost<PublishGPResult>(`/v1/admin/golden-paths/${name}/versions`, body),
  promoteDraftGPRelease: async (name: string, version: string, actor?: string) => {
    const q = actor ? `?actor=${encodeURIComponent(actor)}` : "";
    const path = `/v1/admin/golden-paths/${name}/versions/${encodeURIComponent(version)}/promote${q}`;
    const res = await fetch(`${base}${path}`, {
      method: "POST",
      headers: { ...headers(), "Content-Type": "application/json" },
      body: JSON.stringify({}),
    });
    const data = (await res.json().catch(() => ({}))) as {
      error?: string;
      blockingPins?: CompositionPinBlocker[];
    };
    if (!res.ok) {
      if (res.status === 409 && Array.isArray(data.blockingPins)) {
        throw new PromoteBlockedError(data.error ?? "GP promote blocked", data.blockingPins);
      }
      throw new Error(typeof data.error === "string" ? data.error : `${res.status} ${path}`);
    }
    return data as PublishGPResult;
  },
  catalog: (name: string) =>
    apiGet<CatalogOverview>(`/v1/admin/golden-paths/${name}/catalog`),
  updateCatalog: (
    name: string,
    body: {
      latest: string;
      latestCanary: string;
      minimum: string;
      deprecated: string[];
      actor?: string;
    },
  ) => apiPatch<{ status: string }>(`/v1/admin/golden-paths/${name}/catalog`, body),
  canary: (name: string) =>
    apiGet<CanaryOverview>(`/v1/admin/golden-paths/${name}/canary`),
  updateCanary: (
    name: string,
    body: {
      enabled: boolean;
      canaryPercent: number;
      degradedThresholdPct: number;
      criticalThresholdPct: number;
      criticalConsecutiveFailures: number;
      actor?: string;
    },
  ) => apiPatch<{ status: string }>(`/v1/admin/golden-paths/${name}/canary`, body),
  updateProjectCanaryMode: (projectName: string, mode: string, actor?: string) =>
    apiPatch<{ status: string }>(`/v1/admin/projects/${encodeURIComponent(projectName)}/canary-mode`, {
      mode,
      actor,
    }),
  health: (name: string, version: string, channel = "canary") =>
    apiGet<HealthSummary>(
      `/v1/admin/golden-paths/${name}/versions/${encodeURIComponent(version)}/health?channel=${encodeURIComponent(channel)}`,
    ),
  listArtifacts: (name: string, version: string) =>
    apiList<ArtifactMeta>(
      `/v1/admin/golden-paths/${name}/versions/${encodeURIComponent(version)}/artifacts`,
    ),
  getArtifact: (name: string, version: string, key: string) =>
    apiGet<ArtifactDetail>(
      `/v1/admin/golden-paths/${name}/versions/${encodeURIComponent(version)}/artifacts/${encodeURIComponent(key)}`,
    ),
  saveArtifact: (name: string, version: string, key: string, body: string) =>
    apiPut<{ status: string }>(
      `/v1/admin/golden-paths/${name}/versions/${encodeURIComponent(version)}/artifacts/${encodeURIComponent(key)}`,
      { body },
    ),
  resolvePreview: (
    name: string,
    pin: string,
    project?: string,
    forceChannel?: "canary" | "stable",
  ) => {
    const q = new URLSearchParams({ pin });
    if (project) q.set("project", project);
    if (forceChannel) q.set("forceChannel", forceChannel);
    return apiGet<ResolvePreviewResult>(
      `/v1/admin/golden-paths/${name}/resolve-preview?${q.toString()}`,
    );
  },
  canaryContext: (gpName: string, project: string) =>
    apiGet<CanaryContext>(
      `/v1/admin/golden-paths/${gpName}/projects/${encodeURIComponent(project)}/canary-context`,
    ),
  components: () => apiList<Component>("/v1/admin/components"),
  createComponent: (body: { type: string; name: string; actor?: string }) =>
    apiPost<{ status: string; type: string; name: string }>("/v1/admin/components", body),
  componentDetail: (type: string, name: string) =>
    apiGet<ComponentDetail>(`/v1/admin/components/${type}/${name}`),
  componentVersions: (type: string, name: string) =>
    apiList<ComponentVersion>(`/v1/admin/components/${type}/${name}/versions`),
  componentVersionsOptional: (type: string, name: string) =>
    apiListOptional<ComponentVersion>(`/v1/admin/components/${type}/${name}/versions`),
  componentVersionDetail: (type: string, name: string, version: string) =>
    apiGet<ComponentVersionDetail>(
      `/v1/admin/components/${type}/${name}/versions/${encodeURIComponent(version)}`,
    ),
  publishComponentVersion: (
    type: string,
    name: string,
    body: { version: string; metadata?: Record<string, unknown>; contentRef?: Record<string, unknown>; actor?: string },
  ) =>
    apiPost<{ status: string }>(`/v1/admin/components/${type}/${name}/versions`, body),
  createDraftComponentVersion: (
    type: string,
    name: string,
    body: { version: string; metadata?: Record<string, unknown>; contentRef?: Record<string, unknown>; actor?: string },
  ) =>
    apiPost<DraftComponentResult>(`/v1/admin/components/${type}/${name}/versions/drafts`, body),
  patchComponentVersion: (
    type: string,
    name: string,
    version: string,
    body: { metadata?: Record<string, unknown>; contentRef?: Record<string, unknown>; actor?: string },
  ) =>
    apiPatch<{ status: string }>(
      `/v1/admin/components/${type}/${name}/versions/${encodeURIComponent(version)}`,
      body,
    ),
  listComponentArtifacts: (type: string, name: string, version: string) =>
    apiList<ArtifactMeta>(
      `/v1/admin/components/${type}/${name}/versions/${encodeURIComponent(version)}/artifacts`,
    ),
  getComponentArtifact: (type: string, name: string, version: string, key: string) =>
    apiGet<ArtifactDetail>(
      `/v1/admin/components/${type}/${name}/versions/${encodeURIComponent(version)}/artifacts/${encodeURIComponent(key)}`,
    ),
  saveComponentArtifact: (type: string, name: string, version: string, key: string, body: string) =>
    apiPut<{ status: string }>(
      `/v1/admin/components/${type}/${name}/versions/${encodeURIComponent(version)}/artifacts/${encodeURIComponent(key)}`,
      { body },
    ),
  registerComponentPackage: (
    type: string,
    name: string,
    version: string,
    body?: { manifest?: Record<string, unknown>; actor?: string },
  ) =>
    apiPost<RegisterComponentPackageResult>(
      `/v1/admin/components/${type}/${name}/versions/${encodeURIComponent(version)}/register-package`,
      body ?? {},
    ),
  validateComponentPackage: async (type: string, name: string, version: string) => {
    const path = `/v1/admin/components/${type}/${name}/versions/${encodeURIComponent(version)}/validate-package`;
    const res = await fetch(`${base}${path}`, {
      method: "POST",
      headers: { ...headers(), "Content-Type": "application/json" },
      body: "{}",
    });
    if (res.status === 404 || res.status === 409) {
      throw new Error(await parseError(res, path));
    }
    const data = (await res.json()) as ValidateComponentPackageResult;
    if (res.status !== 200 && res.status !== 422) {
      throw new Error(await parseError(res, path));
    }
    return data;
  },
  promoteComponentVersion: (type: string, name: string, version: string, actor?: string) =>
    apiPost<{ type: string; name: string; version: string; status: string }>(
      `/v1/admin/components/${type}/${name}/versions/${encodeURIComponent(version)}/promote`,
      { actor },
    ),
  auditLog: (params?: {
    entityType?: string;
    action?: string;
    limit?: number;
    offset?: number;
  }) => {
    const q = new URLSearchParams();
    if (params?.entityType) q.set("entityType", params.entityType);
    if (params?.action) q.set("action", params.action);
    if (params?.limit != null) q.set("limit", String(params.limit));
    if (params?.offset != null) q.set("offset", String(params.offset));
    const qs = q.toString();
    return apiList<AuditLogEntry>(`/v1/admin/audit-log${qs ? `?${qs}` : ""}`);
  },
  blastRadius: (name: string, version: string) =>
    apiGet<BlastRadius>(`/v1/admin/golden-paths/${name}/versions/${version}/blast-radius`),
};
