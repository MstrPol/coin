package org.coin.ci

/**
 * Разрешает образ K8s-агента: project .coin/config.yaml + coin-platform.
 *
 * Источники (COIN_PLATFORM_DIR):
 *   golden-paths/<template>/<ver>/profile.yaml  → stack, runtime
 *   agents/catalog.yaml                           → image ref
 *   platform.yaml                                 → jnlp image
 */
class StackImages implements Serializable {

    private static final long serialVersionUID = 1L

    private final def steps
    private String platformDir
    private Map agentsCatalog
    private Map platformConfig

    StackImages(def steps) {
        this.steps = steps
    }

    String jnlpImage() {
        def jnlp = platformYaml().jenkins?.jnlp?.image
        if (!jnlp) {
                steps.error('platform.yaml: jenkins.jnlp.image не задан')
        }
        return jnlp
    }

    String resolveStack(Map cfg) {
        def pin = cfg.jenkins?.agent?.image
        if (pin) {
            return stackFromProfile(cfg) ?: 'pinned'
        }
        return stackFromProfile(cfg)
    }

    String resolveStackImage(Map cfg) {
        def pin = cfg.jenkins?.agent?.image
        if (pin) {
            return pin
        }

        def stack = stackFromProfile(cfg)
        def runtimeKey = runtimeKey(stack)
        def runtime = cfg.jenkins?.runtime?."${runtimeKey}" ?: runtimeFromProfile(cfg, stack)
        def entry = agentsCatalog().stacks?."${stack}"?."${runtime}"
        if (!entry?.image) {
            steps.error("agents/catalog.yaml: нет stacks.${stack}.${runtime}")
        }

        def registry = agentsCatalog().registry?.default
        if (!registry) {
            steps.error('agents/catalog.yaml: registry.default не задан')
        }
        def tag = entry.tag ?: "${runtime}-r${entry.rev ?: 0}"
        def ref = "${registry}/${entry.image}:${tag}"
        return entry.digest ? "${ref}@${entry.digest}" : ref
    }

    private String stackFromProfile(Map cfg) {
        def template = cfg.coin?.template
        if (!template) {
            steps.error('coin.template не задан в .coin/config.yaml')
        }
        def version = cfg.coin?.templateVersion ?: 'v1'
        def profilePath = "${platformRoot()}/golden-paths/${template}/${version}/profile.yaml"
        if (!steps.fileExists(profilePath)) {
            steps.error("profile не найден: ${profilePath}")
        }
        def profile = steps.readYaml(file: profilePath)
        def stack = profile.agent?.stack
        if (!stack) {
            steps.error("profile ${profilePath}: agent.stack не задан")
        }
        return stack
    }

    private String runtimeFromProfile(Map cfg, String stack) {
        def template = cfg.coin?.template
        def version = cfg.coin?.templateVersion ?: 'v1'
        def profilePath = "${platformRoot()}/golden-paths/${template}/${version}/profile.yaml"
        def profile = steps.readYaml(file: profilePath)
        def runtimeKey = runtimeKey(stack)
        def runtime = profile.agent?.runtime?."${runtimeKey}"
        if (!runtime) {
            steps.error("profile ${profilePath}: agent.runtime.${runtimeKey} не задан")
        }
        return runtime.toString()
    }

    private String platformRoot() {
        if (!platformDir) {
            platformDir = steps.env.COIN_PLATFORM_DIR
            if (!platformDir) {
                steps.error('COIN_PLATFORM_DIR не задан (checkout coin-platform в pipeline)')
            }
        }
        return platformDir
    }

    private Map agentsCatalog() {
        if (!agentsCatalog) {
            def path = "${platformRoot()}/agents/catalog.yaml"
            if (!steps.fileExists(path)) {
                steps.error("agents/catalog.yaml не найден: ${path}")
            }
            agentsCatalog = steps.readYaml(file: path)
        }
        return agentsCatalog
    }

    private Map platformYaml() {
        if (!platformConfig) {
            def path = "${platformRoot()}/platform.yaml"
            if (!steps.fileExists(path)) {
                steps.error("platform.yaml не найден: ${path}")
            }
            platformConfig = steps.readYaml(file: path)
        }
        return platformConfig
    }

    private static String runtimeKey(String stack) {
        switch (stack) {
            case 'python-uv':
            case 'python-pip':
                return 'python'
            case 'java-maven':
            case 'java-gradle':
                return 'java'
            default:
                return stack
        }
    }
}
