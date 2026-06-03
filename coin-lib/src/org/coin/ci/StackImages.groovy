package org.coin.ci

/**
 * Разрешает образ K8s-агента по coin.template (или jenkins.stack) и runtime.
 * Маппинг template → stack — в images.yaml (секция templates).
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

    String resolveStack(Map cfg) {
        def template = cfg.coin?.template
        if (template) {
            def entry = catalog().templates?."${template}"
            if (entry?.stack) {
                return entry.stack
            }
        }
        def stack = cfg.jenkins?.stack
        if (!stack) {
            steps.error('stack не определён: задайте coin.template или jenkins.stack в .coin/config.yaml')
        }
        return stack
    }

    String resolveStackImage(Map cfg) {
        def stack = resolveStack(cfg)
        def runtimeKey = runtimeKey(stack)
        def runtime = cfg.jenkins?.runtime
        def version = runtime?."${runtimeKey}" ?: DEFAULT_RUNTIME[stack]
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
