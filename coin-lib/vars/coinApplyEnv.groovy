// Экспорт полей merged config в Jenkins env для shell-шагов и coin-executor.

/**
 * Записывает ключевые поля cfg в переменные окружения пайплайна.
 * coin-executor и sh-скрипты читают COIN_GP, COIN_PIN, COIN_API_URL и др.
 *
 * @param cfg merged config из coinLoadConfig
 */
def call(Map cfg) {
    if (cfg.coin?.goldenPath) {
        env.COIN_GP = cfg.coin.goldenPath.toString()
    }
    if (cfg.coin?.version) {
        env.COIN_PIN = cfg.coin.version.toString()
    }
    if (cfg.coin?.apiUrl) {
        env.COIN_API_URL = cfg.coin.apiUrl.toString()
    }
    if (cfg.jenkins?.registry?.prefix) {
        env.COIN_REGISTRY_PREFIX = cfg.jenkins.registry.prefix.toString()
    }
    if (cfg.jenkins?.credentials?.apiToken) {
        env.COIN_API_TOKEN_CRED = cfg.jenkins.credentials.apiToken.toString()
    }
}
