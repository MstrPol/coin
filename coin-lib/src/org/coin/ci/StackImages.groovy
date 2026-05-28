package org.coin.ci

class StackImages implements Serializable {

    private static final long serialVersionUID = 1L

    private static final Map<String, String> DEFAULT_RUNTIME = [
        'python-uv': '3.13',
        'python-pip': '3.13',
        'java-maven': '17',
        'java-gradle': '17',
        'go': '1.22',
        'node': '20',
    ]

    private final def steps
    private Map imagesCatalog

    StackImages(def steps) {
        this.steps = steps
    }

    private Map catalog() {
        if (!imagesCatalog) {
            imagesCatalog = steps.readYaml(text: steps.libraryResource('images.yaml'))
        }
        return imagesCatalog
    }

    String jnlpImage() {
        return catalog().jnlp.image
    }

    String resolveStackImage(Map cfg) {
        def stack = cfg.project.stack
        def version = Config.runtimeVersion(cfg, runtimeKey(stack), DEFAULT_RUNTIME[stack])
        def stackEntry = catalog().stacks?."${stack}"?."${version}"
        if (!stackEntry?.image) {
            steps.error("No image for stack=${stack} version=${version} in coin-lib/resources/images.yaml")
        }
        def ref = stackEntry.image
        if (stackEntry.digest) {
            ref = "${ref}@${stackEntry.digest}"
        }
        return ref
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
