export type AgentImageParseResult =
  | { ok: true; repository: string; tag: string }
  | { ok: false; message: string };

/** Parse agent image ref — mirrors coin-api ParseAgentImageRef rules. */
export function parseAgentImageRef(image: string): AgentImageParseResult {
  const trimmed = image.trim();
  if (!trimmed) {
    return { ok: false, message: "required" };
  }
  let ref = trimmed;
  const at = ref.indexOf("@sha256:");
  if (at >= 0) {
    ref = ref.slice(0, at);
  }
  const slash = ref.lastIndexOf("/");
  const repoTag = slash >= 0 ? ref.slice(slash + 1) : ref;
  const colon = repoTag.lastIndexOf(":");
  if (colon < 0 || colon === repoTag.length - 1) {
    return { ok: false, message: "must include image tag after repository name" };
  }
  const repository = repoTag.slice(0, colon);
  const tag = repoTag.slice(colon + 1);
  if (!repository || !tag) {
    return { ok: false, message: "invalid image reference" };
  }
  if (tag === "latest") {
    return { ok: false, message: "tag latest is not allowed" };
  }
  if (tag.startsWith("sha256:")) {
    return { ok: false, message: "digest-only reference requires a version tag" };
  }
  return { ok: true, repository, tag };
}

export function parseAgentImageRefForProfile(
  image: string,
  profileName: string,
): AgentImageParseResult & { tag?: string } {
  const parsed = parseAgentImageRef(image);
  if (!parsed.ok) {
    return parsed;
  }
  if (parsed.repository !== profileName) {
    return {
      ok: false,
      message: `repository "${parsed.repository}" must match profile "${profileName}"`,
    };
  }
  return { ok: true, repository: parsed.repository, tag: parsed.tag };
}
