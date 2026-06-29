// Вызов одного stage coin-executor внутри agent-контейнера.

/**
 * Запускает coin-executor run для указанного stage из manifest.pipeline.stages.
 * Подставляет credentials API token и env из coinApplyEnv.
 *
 * @param stageName идентификатор stage (validate, test, build, publish, …)
 */
def call(String stageName) {
    def apiTokenCredId = env.COIN_API_TOKEN_CRED ?: 'coin-api-token'
    withCredentials([string(credentialsId: apiTokenCredId, variable: 'COIN_API_TOKEN')]) {
        sh """
            set -eu
            export PATH="/usr/local/bin:\${PATH}"
            export COIN_REGISTRY_PREFIX="\${COIN_REGISTRY_PREFIX:-localhost:8082/coin-docker}"
            export COIN_API_URL='${env.COIN_API_URL}'
            export COIN_API_TOKEN="\${COIN_API_TOKEN}"
            export COIN_PUBLISH_REQUEST="\${COIN_PUBLISH_REQUEST:-false}"
            coin-executor run --manifest .coin/manifest.json --stage ${stageName}
        """
    }
}
