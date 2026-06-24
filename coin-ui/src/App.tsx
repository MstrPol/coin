import { Link, Navigate, Route, Routes } from "react-router-dom";
import Layout from "./components/Layout";
import RequireAuth from "./components/RequireAuth";
import RequireRole from "./components/RequireRole";
import ComponentDetail from "./pages/ComponentDetail";
import ComponentStudio from "./pages/ComponentStudio";
import PlatformSettings from "./pages/PlatformSettings";
import AuditLog from "./pages/AuditLog";
import BranchingModelsPage from "./pages/BranchingModelsPage";
import BuildReports from "./pages/BuildReports";
import Canary from "./pages/Canary";
import Catalog from "./pages/Catalog";
import Dashboard from "./pages/Dashboard";
import GpReleaseDetail from "./pages/GpReleaseDetail";
import GpReleases from "./pages/GpReleases";
import Login from "./pages/Login";
import LoginCallback from "./pages/LoginCallback";
import Projects from "./pages/Projects";
import CreateGPProfile from "./pages/CreateGPProfile";
import PublishWizard from "./pages/PublishWizard";
import PromoteCanaryPage from "./pages/PromoteCanaryPage";
import ResolvePreview from "./pages/ResolvePreview";
import PlatformRuntimePage from "./pages/platform/PlatformRuntimePage";
import PlatformBuildStacksPage from "./pages/platform/PlatformBuildStacksPage";
import PlatformJenkinsLibPage from "./pages/platform/PlatformJenkinsLibPage";
import PlatformComponentsPage from "./pages/platform/PlatformComponentsPage";

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
          <Route path="releases" element={<GpReleases />} />
          <Route path="releases/:name/:version" element={<GpReleaseDetail />} />
          <Route path="catalog" element={<Catalog />} />
          <Route path="resolve" element={<ResolvePreview />} />
          <Route path="canary" element={<Canary />} />
          <Route path="branching-models" element={<Navigate to="/platform/branching-models" replace />} />
          <Route path="components" element={<Navigate to="/platform/components" replace />} />
          <Route path="components/:type/:name" element={<ComponentDetail />} />
          <Route path="components/:type/:name/:version" element={<ComponentDetail />} />
          <Route path="platform-settings" element={<PlatformSettings />} />
          <Route path="audit" element={<AuditLog />} />
          <Route path="platform/runtime" element={<PlatformRuntimePage />} />
          <Route path="platform/build-stacks" element={<PlatformBuildStacksPage />} />
          <Route path="platform/branching-models" element={<BranchingModelsPage />} />
          <Route path="platform/jenkins-lib" element={<PlatformJenkinsLibPage />} />
          <Route path="platform/components" element={<PlatformComponentsPage />} />
          <Route element={<RequireRole min="publisher" />}>
            <Route path="studio" element={<ComponentStudio />} />
            <Route path="studio/:type/:name/:version" element={<ComponentStudio />} />
            <Route path="releases/new-gp" element={<CreateGPProfile />} />
            <Route path="releases/publish" element={<PublishWizard />} />
            <Route path="promote" element={<PromoteCanaryPage />} />
          </Route>
        </Route>
      </Route>
      <Route path="*" element={<NotFound />} />
    </Routes>
  );
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
