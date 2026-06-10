import type {
  ArtifactDetail,
  ArtifactMeta,
  AuditLogEntry,
  BlastRadius,
  CanaryOverview,
  CatalogOverview,
  Component,
  ComponentVersion,
  DashboardStats,
  DraftGPResult,
  GPProfile,
  GPRelease,
  GPReleaseDetail,
  HealthSummary,
  ListResponse,
  MeResponse,
  Project,
  PublishGPResult,
  ResolvePreviewResult,
} from "../api/types";

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
  ready: () => apiGet<{ status: string }>("/ready"),
  me: () => apiGet<MeResponse>("/v1/admin/me"),
  stats: () => apiGet<DashboardStats>("/v1/admin/stats"),
  projects: (goldenPath?: string, version?: string) => {
    const q = new URLSearchParams();
    if (goldenPath) q.set("goldenPath", goldenPath);
    if (version) q.set("version", version);
    const qs = q.toString();
    return apiGet<ListResponse<Project>>(`/v1/admin/projects${qs ? `?${qs}` : ""}`);
  },
  gpNames: () => apiGet<ListResponse<string>>("/v1/admin/golden-paths/names"),
  gpProfile: (name: string) => apiGet<GPProfile>(`/v1/admin/golden-paths/${name}/profile`),
  gpReleases: (name?: string, includeDrafts = false) => {
    const q = new URLSearchParams();
    if (name) q.set("name", name);
    if (includeDrafts) q.set("includeDrafts", "true");
    const qs = q.toString();
    return apiGet<ListResponse<GPRelease>>(`/v1/admin/golden-paths${qs ? `?${qs}` : ""}`);
  },
  gpRelease: (name: string, version: string) =>
    apiGet<GPReleaseDetail>(`/v1/admin/golden-paths/${name}/versions/${version}`),
  createDraftGPRelease: (
    name: string,
    body: { version: string; composition: Record<string, string>; actor?: string },
  ) => apiPost<DraftGPResult>(`/v1/admin/golden-paths/${name}/drafts`, body),
  publishGPRelease: (
    name: string,
    body: { version: string; composition: Record<string, string>; actor?: string },
  ) => apiPost<PublishGPResult>(`/v1/admin/golden-paths/${name}/versions`, body),
  promoteDraftGPRelease: (name: string, version: string, actor?: string) => {
    const q = actor ? `?actor=${encodeURIComponent(actor)}` : "";
    return apiPost<PublishGPResult>(
      `/v1/admin/golden-paths/${name}/versions/${encodeURIComponent(version)}/promote${q}`,
      {},
    );
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
    apiGet<ListResponse<ArtifactMeta>>(
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
  resolvePreview: (name: string, pin: string, project?: string) => {
    const q = new URLSearchParams({ pin });
    if (project) q.set("project", project);
    return apiGet<ResolvePreviewResult>(
      `/v1/admin/golden-paths/${name}/resolve-preview?${q.toString()}`,
    );
  },
  components: () => apiGet<ListResponse<Component>>("/v1/admin/components"),
  componentVersions: (type: string, name: string) =>
    apiGet<ListResponse<ComponentVersion>>(`/v1/admin/components/${type}/${name}/versions`),
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
    return apiGet<ListResponse<AuditLogEntry>>(`/v1/admin/audit-log${qs ? `?${qs}` : ""}`);
  },
  blastRadius: (name: string, version: string) =>
    apiGet<BlastRadius>(`/v1/admin/golden-paths/${name}/versions/${version}/blast-radius`),
};
