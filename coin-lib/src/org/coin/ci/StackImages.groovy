package org.coin.ci

/**
 * Разрешает образ K8s-агента по agent.stack и agent.runtime из config.
 * Читает только секцию agent: — всё остальное в конфиге не касается Jenkins.
 */
class StackImages implements Serializable {

    private static final long serialVersionUID = 1L

    private static final Map<String, String> DEFAULT_RUNTIME = [
        'python-uv'  : '3.13',
        'python-pip' : '3.13',
        'java-maven' : '17',
        'java-gradle': '17',
        'go'         : '1.22',
        'node'       : '20',
    ]

    private final def steps
    private Map imagesCatalog

    StackImages(def steps) {
        this.steps = steps
    }

    String jnlpImage() {
        return catalog().jnlp.image
    }

    String resolveStackImage(Map cfg) {
        def stack = cfg.agent?.stack
        if (!stack) {
            steps.error('agent.stack не задан в .coin/config.yaml')
        }
        def runtimeKey = runtimeKey(stack)
        def version = cfg.agent?.runtime?."${runtimeKey}" ?: DEFAULT_RUNTIME[stack]
        def entry = catalog().stacks?."${stack}"?."${version}"
        if (!entry?.image) {
            steps.error("Образ не найден для stack=${stack} version=${version} в images.yaml")
        }
        return entry.digest ? "${entry.image}@${entry.digest}" : entry.image
    }

    private Map catalog() {
        if (!imagesCatalog) {
            imagesCatalog = steps.readYaml(text: steps.libraryResource('images.yaml'))
        }
        return imagesCatalog
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
