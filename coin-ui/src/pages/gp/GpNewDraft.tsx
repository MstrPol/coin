import { useParams } from "react-router-dom";
import PublishWizard from "../PublishWizard";

export default function GpNewDraft() {
  const { name = "" } = useParams();
  return <PublishWizard scopedGpName={name} />;
}
