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
 * Главный пайплайн продукта: вызывается из Jenkinsfile как coinPipeline().
 *
 * Фаза 1 (built-in): read config → resolve manifest → merge config → pod YAML.
 * Фаза 2 (pod): checkout → materialize .coin/ → bootstrap buildkitd → dynamic stages → report.
 * manifest/cfg между нодами передаются как JSON-строки (CPS serialization).
 *
 * Параметры сборки: publish (boolean, default false) — управляет выполнением stage publish.
 */
def call() {
    properties([
        parameters([
            booleanParam(
                name: 'publish',
                defaultValue: false,
                description: 'Выполнить stage publish (публикация артефактов в registry)',
            ),
        ]),
    ])

    coinLog.banner()
    coinLog.kv('📋', 'Параметры', "publish=${params.publish}")

    // CPS: Map/объекты из node('built-in') не передаются в pod node — только примитивы/строки через внешний scope.
    def manifestJson   // resolved manifest → JSON для coinParseJson / coinMaterializeDotCoin в pod
    def cfgJson        // merged config → JSON для coinParseJson / coinPodYaml уже отработал на controller
    def dockerCredId   // Jenkins credential ID для withCredentials (build/publish)
    def apiTokenCredId // Jenkins credential ID для coin-executor report → coin-api
    def podYaml        // rendered K8s pod template → podTemplate(yaml: …)

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
            apiTokenCredId = cfg.jenkins?.credentials?.apiToken ?: 'coin-api-token'

            manifestJson = groovy.json.JsonOutput.toJson(manifest)
            cfgJson = groovy.json.JsonOutput.toJson(cfg)
            podYaml = coinPodYaml(cfgJson)

            coinLog.kv('🎯', 'Resolved GP', "${env.COIN_GP}@${env.COIN_GP_VERSION}")
            coinLog.kv('🤖', 'Runtime image', cfg.runtime?.image)
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
            env.COIN_PUBLISH_REQUEST = params.publish ? 'true' : 'false'
            coinMaterializeDotCoin(manifest, cfg)

            stage('Bootstrap') {
                coinLog.section('⚙️', 'Bootstrap')
                coinLog.step('Start podman in agent container')
                sh '''
                    set -eu
                    export PATH="/usr/local/bin:${PATH}"

                    if command -v podman >/dev/null 2>&1; then
                      if [ ! -S /var/run/docker.sock ]; then
                        mkdir -p /var/lib/containers/storage /run/containers/storage /run/podman
                        nohup podman system service --time=0 unix:///var/run/docker.sock \
                          >/tmp/podman.log 2>&1 &
                        echo $! >/tmp/podman.pid
                      fi

                      podman_ready=0
                      for i in $(seq 1 60); do
                        if [ -S /var/run/docker.sock ]; then
                          podman_ready=1
                          break
                        fi
                        if [ -f /tmp/podman.pid ] && ! kill -0 "$(cat /tmp/podman.pid)" 2>/dev/null; then
                          echo "podman service exited" >&2
                          tail -100 /tmp/podman.log >&2 || true
                          exit 1
                        fi
                        sleep 1
                      done

                      if [ "${podman_ready}" != 1 ]; then
                        echo "podman not ready after 60s" >&2
                        tail -100 /tmp/podman.log >&2 || true
                        exit 1
                      fi

                      if [ "${COIN_BUILD_ENGINE:-}" = "buildpack" ] && [ -s /usr/share/coin/paketo-builder.tar ]; then
                        if ! podman images --format '{{.Repository}}:{{.Tag}}' | grep -qx 'nexus:8082/coin-docker/paketo-builder-jammy-base:latest'; then
                          echo "==> load buildpack builder from /usr/share/coin/paketo-builder.tar"
                          podman load -i /usr/share/coin/paketo-builder.tar
                        fi
                      fi
                    fi

                    coin-executor version
                '''
                coinLog.ok('coin-agent ready (podman + coin-executor)')
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
                    if (stageID == 'publish' && !params.publish) {
                        coinLog.section('⏭️', "Stage: ${stageName}")
                        coinLog.skip('Параметр publish=false — stage пропущен')
                        coinLog.sectionEnd()
                    } else {
                        coinLog.section('▶️', "Stage: ${stageName}")
                        coinLog.step("coin-executor run --stage ${stageID}")
                        if (stageID in ['build', 'publish']) {
                            coinWithDockerCredentials(dockerCredId) {
                                coinLog.info('Configure registry auth for BuildKit')
                                coinConfigureRegistryAuth()
                                coinRunStage(stageID)
                            }
                        } else {
                            coinRunStage(stageID)
                        }
                        coinLog.ok("Stage ${stageName} finished")
                        coinLog.sectionEnd()
                    }
                }
            }

            stage('Report') {
                coinLog.section('📊', 'Report')
                coinLog.step('Send build report to coin-api')
                def result = currentBuild.currentResult ?: 'SUCCESS'
                withCredentials([string(credentialsId: apiTokenCredId, variable: 'COIN_API_TOKEN')]) {
                    sh """
                        set -eu
                        export PATH="/usr/local/bin:\${PATH}"
                        export COIN_API_URL='${env.COIN_API_URL}'
                        export COIN_API_TOKEN="\${COIN_API_TOKEN}"
                        coin-executor report --manifest .coin/manifest.json \\
                          --project .coin/config.yaml \\
                          --build-url '${env.BUILD_URL}' --result '${result}'
                    """
                }
                coinLog.ok("Report submitted (result=${currentBuild.currentResult ?: 'SUCCESS'})")
                coinLog.sectionEnd()
            }
        }
    }
}
