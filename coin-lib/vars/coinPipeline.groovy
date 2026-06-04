import org.coin.ci.Config
import org.coin.ci.CoinCli
import org.coin.ci.Platform
import org.coin.ci.PodTemplate
import org.coin.ci.ProfileLoader
import org.coin.ci.StackImages

/**
 * Единая точка входа Coin CI.
 *
 * Ответственность coin-lib:
 *   1. Checkout coin-platform + выбор K8s dynamic agent (GP profile bundle).
 *   2. Bootstrap coin CLI из Nexus (profile coinCli.version).
 *   3. Credentials binding перед coin run publish.
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

    def coinCli = new CoinCli(this)
    def coinBin = coinCli.binDir()

    def coinSh = { String script ->
        ansiColor('xterm') {
            sh "export PATH=\"${coinBin}:\$PATH\" && ${script}"
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
        def profile = new ProfileLoader(this).load(cfg)
        echo "Coin: project=${cfg.project?.name} template=${cfg.coin?.template}/${cfg.coin?.templateVersion ?: 'v1'} stack=${stack} agent=${stackImage} coinCli=${profile.coinCli?.version}"
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
                def cfg = new Config(this).load()
                def profile = new ProfileLoader(this).load(cfg)
                def minCLI = profile.coinCli?.version

                stage('Bootstrap') {
                    coinCli.bootstrap(cfg)
                }

                stage('version') {
                    coinSh 'coin --version'
                }

                stage('Validate') {
                    coinSh "coin validate --min-version ${minCLI}"
                }

                stage('Test') {
                    coinSh 'coin run test'
                }

                stage('Build') {
                    coinSh 'coin run build'
                }

                stage('Publish') {
                    def pubCfg = new Config(this).load()
                    def credId = pubCfg.jenkins?.credentials?.docker ?: 'coin-registry-default'
                    withCredentials([usernamePassword(
                        credentialsId: credId,
                        usernameVariable: 'COIN_REGISTRY_USER',
                        passwordVariable: 'COIN_REGISTRY_PASSWORD',
                    )]) {
                        coinSh '''
                          REG_HOST="${COIN_REGISTRY_PREFIX%%/*}"
                          echo "${COIN_REGISTRY_PASSWORD}" | docker login "${REG_HOST}" \
                            -u "${COIN_REGISTRY_USER}" --password-stdin
                          coin run publish
                        '''
                    }
                }
            }
        }
    }
}
