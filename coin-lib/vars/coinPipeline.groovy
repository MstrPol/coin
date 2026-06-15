// Точка входа Jenkins Shared Library: resolve → pod → creds → stages coin-executor.

/**
 * Возвращает отображаемое имя Jenkins stage из определения в manifest.
 * Приоритет: поле name, иначе capitalized id.
 */
def coinStageName(def stageDef) {
    return stageDef.name ?: stageDef.id?.toString()?.capitalize()
}

/**
 * Возвращает машинный id stage для coin-executor run --stage.
 * Приоритет: поле id, иначе slug из name (lowercase, пробелы → дефис).
 */
def coinStageID(def stageDef) {
    return stageDef.id ?: stageDef.name?.toString()?.toLowerCase()?.replaceAll(/\s+/, '-')
}

/**
 * Оборачивает body в withCredentials для Docker registry (build/publish).
 *
 * @param dockerCredId Jenkins credentialsId (usernamePassword)
 * @param body closure с шагами, использующими COIN_REGISTRY_USER/PASSWORD
 */
def coinWithDockerCredentials(String dockerCredId, Closure body) {
    withCredentials([usernamePassword(
        credentialsId: dockerCredId,
        usernameVariable: 'COIN_REGISTRY_USER',
        passwordVariable: 'COIN_REGISTRY_PASSWORD',
    )]) {
        body()
    }
}

/**
 * Выполняет body внутри K8s-контейнера stack (agent image из manifest).
 */
def coinInStack(Closure body) {
    container('stack') {
        body()
    }
}

/**
 * Главный пайплайн продукта: вызывается из Jenkinsfile как coinPipeline().
 *
 * Фаза 1 (built-in): read config → resolve manifest → merge config → pod YAML.
 * Фаза 2 (pod): checkout → materialize .coin/ → bootstrap executor → dynamic stages → report.
 * manifest/cfg между нодами передаются как JSON-строки (CPS serialization).
 */
def call() {
    coinLog.banner()

    def manifestJson
    def cfgJson
    def dockerCredId
    def nexusCredId
    def apiTokenCredId
    def executorUrl
    def jnlpImage
    def stackImage
    def podYaml

    node('built-in') {
        stage('Resolve manifest') {
            coinLog.section('🔍', 'Resolve manifest')

            def projectCfg = coinReadProjectConfig()
            def bootstrapCfg = coinLoadConfig([:], projectCfg)
            coinLog.kv('📌', 'Pin', "${bootstrapCfg.coin?.goldenPath} ${bootstrapCfg.coin?.version}")

            def manifest = coinResolveManifest(bootstrapCfg)
            def cfg = coinLoadConfig(manifest, projectCfg)
            coinApplyEnv(cfg)

            dockerCredId = cfg.jenkins?.credentials?.docker ?: 'nexus-docker'
            nexusCredId = cfg.jenkins?.credentials?.nexus ?: 'nexus-admin'
            apiTokenCredId = cfg.jenkins?.credentials?.apiToken ?: 'coin-api-token'
            executorUrl = cfg.executor?.url?.toString()
            jnlpImage = (cfg.jnlp?.image ?: cfg.jenkins?.jnlp?.image)?.toString()
            stackImage = cfg.runtime?.image?.toString()

            manifestJson = groovy.json.JsonOutput.toJson(manifest)
            cfgJson = groovy.json.JsonOutput.toJson(cfg)
            podYaml = coinPodYaml(cfgJson)

            coinLog.kv('🎯', 'Resolved GP', "${env.COIN_GP}@${env.COIN_GP_VERSION}")
            coinLog.kv('🤖', 'JNLP image', jnlpImage)
            coinLog.kv('📦', 'Stack image', stackImage)
            coinLog.ok('Manifest resolved, pod template ready')
            coinLog.sectionEnd()
        }
    }

    def podLabel = "coin-${env.BUILD_NUMBER}"

    podTemplate(yaml: podYaml, cloud: 'kubernetes', label: podLabel) {
        node(podLabel) {
            checkout scm
            def manifest = coinParseJson(manifestJson)
            def cfg = coinParseJson(cfgJson)
            coinMaterializeDotCoin(manifest, cfg)

            stage('Bootstrap') {
                coinLog.section('⚙️', 'Bootstrap')
                coinLog.step("Download coin-executor from Nexus")
                coinInStack {
                    withCredentials([usernamePassword(
                        credentialsId: nexusCredId,
                        usernameVariable: 'NEXUS_USER',
                        passwordVariable: 'NEXUS_PASS',
                    )]) {
                        sh """
                            set -eu
                            curl -fsS -u "\${NEXUS_USER}:\${NEXUS_PASS}" '${executorUrl}' -o coin-executor
                            chmod +x coin-executor
                            export PATH="${WORKSPACE}:\${PATH}"
                            coin-executor version
                        """
                    }
                }
                coinLog.ok('coin-executor ready')
                coinLog.sectionEnd()
            }

            def manifestStages = cfg.pipeline?.stages ?: []
            for (int i = 0; i < manifestStages.size(); i++) {
                def stageDef = manifestStages[i]
                def stageID = coinStageID(stageDef)
                def stageName = coinStageName(stageDef)
                if (!stageID) {
                    error "Coin: manifest stage is missing id/name: ${stageDef}"
                }
                stage(stageName) {
                    if (stageID == 'publish' && !env.TAG_NAME?.trim()) {
                        coinLog.section('⏭️', "Stage: ${stageName}")
                        coinLog.skip('Non-tag build — publish omitted')
                        coinLog.sectionEnd()
                    } else {
                        coinLog.section('▶️', "Stage: ${stageName}")
                        coinLog.step("coin-executor run --stage ${stageID}")
                        coinInStack {
                            if (stageID in ['build', 'publish']) {
                                coinWithDockerCredentials(dockerCredId) {
                                    if (stageID == 'build') {
                                        coinLog.info('Docker registry login')
                                        sh '''
                                            set -eu
                                            REG_HOST="${COIN_REGISTRY_PREFIX:-localhost:8082/coin-docker}"
                                            REG_HOST="${REG_HOST%%/*}"
                                            echo "${COIN_REGISTRY_PASSWORD}" | docker login "${REG_HOST}" \
                                              -u "${COIN_REGISTRY_USER}" --password-stdin || true
                                        '''
                                    }
                                    coinRunStage(stageID)
                                }
                            } else {
                                coinRunStage(stageID)
                            }
                        }
                        coinLog.ok("Stage ${stageName} finished")
                        coinLog.sectionEnd()
                    }
                }
            }

            stage('Report') {
                coinLog.section('📊', 'Report')
                coinLog.step('Send build report to coin-api')
                coinInStack {
                    def result = currentBuild.currentResult ?: 'SUCCESS'
                    withCredentials([string(credentialsId: apiTokenCredId, variable: 'COIN_API_TOKEN')]) {
                        sh """
                            set -eu
                            export PATH="${WORKSPACE}:\${PATH}"
                            export COIN_API_URL='${env.COIN_API_URL}'
                            export COIN_API_TOKEN="\${COIN_API_TOKEN}"
                            ./coin-executor report --manifest .coin/manifest.json \\
                              --project .coin/config.yaml \\
                              --build-url '${env.BUILD_URL}' --result '${result}'
                        """
                    }
                }
                coinLog.ok("Report submitted (result=${currentBuild.currentResult ?: 'SUCCESS'})")
                coinLog.sectionEnd()
            }
        }
    }
}
