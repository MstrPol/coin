import org.coin.ci.Config
import org.coin.ci.PodTemplate
import org.coin.ci.StackImages

/**
 * Единая точка входа Coin CI.
 *
 * Ответственность coin-lib:
 *   1. Подготовить K8s dynamic agent (выбор образа по stack из config).
 *   2. Забиндить credentials перед вызовом coin CLI.
 *
 * Вся бизнес-логика (версионирование, сборка, публикация, валидация) — в coin CLI.
 */
def call(Map args = [:]) {
    def cloudName    = args.cloud ?: null
    def prepareLabel = args.prepareAgent ?: null

    // ── Шаг 1: лёгкий агент — читаем минимум конфига для выбора образа ──
    def stackImage
    node(prepareLabel) {
        checkout scm
        def cfg = new Config(this).load()
        stackImage = new StackImages(this).resolveStackImage(cfg)
        echo "Coin: project=${cfg.project?.name} stack=${cfg.agent?.stack} agent=${stackImage}"
    }

    // ── Шаг 2: K8s pod с нужным toolchain-образом ──
    def podYaml = new PodTemplate().build(
        new StackImages(this).jnlpImage(),
        stackImage
    )
    def podArgs = [yaml: podYaml, label: "coin-${env.BUILD_NUMBER}"]
    if (cloudName) { podArgs.cloud = cloudName }

    podTemplate(podArgs) {
        node(POD_LABEL) {
            checkout scm

            container('stack') {
                // ── Шаг 3: валидация ──
                stage('Validate') {
                    sh 'coin validate'
                }

                // ── Шаг 4: тесты ──
                stage('Test') {
                    sh 'coin run test'
                }

                // ── Шаг 5: сборка ──
                stage('Build') {
                    sh 'coin run build'
                }

                // ── Шаг 6: публикация — credentials подготавливает Jenkins ──
                stage('Publish') {
                    def cfg       = new Config(this).load()
                    def credId    = cfg.agent?.publishRegistry ?: 'coin-registry-default'
                    withCredentials([usernamePassword(
                        credentialsId: credId,
                        usernameVariable: 'COIN_REGISTRY_USER',
                        passwordVariable: 'COIN_REGISTRY_PASSWORD',
                    )]) {
                        sh 'coin run publish'
                    }
                }
            }
        }
    }
}
