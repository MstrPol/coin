import { Link, Route, Routes } from "react-router-dom";
import Layout from "./components/Layout";
import RequireAuth from "./components/RequireAuth";
import RequireRole from "./components/RequireRole";
import ComponentDetail from "./pages/ComponentDetail";
import Components from "./pages/Components";
import PlatformSettings from "./pages/PlatformSettings";
import AuditLog from "./pages/AuditLog";
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
import ResolvePreview from "./pages/ResolvePreview";

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
          <Route path="components" element={<Components />} />
          <Route path="components/:type/:name" element={<ComponentDetail />} />
          <Route path="components/:type/:name/:version" element={<ComponentDetail />} />
          <Route path="platform-settings" element={<PlatformSettings />} />
          <Route path="audit" element={<AuditLog />} />
          <Route element={<RequireRole min="publisher" />}>
            <Route path="releases/new-gp" element={<CreateGPProfile />} />
            <Route path="releases/publish" element={<PublishWizard />} />
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
