#!/usr/bin/env bash
set -euo pipefail

coin_record_output() {
  local name="$1" type="$2" ref="${3:-}" sha="${4:-}" format="${5:-}"
  python3 - "$name" "$type" "$ref" "$sha" "$format" <<'PY'
import json, os, sys
name, typ, ref, sha, fmt = sys.argv[1:6]
path = ".coin/outputs.json"
items = []
if os.path.exists(path):
    with open(path) as f:
        items = json.load(f)
items = [o for o in items if o.get("name") != name]
entry = {"name": name, "type": typ}
if ref:
    entry["ref"] = ref
if sha:
    entry["sha256"] = sha
if fmt:
    entry["format"] = fmt
items.append(entry)
with open(path, "w") as f:
    json.dump(items, f)
PY
}

echo "==> coin standard publish (go/container)"
if [[ -f .coin/build.env ]]; then
  # shellcheck source=/dev/null
  source .coin/build.env
fi

project="$(grep -E '^  name:' .coin/config.yaml 2>/dev/null | awk '{print $2}' || true)"
project="${project:-app}"
product_version="${COIN_IMAGE_TAG:-$(python3 -c 'import json; print(json.load(open(".coin/manifest.json"))["goldenPath"]["version"])' 2>/dev/null || echo local)}"

if [[ -n "${COIN_BUILT_IMAGE:-}" ]]; then
  if command -v /kaniko/executor >/dev/null 2>&1; then
    echo "Image pushed during kaniko build: ${COIN_BUILT_IMAGE}"
  elif command -v docker >/dev/null 2>&1; then
    docker push "${COIN_BUILT_IMAGE}"
  fi
  coin_record_output "app" "image" "${COIN_BUILT_IMAGE}"
else
  echo "No container image to publish"
fi

python3 - "$project" "$product_version" <<'PY'
import json, os, pathlib, subprocess, sys, zipfile, hashlib

project, version = sys.argv[1:3]
deliverables_path = pathlib.Path(".coin/deliverables.json")
if not deliverables_path.exists():
    raise SystemExit(0)
deliverables = json.loads(deliverables_path.read_text())
has_artifact = any((spec or {}).get("type") == "artifact" for spec in deliverables.values())
if not has_artifact:
    raise SystemExit(0)

import yaml
cfg = yaml.safe_load(open(".coin/config.yaml"))
merged = cfg.get("deliverables") or deliverables

def record_output(name, typ, ref="", sha="", fmt=""):
    path = pathlib.Path(".coin/outputs.json")
    items = json.loads(path.read_text()) if path.exists() else []
    items = [o for o in items if o.get("name") != name]
    entry = {"name": name, "type": typ}
    if ref:
        entry["ref"] = ref
    if sha:
        entry["sha256"] = sha
    if fmt:
        entry["format"] = fmt
    items.append(entry)
    path.write_text(json.dumps(items))

for name, spec in merged.items():
    if spec.get("type") != "artifact":
        continue
    sources = spec.get("sources") or []
    if not sources and spec.get("source"):
        sources = [{"path": spec["source"]}]
    if not sources:
        print(f"No sources for artifact {name}", file=sys.stderr)
        raise SystemExit(1)
    out_dir = pathlib.Path(".coin/out/artifacts")
    out_dir.mkdir(parents=True, exist_ok=True)
    zip_path = out_dir / f"{project}-{name}.zip"
    with zipfile.ZipFile(zip_path, "w", zipfile.ZIP_DEFLATED) as zf:
        for item in sources:
            src = pathlib.Path(item["path"])
            arc = item.get("as") or src.name
            if src.is_dir():
                for p in sorted(src.rglob("*")):
                    if p.is_file():
                        rel = p.relative_to(src)
                        zf.write(p, f"{arc}/{rel}" if arc else str(rel))
            elif src.is_file():
                zf.write(src, arc)
            else:
                print(f"missing source path: {src}", file=sys.stderr)
                raise SystemExit(1)
    digest = hashlib.sha256(zip_path.read_bytes()).hexdigest()
    nexus = os.environ.get("NEXUS_URL", "http://nexus:8081")
    repo = os.environ.get("NEXUS_MAVEN_RELEASES", "maven-releases")
    url = f"{nexus}/repository/{repo}/coin/deliverables/{project}/{version}/{project}-{name}.zip"
    auth = f"{os.environ.get('NEXUS_USER', '')}:{os.environ.get('NEXUS_PASSWORD', '')}"
    if not os.environ.get("NEXUS_USER"):
        print(f"Skip artifact {name}: NEXUS credentials not set")
        continue
    print(f"==> artifact zip {zip_path} -> {url}")
    subprocess.run(["curl", "-fsS", "-u", auth, "--upload-file", str(zip_path), url], check=True)
    record_output(name, "artifact", url, f"sha256:{digest}", "zip")
PY
