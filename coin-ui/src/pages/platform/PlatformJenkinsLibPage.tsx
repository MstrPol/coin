import PlatformCatalogPage from "./PlatformCatalogPage";

export default function PlatformJenkinsLibPage() {
  return (
    <PlatformCatalogPage
      config={{
        title: "Jenkins library",
        description: "coin-lib — Jenkins Shared Library (glue only)",
        types: ["lib"],
        emptyLabel: "Нет lib-компонентов",
        showType: false,
        studioType: "lib",
      }}
    />
  );
}
