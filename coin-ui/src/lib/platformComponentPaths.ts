import {
  familyByCompType,
  familyCatalogPath,
  familyReleaseDetailPath,
  type PlatformFamilyId,
} from "./platformFamilyConfig";

export function platformCatalogPath(type: string): string | null {
  const family = familyByCompType(type);
  if (!family) return null;
  return familyCatalogPath(family.id);
}

export function platformHubPath(type: string, name: string): string | null {
  const family = familyByCompType(type);
  if (!family) return null;
  return `${familyCatalogPath(family.id)}/${encodeURIComponent(name)}`;
}

export function platformEditPath(type: string, name: string, version: string): string | null {
  const family = familyByCompType(type);
  if (!family) return null;
  return `${familyCatalogPath(family.id)}/${encodeURIComponent(name)}/${encodeURIComponent(version)}/edit`;
}

export function platformDetailPath(type: string, name: string, version: string): string | null {
  const family = familyByCompType(type);
  if (!family) return null;
  return familyReleaseDetailPath(family.id, name, version);
}

export function platformReleaseDetailPathForFamily(
  familyId: PlatformFamilyId,
  name: string,
  version: string,
): string {
  return familyReleaseDetailPath(familyId, name, version);
}

export function derivedExecutorPin(
  agentName: string,
  version: string,
): { type: string; name: string; version: string } | null {
  if (agentName === "coin-agent") {
    return { type: "executor", name: "coin-executor", version };
  }
  return null;
}
