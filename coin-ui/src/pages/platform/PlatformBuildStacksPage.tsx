import PlatformCatalogPage from "./PlatformCatalogPage";

export default function PlatformBuildStacksPage() {
  return (
    <PlatformCatalogPage
      config={{
        title: "Build stacks",
        description: "gp-content — Dockerfile, scripts и schema для каждого GP profile",
        types: ["gp-content"],
        emptyLabel: "Нет build stacks",
        showType: false,
        studioType: "gp-content",
      }}
    />
  );
}
