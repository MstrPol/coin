export type ComponentVersionItem = { version: string; status: string };

export function versionLabels(items: ComponentVersionItem[], publishedOnly: boolean): string[] {
  if (publishedOnly) {
    return items.filter((v) => v.status === "published").map((v) => v.version);
  }
  return items
    .filter((v) => v.status === "published" || v.status === "draft")
    .map((v) => v.version);
}

export function publishedVersions(items: ComponentVersionItem[]): string[] {
  return versionLabels(items, true);
}

export function versionStatusMap(items: ComponentVersionItem[]): Record<string, string> {
  return Object.fromEntries(items.map((v) => [v.version, v.status]));
}

export function gpSlotEmptyVersionHint(slot: string): string {
  switch (slot) {
    case "agent":
      return "Нет published версий — опубликуйте agent на Platform → Runtime";
    case "branching-model":
      return "Нет версий — создайте draft на Platform → Branching models";
    default:
      return "Нет версий";
  }
}
