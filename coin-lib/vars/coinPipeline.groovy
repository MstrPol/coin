import org.coin.ci.Config
import org.coin.ci.PodTemplate
import org.coin.ci.StackExecutor
import org.coin.ci.StackImages

/**
 * Единая точка входа Coin CI.
 *
 * @param configPath путь к .coin/config.yaml
 * @param kubernetes true — K8s pod (stack-образ из config); false — agent any
 * @param cloud имя K8s cloud в Jenkins
 * @param prepareAgent label агента для checkout перед pod (по умолчанию любой доступный)
 */
def call(Map args = [:]) {
    def configPath = args.configPath ?: '.coin/config.yaml'
    def useKubernetes = args.kubernetes != false
    def cloudName = args.cloud ?: null
    def prepareAgent = args.prepareAgent ?: null

    if (!useKubernetes) {
        node {
            stage('Checkout') {
                checkout scm
            }
            def cfg = new Config(this).load(configPath)
            echo "Coin CI: project=${cfg.project.name} stack=${cfg.project.stack}"
            new StackExecutor(this).run(cfg)
        }
        return
    }

    def stackImage
    node(prepareAgent) {
        stage('Prepare') {
            checkout scm
            def cfg = new Config(this).load(configPath)
            stackImage = new StackImages(this).resolveStackImage(cfg)
            echo "Coin CI: project=${cfg.project.name} stack=${cfg.project.stack} image=${stackImage}"
        }
    }

    def images = new StackImages(this)
    def podYaml = new PodTemplate().build(images.jnlpImage(), stackImage)
    def podArgs = [yaml: podYaml, label: "coin-${env.BUILD_NUMBER}"]
    if (cloudName) {
        podArgs.cloud = cloudName
    }

    podTemplate(podArgs) {
        node(POD_LABEL) {
            stage('Checkout') {
                checkout scm
            }
            container('stack') {
                def cfg = new Config(this).load(configPath)
                new StackExecutor(this).run(cfg)
            }
        }
    }
}
