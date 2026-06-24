import PlatformCatalogPage from "./PlatformCatalogPage";

export default function PlatformRuntimePage() {
  return (
    <PlatformCatalogPage
      config={{
        title: "Runtime",
        description: "Agent и executor — script-first publish, версии в registry",
        types: ["agent", "executor"],
        emptyLabel: "Нет runtime-компонентов",
        hint: "Публикация через publish-скрипты и Nexus; Studio не требуется для agent/executor.",
      }}
    />
  );
}
