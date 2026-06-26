import { useEffect, useState } from "react";
import type { Component, ComponentVersion } from "../../api/types";
import { api } from "../../lib/api";
import { PLATFORM_FAMILIES } from "../../lib/platformFamilyConfig";
import PlatformProfileCatalogPage, { buildProfileRows } from "./PlatformProfileCatalogPage";

export default function PlatformRuntimePage() {
  const family = PLATFORM_FAMILIES.runtime;
  const [rows, setRows] = useState<ReturnType<typeof buildProfileRows>>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoading(true);
      setError(null);
      try {
        const all = await api.components();
        const profiles = all.items.filter((c: Component) => c.type === "agent");
        const versionsByName: Record<string, ComponentVersion[]> = {};
        const gpUsageByName: Record<string, number> = {};
        await Promise.all(
          profiles.map(async (c) => {
            const [vers, detail] = await Promise.all([
              api.componentVersions(family.compType, c.name),
              api.componentDetail(family.compType, c.name),
            ]);
            versionsByName[c.name] = vers.items;
            gpUsageByName[c.name] = detail.gpUsage?.length ?? 0;
          }),
        );
        if (!cancelled) {
          setRows(buildProfileRows(profiles, versionsByName, gpUsageByName));
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "load failed");
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [family.compType]);

  return (
    <PlatformProfileCatalogPage family={family} rows={rows} loading={loading} error={error} />
  );
}
