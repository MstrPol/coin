import { Navigate, useParams } from "react-router-dom";
import { familyReleaseDetailPath, type PlatformFamilyId } from "../../lib/platformFamilyConfig";

export default function PlatformFlatReleaseRedirect({ familyId }: { familyId: PlatformFamilyId }) {
  const { name = "", version = "" } = useParams();
  if (!name || !version || version === "releases" || version === "new") {
    return <Navigate to={`/platform/${familyId === "runtime" ? "runtime" : familyId}/${name}`} replace />;
  }
  return <Navigate to={familyReleaseDetailPath(familyId, name, version)} replace />;
}
