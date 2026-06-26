export function platformCatalogPath(type: string): string | null {
  if (type === "gp-content") return "/platform/build-stacks";
  if (type === "branching-model") return "/platform/branching-models";
  return null;
}

export function platformEditPath(type: string, name: string, version: string): string | null {
  const base = platformCatalogPath(type);
  if (!base) return null;
  return `${base}/${encodeURIComponent(name)}/${encodeURIComponent(version)}/edit`;
}

export function platformDetailPath(type: string, name: string, version: string): string | null {
  const base = platformCatalogPath(type);
  if (!base) return null;
  return `${base}/${encodeURIComponent(name)}/${encodeURIComponent(version)}`;
}
