export type PlatformEditorKind = "branching-model" | "gp-content" | "generic";

export type PlatformComponentType = {
  type: string;
  label: string;
  description: string;
  primaryArtifact: string;
  editor: PlatformEditorKind;
};

export const PLATFORM_COMPONENT_TYPES: PlatformComponentType[] = [
  {
    type: "branching-model",
    label: "Branching model",
    description: "Правила ветвления и публикации для GP (model.yaml)",
    primaryArtifact: "model.yaml",
    editor: "branching-model",
  },
];

export function platformTypeConfig(type: string): PlatformComponentType | undefined {
  return PLATFORM_COMPONENT_TYPES.find((t) => t.type === type);
}

export function isPlatformEditableType(type: string): boolean {
  return PLATFORM_COMPONENT_TYPES.some((t) => t.type === type);
}
