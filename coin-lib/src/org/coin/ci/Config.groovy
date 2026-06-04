package org.coin.ci

/**
 * Минимальное чтение .coin/config.yaml — только для оркестрации Jenkins.
 *
 * Jenkins читает:
 *   coin.template          — golden path (stack/runtime из coin-platform)
 *   jenkins.agent.image    — optional pin образа агента
 *   jenkins.runtime        — optional override версии toolchain
 *   jenkins.credentials    — binding credentials перед публикацией
 *
 * Образ агента: project config + COIN_PLATFORM_DIR (agents/catalog.yaml, GP profile).
 * Валидация и бизнес-логика — в coin CLI.
 */
class Config implements Serializable {

    private static final long serialVersionUID = 1L

    private final def steps

    Config(def steps) {
        this.steps = steps
    }

    Map load(String configPath = '.coin/config.yaml') {
        if (!steps.fileExists(configPath)) {
            steps.error("Coin config not found: ${configPath}")
        }
        return steps.readYaml(file: configPath)
    }
}
