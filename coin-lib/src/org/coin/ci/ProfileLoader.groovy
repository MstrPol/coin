package org.coin.ci

/**
 * Загрузка GP profile.yaml по coin.template + templateVersion из проекта.
 */
class ProfileLoader implements Serializable {

    private static final long serialVersionUID = 1L

    private final def steps

    ProfileLoader(def steps) {
        this.steps = steps
    }

    Map load(Map cfg) {
        def path = path(cfg)
        return steps.readYaml(file: path)
    }

    String path(Map cfg) {
        def template = cfg.coin?.template
        if (!template) {
            steps.error('coin.template не задан в .coin/config.yaml')
        }
        def version = cfg.coin?.templateVersion ?: 'v1'
        def profilePath = "${platformRoot()}/golden-paths/${template}/${version}/profile.yaml"
        if (!steps.fileExists(profilePath)) {
            steps.error("profile не найден: ${profilePath}")
        }
        return profilePath
    }

    private String platformRoot() {
        def dir = steps.env.COIN_PLATFORM_DIR
        if (!dir) {
            steps.error('COIN_PLATFORM_DIR не задан (checkout coin-platform в pipeline)')
        }
        return dir
    }
}
