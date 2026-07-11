import { Navigate, useParams, useSearchParams } from "react-router-dom";

export function LegacyCatalogRedirect() {
  const [searchParams] = useSearchParams();
  const name = searchParams.get("name");
  if (name) {
    return <Navigate to={`/gp/${encodeURIComponent(name)}/policy`} replace />;
  }
  return <Navigate to="/gp" replace />;
}

export function LegacyCanaryRedirect() {
  const [searchParams] = useSearchParams();
  const name = searchParams.get("name");
  if (name) {
    return <Navigate to={`/gp/${encodeURIComponent(name)}/canary`} replace />;
  }
  return <Navigate to="/gp" replace />;
}

export function LegacyPublishRedirect() {
  const [searchParams] = useSearchParams();
  const name = searchParams.get("name");
  if (name) {
    return <Navigate to={`/gp/${encodeURIComponent(name)}/releases/new-draft`} replace />;
  }
  return <Navigate to="/gp" replace />;
}

export function LegacyReleasesRedirect() {
  const [searchParams] = useSearchParams();
  const name = searchParams.get("name");
  if (name) {
    return <Navigate to={`/gp/${encodeURIComponent(name)}/releases`} replace />;
  }
  return <Navigate to="/gp" replace />;
}

export function LegacyReleaseDetailRedirect() {
  const { name, version } = useParams();
  if (name && version) {
    return (
      <Navigate
        to={`/gp/${encodeURIComponent(name)}/releases/${encodeURIComponent(version)}`}
        replace
      />
    );
  }
  return <Navigate to="/gp" replace />;
}
