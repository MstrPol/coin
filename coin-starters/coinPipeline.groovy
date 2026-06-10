// Coin orchestration — loaded from manifest.orchestration.url (Nexus).
// Product Jenkinsfile only resolves manifest and calls coinPipeline.run(this).

def coinPodYaml(String jnlpImage, String stackImage) {
    return """\
apiVersion: v1
kind: Pod
spec:
  containers:
    - name: jnlp
      image: ${jnlpImage}
      resources:
        requests:
          cpu: "100m"
          memory: "256Mi"
        limits:
          memory: "512Mi"
    - name: stack
      image: ${stackImage}
      command: ["sleep"]
      args: ["infinity"]
      tty: true
      securityContext:
        runAsUser: 0
      env:
        - name: COIN_REGISTRY_PREFIX
          value: "localhost:8082/coin-docker"
      volumeMounts:
        - name: docker-sock
          mountPath: /var/run/docker.sock
      resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          memory: "4Gi"
  volumes:
    - name: docker-sock
      hostPath:
        path: /var/run/docker.sock
        type: Socket
""".stripIndent()
}

def coinRunStage(String stageName) {
    sh """
        set -eu
        export PATH="\${WORKSPACE}:\${PATH}"
        export COIN_REGISTRY_PREFIX="\${COIN_REGISTRY_PREFIX:-localhost:8082/coin-docker}"
        coin-executor run --manifest .coin/manifest.json --stage ${stageName}
    """
}

def run(script) {
    def stackImage
    def dockerCredId = 'nexus-docker'
    unstash 'coin-manifest'
    def manifest = readJSON file: '.coin/manifest.json'
    stackImage = manifest.runtime.image
    if (manifest.credentials?.docker) {
        dockerCredId = manifest.credentials.docker
    }
    echo "Coin v2: gp=${env.COIN_GP}@${env.COIN_GP_VERSION} agent=${stackImage}"

    def jnlpImage = env.COIN_JNLP_IMAGE ?: 'jenkins/inbound-agent:3327.v868139a_d00e0-8'
    def podLabel = "coin-${env.BUILD_NUMBER}"

    podTemplate(yaml: coinPodYaml(jnlpImage, stackImage), cloud: 'kubernetes', label: podLabel) {
        node(podLabel) {
            checkout scm
            unstash 'coin-manifest'

            container('stack') {
                stage('Bootstrap') {
                    sh '''
                        set -eu
                        EXEC_URL="$(python3 -c 'import json; print(json.load(open(".coin/manifest.json"))["executor"]["url"])')"
                        curl -fsS "${EXEC_URL}" -o coin-executor
                        chmod +x coin-executor
                        export PATH="${WORKSPACE}:${PATH}"
                        coin-executor version
                    '''
                }

                stage('Validate') {
                    coinRunStage('validate')
                }

                stage('Test') {
                    coinRunStage('test')
                }

                stage('Build') {
                    withCredentials([usernamePassword(
                        credentialsId: dockerCredId,
                        usernameVariable: 'COIN_REGISTRY_USER',
                        passwordVariable: 'COIN_REGISTRY_PASSWORD',
                    )]) {
                        sh '''
                            set -eu
                            REG_HOST="${COIN_REGISTRY_PREFIX:-localhost:8082/coin-docker}"
                            REG_HOST="${REG_HOST%%/*}"
                            echo "${COIN_REGISTRY_PASSWORD}" | docker login "${REG_HOST}" \
                              -u "${COIN_REGISTRY_USER}" --password-stdin || true
                        '''
                        coinRunStage('build')
                    }
                }

                if (env.TAG_NAME?.trim()) {
                    stage('Publish') {
                        withCredentials([usernamePassword(
                            credentialsId: dockerCredId,
                            usernameVariable: 'COIN_REGISTRY_USER',
                            passwordVariable: 'COIN_REGISTRY_PASSWORD',
                        )]) {
                            coinRunStage('publish')
                        }
                    }
                }

                stage('Report') {
                    def result = currentBuild.currentResult ?: 'SUCCESS'
                    withCredentials([string(credentialsId: 'coin-api-token', variable: 'COIN_API_TOKEN')]) {
                        sh """
                            set -eu
                            export COIN_API_URL='${env.COIN_API_URL ?: 'http://coin-api:8090'}'
                            export COIN_API_TOKEN="\${COIN_API_TOKEN}"
                            ./coin-executor report --manifest .coin/manifest.json \\
                              --project .coin/config.yaml \\
                              --build-url '${env.BUILD_URL}' --result '${result}'
                        """
                    }
                }
            }
        }
    }
}

return this
