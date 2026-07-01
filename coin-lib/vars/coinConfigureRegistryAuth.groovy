// Registry auth для BuildKit (buildkitd читает ~/.docker/config.json).

/**
 * Материализует Docker config с Nexus credentials из withCredentials.
 * Требует COIN_REGISTRY_USER, COIN_REGISTRY_PASSWORD, COIN_REGISTRY_PREFIX в env.
 */
def call() {
    sh '''
        set -eu
        REG_HOST="${COIN_REGISTRY_PREFIX:-nexus:8082/coin-docker}"
        REG_HOST="${REG_HOST%%/*}"
        mkdir -p ~/.docker
        AUTH="$(printf '%s' "${COIN_REGISTRY_USER}:${COIN_REGISTRY_PASSWORD}" | base64 | tr -d '\\n')"
        printf '{"auths":{"%s":{"auth":"%s"}}}\\n' "${REG_HOST}" "${AUTH}" > ~/.docker/config.json
        chmod 600 ~/.docker/config.json
    '''
}
