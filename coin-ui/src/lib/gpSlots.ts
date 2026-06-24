import type { GPProfileSlot } from "../api/types";

export const CANONICAL_SLOT_ORDER = [
  "agent",
  "executor",
  "lib",
  "gp-content",
  "branching-model",
] as const;

export const SLOT_LABELS: Record<string, string> = {
  agent: "CI agent stack (coin-agent)",
  executor: "coin-executor",
  lib: "coin-lib (Jenkins Shared Library)",
  "gp-content": "GP content (build policy, Containerfile, schema)",
  "branching-model": "Branching model (versioning + publish policy)",
};

export type SlotPickerSpec = {
  key: (typeof CANONICAL_SLOT_ORDER)[number];
  type: string;
  fixedName?: string;
  pickComponent?: boolean;
  excludeNames?: string[];
};

export const SLOT_PICKER_SPECS: SlotPickerSpec[] = [
  { key: "agent", type: "agent", fixedName: "coin-agent" },
  { key: "executor", type: "executor", fixedName: "coin-executor" },
  { key: "lib", type: "lib", fixedName: "coin-lib" },
  { key: "gp-content", type: "gp-content", pickComponent: true },
  {
    key: "branching-model",
    type: "branching-model",
    pickComponent: true,
    excludeNames: [],
  },
];

export function sortProfileSlots<T extends { key: string }>(slots: T[]): T[] {
  return [...slots].sort((a, b) => {
    const ai = CANONICAL_SLOT_ORDER.indexOf(a.key as (typeof CANONICAL_SLOT_ORDER)[number]);
    const bi = CANONICAL_SLOT_ORDER.indexOf(b.key as (typeof CANONICAL_SLOT_ORDER)[number]);
    return (ai === -1 ? 99 : ai) - (bi === -1 ? 99 : bi);
  });
}

export function isCanonicalProfile(slots: { key: string }[]): boolean {
  if (slots.length !== CANONICAL_SLOT_ORDER.length) return false;
  const keys = new Set(slots.map((s) => s.key));
  return CANONICAL_SLOT_ORDER.every((k) => keys.has(k));
}

export function slotsToProfile(slots: {
  key: string;
  type: string;
  componentName: string;
}[]): GPProfileSlot[] {
  return slots.map((s) => ({ key: s.key, type: s.type, name: s.componentName }));
}
