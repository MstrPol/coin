package org.coin.ci

class StackExecutor implements Serializable {

    private static final long serialVersionUID = 1L

    private final def steps
    private final Config config

    StackExecutor(def steps) {
        this.steps = steps
        this.config = new Config(steps)
    }

    void run(Map cfg) {
        if (config.stageEnabled(cfg, 'test')) {
            steps.stage('Test') {
                runStage(cfg, 'test')
            }
        }
        if (config.stageEnabled(cfg, 'build')) {
            steps.stage('Build') {
                withBuildEnv(cfg) {
                    runStage(cfg, 'build')
                }
            }
        }
        if (config.shouldPublish(cfg)) {
            steps.stage('Publish') {
                withPublishCredentials(cfg) {
                    runStage(cfg, 'publish')
                }
            }
        }
    }

    private static String buildTarget(Map cfg) {
        return cfg.pipeline?.build?.target ?: 'package'
    }

    private void withBuildEnv(Map cfg, Closure body) {
        def target = buildTarget(cfg)
        def name = cfg.project.name
        def version = new Versioning(steps).compute(cfg)
        def tag = version.imageTag
        def registry = steps.env.COIN_REGISTRY_PREFIX ?: ''
        def ref = registry ? "${registry}/${name}:${tag}" : "${name}:${tag}"
        def dockerfilePath = ''
        if (target == 'container') {
            dockerfilePath = renderManagedDockerfile(cfg)
        }
        steps.echo "Coin version: ${version.version} (${version.source}), image tag: ${version.imageTag}"
        steps.withEnv([
            "COIN_BUILD_TARGET=${target}",
            "COIN_VERSION=${version.version}",
            "COIN_VERSION_SOURCE=${version.source}",
            "COIN_IMAGE_NAME=${name}",
            "COIN_IMAGE_TAG=${tag}",
            "COIN_REGISTRY_PREFIX=${registry}",
            "COIN_IMAGE_REF=${ref}",
            "COIN_DOCKERFILE=${dockerfilePath}",
        ]) {
            body.call()
        }
    }

    private String renderManagedDockerfile(Map cfg) {
        if (steps.fileExists('Dockerfile')) {
            steps.error(
                "Dockerfile in service repository is not allowed. " +
                "Coin uses managed Dockerfiles from coin-lib/resources/dockerfiles/."
            )
        }

        def template = cfg.pipeline?.build?.dockerfileTemplate ?: cfg.project.stack
        def resourcePath = "dockerfiles/${template}/Dockerfile"
        def dockerfile = steps.libraryResource(resourcePath)

        def appPort = "${cfg.container?.port ?: 8080}"
        def appCommand = cfg.container?.command ?: defaultCommand(cfg.project.stack, cfg.project.name)
        def rendered = dockerfile
            .replace('{{PYTHON_VERSION}}', "${cfg.runtime?.python ?: '3.13'}")
            .replace('{{JAVA_VERSION}}', "${cfg.runtime?.java ?: '17'}")
            .replace('{{GO_VERSION}}', "${cfg.runtime?.go ?: '1.22'}")
            .replace('{{APP_PORT}}', appPort)
            .replace('{{APP_CMD}}', commandAsJson(appCommand))

        def path = '.coin/generated/Dockerfile'
        steps.sh 'mkdir -p .coin/generated'
        steps.writeFile(file: path, text: rendered)
        renderManagedDockerignore(template)
        steps.echo "Coin managed Dockerfile: ${resourcePath} -> ${path}"
        return path
    }

    private void renderManagedDockerignore(String template) {
        def resourcePath = "dockerignore/${template}/.dockerignore"
        try {
            def dockerignore = steps.libraryResource(resourcePath)
            // Docker/kaniko read .dockerignore from context root; generated workspace file is not committed.
            steps.writeFile(file: '.dockerignore', text: dockerignore)
            steps.echo "Coin managed .dockerignore: ${resourcePath} -> .dockerignore"
        } catch (Exception ignored) {
            steps.echo "No managed .dockerignore for template=${template}; build context will use project defaults"
        }
    }

    private static String commandAsJson(def command) {
        if (command instanceof List) {
            return '[' + command.collect { "\"${escapeJson("${it}")}\"" }.join(', ') + ']'
        }
        def raw = "${command}".trim()
        if (raw.startsWith('[')) {
            return raw
        }
        return '["sh", "-c", "' + escapeJson(raw) + '"]'
    }

    private static String escapeJson(String value) {
        return value.replace('\\', '\\\\').replace('"', '\\"')
    }

    private static def defaultCommand(String stack, String name) {
        switch (stack) {
            case 'python-uv':
            case 'python-pip':
                return ['python', '-m', name.replace('-', '_')]
            case 'go':
                return ['/app/app']
            case 'java-maven':
            case 'java-gradle':
                return ['java', '-jar', '/app/app.jar']
            default:
                return ['sh', '-c', 'echo "No default command configured" && sleep 3600']
        }
    }

    /**
     * Запускает стандартный сценарий из coin-lib.
     * Проект может добавить preCommands/postCommands или полностью заменить commands в .coin/config.yaml.
     */
    private void runStage(Map cfg, String stage) {
        runCommands(cfg.pipeline?."${stage}"?.preCommands)

        def overrideCommands = cfg.pipeline?."${stage}"?.commands
        if (overrideCommands) {
            runCommands(overrideCommands)
        } else {
            runStandardScript(cfg.project.stack, stage)
        }

        runCommands(cfg.pipeline?."${stage}"?.postCommands)
    }

    private void runStandardScript(String stack, String stage) {
        def resourcePath = "scripts/${stack}/${stage}.sh"
        def body = steps.libraryResource(resourcePath)
        def path = ".coin/generated/${stage}.sh"
        steps.sh 'mkdir -p .coin/generated'
        steps.writeFile(file: path, text: body)
        steps.sh(script: "chmod +x '${path}' && ./'${path}'")
    }

    private void runCommands(def commands) {
        if (!commands) {
            return
        }
        if (commands instanceof CharSequence) {
            steps.sh(script: "${commands}")
            return
        }
        commands.each { cmd ->
            steps.sh(script: "${cmd}")
        }
    }

    private void withPublishCredentials(Map cfg, Closure body) {
        def registry = cfg.pipeline?.publish?.registry ?: 'default'
        def credId = "coin-publish-${registry}"
        def url = publishUrl(registry, steps)

        def runBody = {
            if (url) {
                steps.withEnv(["UV_PUBLISH_URL=${url}"]) {
                    body.call()
                }
            } else {
                body.call()
            }
        }

        try {
            steps.withCredentials([
                steps.usernamePassword(
                    credentialsId: credId,
                    usernameVariable: 'UV_PUBLISH_USERNAME',
                    passwordVariable: 'UV_PUBLISH_PASSWORD',
                ),
            ]) {
                runBody.call()
            }
        } catch (Exception ignored) {
            steps.echo "Credential '${credId}' unavailable — publish without injected credentials"
            runBody.call()
        }
    }

    private static String publishUrl(String registry, def steps) {
        switch (registry) {
            case 'nexus-pypi':
                return steps.env.COIN_NEXUS_PYPI_URL ?: ''
            default:
                return steps.env.UV_PUBLISH_URL ?: ''
        }
    }
}
