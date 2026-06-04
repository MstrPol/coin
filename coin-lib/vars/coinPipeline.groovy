import org.coin.ci.Config
import org.coin.ci.Platform
import org.coin.ci.PodTemplate
import org.coin.ci.StackImages

/**
 * Единая точка входа Coin CI.
 *
 * Ответственность coin-lib:
 *   1. Checkout coin-platform + выбор K8s dynamic agent.
 *   2. Credentials binding перед coin CLI.
 */
def call(Map args = [:]) {
    def cloudName    = args.cloud ?: null
    def prepareLabel = args.prepareAgent ?: null
    def platformOpts = [
        dir          : args.platformDir ?: '.coin/platform',
        repo         : args.platformRepo ?: 'http://gitea:3000/coin/coin-platform.git',
        branch       : args.platformBranch ?: 'main',
        credentialsId: args.platformCredentials ?: 'gitea-git',
    ]

    def coinSh = { String script ->
        ansiColor('xterm') {
            sh script
        }
    }

    def stackImage
    def jnlpImage
    node(prepareLabel) {
        checkout scm
        new Platform(this).checkout(platformOpts)
        def cfg = new Config(this).load()
        def images = new StackImages(this)
        stackImage = images.resolveStackImage(cfg)
        jnlpImage = images.jnlpImage()
        def stack = images.resolveStack(cfg)
        echo "Coin: project=${cfg.project?.name} template=${cfg.coin?.template} stack=${stack} agent=${stackImage}"
    }

    def podYaml = new PodTemplate().build(jnlpImage, stackImage)
    def podLabel = "coin-${env.BUILD_NUMBER}"
    def podArgs = [yaml: podYaml, label: podLabel]
    if (cloudName) { podArgs.cloud = cloudName }

    podTemplate(podArgs) {
        node(podLabel) {
            checkout scm
            new Platform(this).checkout(platformOpts)

            container('stack') {
                stage('version') {
                    coinSh 'coin --version'
                }

                stage('Validate') {
                    coinSh 'coin validate'
                }

                stage('Test') {
                    coinSh 'coin run test'
                }

                stage('Build') {
                    coinSh 'coin run build'
                }

                stage('Publish') {
                    def cfg = new Config(this).load()
                    def credId = cfg.jenkins?.credentials?.docker ?: 'coin-registry-default'
                    withCredentials([usernamePassword(
                        credentialsId: credId,
                        usernameVariable: 'COIN_REGISTRY_USER',
                        passwordVariable: 'COIN_REGISTRY_PASSWORD',
                    )]) {
                        coinSh 'coin run publish'
                    }
                }
            }
        }
    }
}
