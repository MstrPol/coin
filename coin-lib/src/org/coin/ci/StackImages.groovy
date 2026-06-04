package org.coin.ci

/**
 * Разрешает образ K8s-агента: GP profile (bundle) + agents/catalog.yaml.
 *
 * profile.yaml — единственный platform pin для продукта:
 *   agent.stack, agent.runtime, agent.rev  → image tag
 *   coinCli.version                         → Nexus Maven (CoinCli.groovy)
 */
class StackImages implements Serializable {

    private static final long serialVersionUID = 1L

    private final def steps
    private String platformDir
    private Map agentsCatalog

    StackImages(def steps) {
        this.steps = steps
    }

    String jnlpImage() {
        def path = "${platformRoot()}/platform.yaml"
        def platform = steps.readYaml(file: path)
        def jnlp = platform.jenkins?.jnlp?.image
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

        def profile = new ProfileLoader(steps).load(cfg)
        def stack = profile.agent?.stack
        if (!stack) {
            steps.error('profile: agent.stack не задан')
        }

        def runtimeKey = runtimeKey(stack)
        def runtime = cfg.jenkins?.runtime?."${runtimeKey}" ?: profile.agent?.runtime?."${runtimeKey}"
        if (!runtime) {
            steps.error("profile: agent.runtime.${runtimeKey} не задан")
        }
        runtime = runtime.toString()

        def profileRev = profile.agent?.rev
        if (profileRev == null) {
            steps.error('profile: agent.rev не задан (platform bundle)')
        }

        def entry = agentsCatalog().stacks?."${stack}"?."${runtime}"
        if (!entry?.image) {
            steps.error("agents/catalog.yaml: нет stacks.${stack}.${runtime}")
        }

        def registry = agentsCatalog().registry?.default
        if (!registry) {
            steps.error('agents/catalog.yaml: registry.default не задан')
        }

        def tag = resolveTag(entry, runtime.toString(), profileRev as int)
        return "${registry}/${entry.image}:${tag}"
    }

    private static String resolveTag(Map entry, String runtime, int profileRev) {
        def catalogRev = entry.rev != null ? entry.rev as int : 0
        if (entry.tag && catalogRev == profileRev) {
            return entry.tag.toString()
        }
        return "${runtime}-r${profileRev}"
    }

    private String stackFromProfile(Map cfg) {
        def profile = new ProfileLoader(steps).load(cfg)
        def stack = profile.agent?.stack
        if (!stack) {
            steps.error('profile: agent.stack не задан')
        }
        return stack
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
