import { Link, Navigate, Route, Routes, useParams } from "react-router-dom";
import Layout from "./components/Layout";
import RequireAuth from "./components/RequireAuth";
import RequireRole from "./components/RequireRole";
import ComponentDetail from "./pages/ComponentDetail";
import AuditLog from "./pages/AuditLog";
import BranchingModelsPage from "./pages/BranchingModelsPage";
import BuildReports from "./pages/BuildReports";
import Dashboard from "./pages/Dashboard";
import GpReleaseDetail from "./pages/GpReleaseDetail";
import Login from "./pages/Login";
import LoginCallback from "./pages/LoginCallback";
import Projects from "./pages/Projects";
import CreateGPProfile from "./pages/CreateGPProfile";
import PromoteCanaryPage from "./pages/PromoteCanaryPage";
import ResolvePreview from "./pages/ResolvePreview";
import PlatformRuntimePage from "./pages/platform/PlatformRuntimePage";
import PlatformBuildStacksPage from "./pages/platform/PlatformBuildStacksPage";
import PlatformComponentEditorPage from "./pages/platform/PlatformComponentEditorPage";
import PlatformComponentsPage from "./pages/platform/PlatformComponentsPage";
import PlatformComponentHubLayout from "./pages/platform/PlatformComponentHubLayout";
import PlatformOverviewTab from "./pages/platform/tabs/PlatformOverviewTab";
import PlatformReleasesTab from "./pages/platform/tabs/PlatformReleasesTab";
import PlatformNewDraftPage from "./pages/platform/PlatformNewDraftPage";
import PlatformNewProfilePage from "./pages/platform/PlatformNewProfilePage";
import PlatformComponentReleaseDetail from "./pages/platform/PlatformComponentReleaseDetail";
import PlatformAgentMetadataEditorPage from "./pages/platform/PlatformAgentMetadataEditorPage";
import PlatformFlatReleaseRedirect from "./pages/platform/PlatformFlatReleaseRedirect";
import GpCatalogPage from "./pages/gp/GpCatalogPage";
import GpHubLayout from "./pages/gp/GpHubLayout";
import GpNewDraft from "./pages/gp/GpNewDraft";
import GpOverviewTab from "./pages/gp/tabs/GpOverviewTab";
import GpReleasesTab from "./pages/gp/tabs/GpReleasesTab";
import GpPolicyTab from "./pages/gp/tabs/GpPolicyTab";
import GpCanaryTab from "./pages/gp/tabs/GpCanaryTab";
import {
  LegacyCanaryRedirect,
  LegacyCatalogRedirect,
  LegacyPublishRedirect,
  LegacyReleaseDetailRedirect,
  LegacyReleasesRedirect,
} from "./pages/gp/LegacyRedirects";

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/login/callback" element={<LoginCallback />} />
      <Route element={<RequireAuth />}>
        <Route element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="projects" element={<Projects />} />
          <Route path="build-reports" element={<BuildReports />} />
          <Route path="gp" element={<GpCatalogPage />} />
          <Route path="gp/:name" element={<GpHubLayout />}>
            <Route index element={<GpOverviewTab />} />
            <Route path="releases" element={<GpReleasesTab />} />
            <Route element={<RequireRole min="publisher" />}>
              <Route path="releases/new-draft" element={<GpNewDraft />} />
            </Route>
            <Route path="releases/:version" element={<GpReleaseDetail />} />
            <Route path="policy" element={<GpPolicyTab />} />
            <Route path="canary" element={<GpCanaryTab />} />
          </Route>
          <Route path="releases" element={<LegacyReleasesRedirect />} />
          <Route path="releases/:name/:version" element={<LegacyReleaseDetailRedirect />} />
          <Route path="catalog" element={<LegacyCatalogRedirect />} />
          <Route path="canary" element={<LegacyCanaryRedirect />} />
          <Route path="resolve" element={<ResolvePreview />} />
          <Route path="branching-models" element={<Navigate to="/platform/branching-models" replace />} />
          <Route path="components" element={<Navigate to="/platform/components" replace />} />
          <Route path="components/:type/:name" element={<LegacyComponentDetailRedirect />} />
          <Route path="components/:type/:name/:version" element={<ComponentDetail />} />
          <Route path="platform-settings" element={<Navigate to="/audit" replace />} />
          <Route path="audit" element={<AuditLog />} />
          <Route path="platform/runtime" element={<PlatformRuntimePage />} />
          <Route path="platform/build-stacks" element={<PlatformBuildStacksPage />} />
          <Route path="platform/branching-models" element={<BranchingModelsPage />} />
          <Route element={<RequireRole min="publisher" />}>
            <Route path="platform/runtime/new" element={<PlatformNewProfilePage familyId="runtime" />} />
            <Route
              path="platform/build-stacks/new"
              element={<PlatformNewProfilePage familyId="build-stacks" />}
            />
            <Route
              path="platform/branching-models/new"
              element={<PlatformNewProfilePage familyId="branching-models" />}
            />
          </Route>
          <Route path="platform/runtime/:name" element={<PlatformComponentHubLayout familyId="runtime" />}>
            <Route index element={<PlatformOverviewTab />} />
            <Route path="releases" element={<PlatformReleasesTab />} />
            <Route element={<RequireRole min="publisher" />}>
              <Route path="releases/new-draft" element={<PlatformNewDraftPage />} />
            </Route>
            <Route path="releases/:version" element={<PlatformComponentReleaseDetail />} />
          </Route>
          <Route
            path="platform/build-stacks/:name"
            element={<PlatformComponentHubLayout familyId="build-stacks" />}
          >
            <Route index element={<PlatformOverviewTab />} />
            <Route path="releases" element={<PlatformReleasesTab />} />
            <Route element={<RequireRole min="publisher" />}>
              <Route path="releases/new-draft" element={<PlatformNewDraftPage />} />
            </Route>
            <Route path="releases/:version" element={<PlatformComponentReleaseDetail />} />
          </Route>
          <Route
            path="platform/branching-models/:name"
            element={<PlatformComponentHubLayout familyId="branching-models" />}
          >
            <Route index element={<PlatformOverviewTab />} />
            <Route path="releases" element={<PlatformReleasesTab />} />
            <Route element={<RequireRole min="publisher" />}>
              <Route path="releases/new-draft" element={<PlatformNewDraftPage />} />
            </Route>
            <Route path="releases/:version" element={<PlatformComponentReleaseDetail />} />
          </Route>
          <Route
            path="platform/runtime/:name/:version"
            element={<PlatformFlatReleaseRedirect familyId="runtime" />}
          />
          <Route
            path="platform/build-stacks/:name/:version"
            element={<PlatformFlatReleaseRedirect familyId="build-stacks" />}
          />
          <Route
            path="platform/branching-models/:name/:version"
            element={<PlatformFlatReleaseRedirect familyId="branching-models" />}
          />
          <Route path="platform/jenkins-lib" element={<Navigate to="/platform/runtime" replace />} />
          <Route path="platform/components" element={<PlatformComponentsPage />} />
          <Route path="studio" element={<Navigate to="/platform/build-stacks" replace />} />
          <Route path="studio/*" element={<Navigate to="/platform/build-stacks" replace />} />
          <Route element={<RequireRole min="publisher" />}>
            <Route
              path="platform/runtime/:name/:version/edit"
              element={<PlatformAgentMetadataEditorPage familyId="runtime" />}
            />
            <Route
              path="platform/build-stacks/:name/:version/edit"
              element={<PlatformComponentEditorPage compType="gp-content" />}
            />
            <Route
              path="platform/branching-models/:name/:version/edit"
              element={<PlatformComponentEditorPage compType="branching-model" />}
            />
            <Route path="gp/new" element={<CreateGPProfile />} />
            <Route path="releases/new-gp" element={<Navigate to="/gp/new" replace />} />
            <Route path="releases/publish" element={<LegacyPublishRedirect />} />
            <Route path="promote" element={<PromoteCanaryPage />} />
          </Route>
        </Route>
      </Route>
      <Route path="*" element={<NotFound />} />
    </Routes>
  );
}

function LegacyComponentDetailRedirect() {
  const { type = "", name = "" } = useParams();
  if (type === "agent" && name) {
    return <Navigate to={`/platform/runtime/${encodeURIComponent(name)}`} replace />;
  }
  return <ComponentDetail />;
}

function NotFound() {
  return (
    <div className="flex min-h-screen items-center justify-center p-8">
      <div>
        <h1 className="text-xl font-semibold">404</h1>
        <Link to="/" className="text-sky-400 hover:underline">
          На главную
        </Link>
      </div>
    </div>
  );
}
