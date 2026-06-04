package org.coin.ci

/**
 * Bootstrap coin CLI из Nexus Maven по pin из GP profile (coinCli.version).
 */
class CoinCli implements Serializable {

    private static final long serialVersionUID = 1L

    private final def steps

    CoinCli(def steps) {
        this.steps = steps
    }

    String binDir() {
        return "${steps.env.WORKSPACE}/.coin/bin"
    }

    void bootstrap(Map cfg) {
        def profile = new ProfileLoader(steps).load(cfg)
        def version = profile.coinCli?.version?.toString()?.trim()
        if (!version) {
            steps.error('profile: coinCli.version не задан (platform bundle)')
        }

        def platform = platformYaml()
        def floor = platform.coinCli?.min?.toString()?.trim()
        if (floor && version < floor) {
            steps.error("profile coinCli.version=${version} ниже platform coinCli.min=${floor}")
        }

        def mavenBase = platform.nexus?.mavenBase ?: 'http://nexus:8081/repository'
        def credId = platform.nexus?.credentialsId ?: 'nexus-admin'
        def repo = version.contains('SNAPSHOT') ? 'maven-snapshots' : 'maven-releases'
        def groupPath = 'coin/platform/coin-cli'
        def binDir = binDir()

        steps.withCredentials([steps.usernamePassword(
            credentialsId: credId,
            usernameVariable: 'NEXUS_USER',
            passwordVariable: 'NEXUS_PASSWORD',
        )]) {
            steps.sh """\
set -eu
ARCH=\$(uname -m)
case "\$ARCH" in
  arm64|aarch64) GOARCH=arm64 ;;
  x86_64|amd64) GOARCH=amd64 ;;
  *) echo "unsupported arch: \$ARCH" >&2; exit 1 ;;
esac
ZIP="${binDir}/coin-cli.zip"
mkdir -p "${binDir}"
REMOTE="${mavenBase}/${repo}/${groupPath}/${version}/coin-cli-${version}-linux-\${GOARCH}.zip"
curl -f -S -u "\${NEXUS_USER}:\${NEXUS_PASSWORD}" -o "\$ZIP" "\$REMOTE"
unzip -o -j "\$ZIP" -d "${binDir}"
chmod +x "${binDir}/coin"
rm -f "\$ZIP"
"${binDir}/coin" --version
""".stripIndent()
        }

        steps.echo "coin CLI ${version} → ${binDir}/coin"
    }

    private Map platformYaml() {
        def path = "${steps.env.COIN_PLATFORM_DIR}/platform.yaml"
        if (!steps.fileExists(path)) {
            steps.error("platform.yaml не найден: ${path}")
        }
        return steps.readYaml(file: path)
    }
}
