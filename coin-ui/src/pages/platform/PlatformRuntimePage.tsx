import PlatformCatalogPage from "./PlatformCatalogPage";

export default function PlatformRuntimePage() {
  return (
    <PlatformCatalogPage
      config={{
        title: "Runtime components",
        description: "Agent и executor (CI runtime stacks); agent stack выбирается в GP draft",
        types: ["agent", "executor"],
        emptyLabel: "Нет runtime-компонентов",
        hint: "Agent stack выбирается в GP draft composition.",
      }}
    />
  );
}
