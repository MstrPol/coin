package org.coin.ci

/**
 * Минимальное чтение .coin/config.yaml — только для оркестрации Jenkins.
 *
 * Jenkins читает исключительно секцию agent:
 *   agent.stack           — для выбора образа динамического агента
 *   agent.runtime         — для выбора версии образа
 *   agent.publishRegistry — для binding credentials перед публикацией
 *
 * Валидация, версионирование и вся бизнес-логика — в coin CLI.
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
