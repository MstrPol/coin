export type StudioEditorKind = "branching-model" | "gp-content" | "generic";

export type StudioComponentType = {
  type: string;
  label: string;
  description: string;
  primaryArtifact: string;
  editor: StudioEditorKind;
};

export const STUDIO_COMPONENT_TYPES: StudioComponentType[] = [
  {
    type: "branching-model",
    label: "Branching model",
    description: "Правила ветвления и публикации для GP (model.yaml)",
    primaryArtifact: "model.yaml",
    editor: "branching-model",
  },
  {
    type: "gp-content",
    label: "GP content",
    description: "Build engine policy, pipeline stages, Containerfile (content.yaml)",
    primaryArtifact: "content.yaml",
    editor: "gp-content",
  },
];

export function studioTypeConfig(type: string): StudioComponentType | undefined {
  return STUDIO_COMPONENT_TYPES.find((t) => t.type === type);
}

export function isStudioType(type: string): boolean {
  return STUDIO_COMPONENT_TYPES.some((t) => t.type === type);
}
