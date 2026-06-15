import type { GPProfileSlot } from "../api/types";

export const CANONICAL_SLOT_ORDER = [
  "jnlp",
  "agent",
  "executor",
  "lib",
  "gp-content",
] as const;

export const SLOT_LABELS: Record<string, string> = {
  jnlp: "JNLP inbound agent",
  agent: "CI agent stack",
  executor: "coin-executor",
  lib: "coin-lib (Jenkins Shared Library)",
  "gp-content": "GP content (scripts, Dockerfile, schema)",
};

export type SlotPickerSpec = {
  key: (typeof CANONICAL_SLOT_ORDER)[number];
  type: string;
  fixedName?: string;
  pickComponent?: boolean;
  excludeNames?: string[];
};

export const SLOT_PICKER_SPECS: SlotPickerSpec[] = [
  { key: "jnlp", type: "agent", fixedName: "jnlp" },
  { key: "agent", type: "agent", pickComponent: true, excludeNames: ["jnlp"] },
  { key: "executor", type: "executor", fixedName: "coin-executor" },
  { key: "lib", type: "lib", fixedName: "coin-lib" },
  { key: "gp-content", type: "gp-content", pickComponent: true },
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
