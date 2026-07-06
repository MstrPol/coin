// Registry auth для BuildKit/podman (читают /root/.docker/config.json).

/**
 * Материализует Docker config с credentials из withCredentials.
 * Registry host берётся из .coin/manifest.json destinations.imageRegistryPrefix.
 */
def call() {
    sh '''
        set -eu
        REG_HOST="$(python3 - <<'PY'
import json
with open(".coin/manifest.json", "r", encoding="utf-8") as f:
    prefix = json.load(f)["destinations"]["imageRegistryPrefix"]
print(prefix.split("/", 1)[0])
PY
)"
        REG_HOST="${REG_HOST%%/*}"
        mkdir -p /root/.docker
        AUTH="$(printf '%s' "${COIN_REGISTRY_USER}:${COIN_REGISTRY_PASSWORD}" | base64 | tr -d '\\n')"
        printf '{"auths":{"%s":{"auth":"%s"}}}\\n' "${REG_HOST}" "${AUTH}" > /root/.docker/config.json
        chmod 600 /root/.docker/config.json
    '''
}
