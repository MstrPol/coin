import { useCallback, useEffect, useState } from "react";
import type { CompositionItem } from "../api/types";
import { api } from "./api";
import {
  defaultAgentStackName,
  defaultBranchingModelForGP,
  GP_DRAFT_SLOT_ORDER,
} from "./gpSlots";

function versionLabels(
  items: { version: string; status?: string }[],
  publishedOnly: boolean,
): string[] {
  return items
    .filter((i) => !publishedOnly || i.status === "published")
    .map((i) => i.version);
}

function versionStatusMap(items: { version: string; status?: string }[]): Record<string, string> {
  const m: Record<string, string> = {};
  for (const i of items) {
    if (i.status) m[i.version] = i.status;
  }
  return m;
}

export function useGpCompositionEditor(gpName: string, composition?: CompositionItem[]) {
  const [agentStackName, setAgentStackName] = useState("");
  const [agentStackOptions, setAgentStackOptions] = useState<string[]>([]);
  const [branchingModelName, setBranchingModelName] = useState("");
  const [branchingModelOptions, setBranchingModelOptions] = useState<string[]>([]);
  const [compositionState, setCompositionState] = useState<Record<string, string>>({});
  const [versionOptions, setVersionOptions] = useState<Record<string, string[]>>({});
  const [versionStatuses, setVersionStatuses] = useState<Record<string, Record<string, string>>>(
    {},
  );

  useEffect(() => {
    if (!composition) return;
    const comp: Record<string, string> = {};
    for (const row of composition) {
      if (row.type === "agent") {
        comp.agent = row.version;
        setAgentStackName(row.name);
      }
      if (row.type === "branching-model") {
        comp["branching-model"] = row.version;
        setBranchingModelName(row.name);
      }
    }
    setCompositionState(comp);
  }, [composition]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      const catalog = await api.components();
      if (cancelled) return;
      const agentNames = catalog.items
        .filter((c) => c.type === "agent")
        .map((c) => c.name)
        .sort();
      const bmNames = catalog.items
        .filter((c) => c.type === "branching-model")
        .map((c) => c.name)
        .sort();
      setAgentStackOptions(agentNames.length ? agentNames : [defaultAgentStackName()]);
      setBranchingModelOptions(bmNames.length ? bmNames : [defaultBranchingModelForGP(gpName)]);
      setAgentStackName((prev) => {
        if (prev && agentNames.includes(prev)) return prev;
        return agentNames[0] ?? defaultAgentStackName();
      });
      setBranchingModelName((prev) => {
        if (prev && bmNames.includes(prev)) return prev;
        return bmNames[0] ?? defaultBranchingModelForGP(gpName);
      });
    })();
    return () => {
      cancelled = true;
    };
  }, [gpName]);

  const loadVersions = useCallback(async (agent: string, bm: string) => {
    const versions: Record<string, string[]> = {};
    const statuses: Record<string, Record<string, string>> = {};
    const agentR = await api.componentVersionsOptional("agent", agent);
    versions.agent = versionLabels(agentR?.items ?? [], true);
    statuses.agent = versionStatusMap(agentR?.items ?? []);
    const bmR = await api.componentVersionsOptional("branching-model", bm);
    versions["branching-model"] = versionLabels(bmR?.items ?? [], false);
    statuses["branching-model"] = versionStatusMap(bmR?.items ?? []);
    setVersionOptions(versions);
    setVersionStatuses(statuses);
    setCompositionState((prev) => {
      const next = { ...prev };
      for (const key of GP_DRAFT_SLOT_ORDER) {
        const vers = versions[key] ?? [];
        if (!prev[key] && vers.length > 0) next[key] = vers[0];
        else if (prev[key] && !vers.includes(prev[key])) next[key] = vers[0] ?? "";
      }
      return next;
    });
  }, []);

  useEffect(() => {
    if (!agentStackName || !branchingModelName) return;
    void loadVersions(agentStackName, branchingModelName);
  }, [agentStackName, branchingModelName, loadVersions]);

  return {
    agentStackName,
    setAgentStackName,
    agentStackOptions,
    branchingModelName,
    setBranchingModelName,
    branchingModelOptions,
    composition: compositionState,
    setComposition: setCompositionState,
    versionOptions,
    versionStatuses,
  };
}
