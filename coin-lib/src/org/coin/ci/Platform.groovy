package org.coin.ci

/**
 * Checkout coin-platform в workspace (для StackImages и coin CLI).
 */
class Platform implements Serializable {

    private static final long serialVersionUID = 1L

    private final def steps

    Platform(def steps) {
        this.steps = steps
    }

    String checkout(Map opts = [:]) {
        def relDir = opts.dir ?: '.coin/platform'
        def repoUrl = opts.repo ?: 'http://gitea:3000/coin/coin-platform.git'
        def branch = opts.branch ?: 'main'
        def credId = opts.credentialsId ?: 'gitea-git'

        steps.sh "mkdir -p '${relDir}'"
        steps.dir(relDir) {
            steps.checkout([
                $class           : 'GitSCM',
                branches         : [[name: "*/${branch}"]],
                extensions         : [[
                    $class: 'CloneOption',
                    depth : 1,
                    shallow: true,
                    noTags: true,
                ]],
                userRemoteConfigs: [[
                    url          : repoUrl,
                    credentialsId: credId,
                ]],
            ])
        }

        def absDir = "${steps.env.WORKSPACE}/${relDir}"
        steps.env.COIN_PLATFORM_DIR = absDir
        return absDir
    }
}
