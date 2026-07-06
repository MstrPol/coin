export const GP_DRAFT_SLOT_ORDER = ["agent", "branching-model"] as const;

export const SLOT_LABELS: Record<string, string> = {
  agent: "Agent / executor stack (CI runtime)",
  "branching-model": "Branching model (versioning + publish policy)",
};

export type GpDraftSlotSpec = {
  key: (typeof GP_DRAFT_SLOT_ORDER)[number];
  type: string;
  fixedName?: string;
  pickComponent?: boolean;
};

export const GP_DRAFT_SLOT_SPECS: GpDraftSlotSpec[] = [
  { key: "agent", type: "agent" },
  { key: "branching-model", type: "branching-model", pickComponent: true },
];

export function defaultBranchingModelForGP(gpName: string): string {
  if (gpName === "go-lib" || gpName === "java-maven-app") return "semver-tag";
  return "trunk-based";
}

export function defaultAgentStackName(): string {
  return "coin-agent";
}
