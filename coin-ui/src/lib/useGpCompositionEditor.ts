import { useEffect, useState } from "react";
import type { CompositionItem } from "../api/types";
import { api } from "./api";
import { defaultAgentStackName, defaultBranchingModelForGP } from "./gpSlots";

function versionLabels(items: { version: string; status: string }[], publishedOnly: boolean): string[] {
  if (publishedOnly) {
    return items.filter((v) => v.status === "published").map((v) => v.version);
  }
  return items
    .filter((v) => v.status === "published" || v.status === "draft")
    .map((v) => v.version);
}

function publishedVersions(items: { version: string; status: string }[]): string[] {
  return versionLabels(items, true);
}

export function useGpCompositionEditor(gpName: string, initial?: CompositionItem[]) {
  const [agentStackName, setAgentStackName] = useState(defaultAgentStackName());
  const [agentStackOptions, setAgentStackOptions] = useState<string[]>([]);
  const [gpContentName, setGpContentName] = useState("");
  const [gpContentOptions, setGpContentOptions] = useState<string[]>([]);
  const [branchingModelName, setBranchingModelName] = useState("");
  const [branchingModelOptions, setBranchingModelOptions] = useState<string[]>([]);
  const [composition, setComposition] = useState<Record<string, string>>({});
  const [versionOptions, setVersionOptions] = useState<Record<string, string[]>>({});
  const [versionStatuses, setVersionStatuses] = useState<Record<string, Record<string, string>>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!gpName) return;
    let cancelled = false;
    setLoading(true);
    setError(null);

    const fromRows = (rows?: CompositionItem[]) => {
      const comp: Record<string, string> = {};
      let agent = defaultAgentStackName();
      let gc = "";
      let bm = defaultBranchingModelForGP(gpName);
      for (const row of rows ?? []) {
        if (row.type === "agent") {
          comp.agent = row.version;
          agent = row.name;
        }
        if (row.type === "gp-content") {
          comp["gp-content"] = row.version;
          gc = row.name;
        }
        if (row.type === "branching-model") {
          comp["branching-model"] = row.version;
          bm = row.name;
        }
      }
      return { comp, agent, gc, bm };
    };

    (async () => {
      try {
        const components = await api.components();
        if (cancelled) return;

        const agentNames = components.items
          .filter((c) => c.type === "agent")
          .map((c) => c.name)
          .sort();
        const gcNames = components.items
          .filter((c) => c.type === "gp-content")
          .map((c) => c.name)
          .sort();
        const bmNames = components.items
          .filter((c) => c.type === "branching-model")
          .map((c) => c.name)
          .sort();

        setAgentStackOptions(agentNames);
        setGpContentOptions(gcNames);
        setBranchingModelOptions(bmNames);

        const parsed = fromRows(initial);
        let agent = parsed.agent;
        let gc = parsed.gc;
        let bm = parsed.bm;
        if (!gc && gcNames.length > 0) gc = gcNames[0];
        if (!bmNames.includes(bm) && bmNames.length > 0) {
          bm = bmNames.includes(defaultBranchingModelForGP(gpName))
            ? defaultBranchingModelForGP(gpName)
            : bmNames[0];
        }
        if (!agentNames.includes(agent) && agentNames.length > 0) {
          agent = agentNames.includes(defaultAgentStackName())
            ? defaultAgentStackName()
            : agentNames[0];
        }

        setAgentStackName(agent);
        setGpContentName(gc);
        setBranchingModelName(bm);

        const versions: Record<string, string[]> = {};
        const statuses: Record<string, Record<string, string>> = {};
        if (agent) {
          const r = await api.componentVersions("agent", agent);
          versions.agent = publishedVersions(r.items);
          statuses.agent = Object.fromEntries(r.items.map((v) => [v.version, v.status]));
        }
        if (gc) {
          const r = await api.componentVersionsOptional("gp-content", gc);
          versions["gp-content"] = versionLabels(r?.items ?? [], false);
          statuses["gp-content"] = Object.fromEntries((r?.items ?? []).map((v) => [v.version, v.status]));
        }
        if (bm) {
          const r = await api.componentVersions("branching-model", bm);
          versions["branching-model"] = versionLabels(r.items, false);
          statuses["branching-model"] = Object.fromEntries(r.items.map((v) => [v.version, v.status]));
        }
        if (cancelled) return;

        const comp = { ...parsed.comp };
        for (const key of ["agent", "gp-content", "branching-model"] as const) {
          if (!comp[key] && versions[key]?.length) {
            comp[key] = versions[key][0];
          }
        }

        setVersionOptions(versions);
        setVersionStatuses(statuses);
        setComposition(comp);
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
  }, [gpName, initial ? JSON.stringify(initial) : ""]);

  useEffect(() => {
    if (!agentStackName) return;
    api
      .componentVersions("agent", agentStackName)
      .then((r) => {
        const vers = publishedVersions(r.items);
        setVersionOptions((prev) => ({ ...prev, agent: vers }));
        setComposition((prev) => {
          if (prev.agent && vers.includes(prev.agent)) return prev;
          return { ...prev, agent: vers[0] ?? "" };
        });
      })
      .catch(() => {});
  }, [agentStackName]);

  useEffect(() => {
    if (!gpContentName) return;
    api
      .componentVersionsOptional("gp-content", gpContentName)
      .then((r) => {
        const vers = versionLabels(r?.items ?? [], false);
        setVersionOptions((prev) => ({ ...prev, "gp-content": vers }));
        setVersionStatuses((prev) => ({
          ...prev,
          "gp-content": Object.fromEntries((r?.items ?? []).map((v) => [v.version, v.status])),
        }));
        setComposition((prev) => {
          if (prev["gp-content"] && vers.includes(prev["gp-content"])) return prev;
          return { ...prev, "gp-content": vers[0] ?? "" };
        });
      })
      .catch(() => {});
  }, [gpContentName]);

  useEffect(() => {
    if (!branchingModelName) return;
    api
      .componentVersions("branching-model", branchingModelName)
      .then((r) => {
        const vers = versionLabels(r.items, false);
        setVersionOptions((prev) => ({ ...prev, "branching-model": vers }));
        setVersionStatuses((prev) => ({
          ...prev,
          "branching-model": Object.fromEntries(r.items.map((v) => [v.version, v.status])),
        }));
        setComposition((prev) => {
          if (prev["branching-model"] && vers.includes(prev["branching-model"])) return prev;
          return { ...prev, "branching-model": vers[0] ?? "" };
        });
      })
      .catch(() => {});
  }, [branchingModelName]);

  return {
    agentStackName,
    setAgentStackName,
    agentStackOptions,
    gpContentName,
    setGpContentName,
    gpContentOptions,
    branchingModelName,
    setBranchingModelName,
    branchingModelOptions,
    composition,
    setComposition,
    versionOptions,
    versionStatuses,
    loading,
    error,
  };
}
