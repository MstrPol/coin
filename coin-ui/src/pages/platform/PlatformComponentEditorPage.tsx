import { useParams } from "react-router-dom";
import PlatformComponentEditor from "../../components/platform/PlatformComponentEditor";
import { useAuth } from "../../context/AuthContext";

type Props = { compType: "gp-content" | "branching-model" };

export default function PlatformComponentEditorPage({ compType }: Props) {
  const { name = "", version = "" } = useParams();
  const { can } = useAuth();
  return (
    <PlatformComponentEditor
      type={compType}
      name={name}
      version={version}
      canEdit={can("publisher")}
    />
  );
}
