// Resolve manifest через coin-api с fallback на Nexus pointer → blob. Без python.

/**
 * Извлекает фактическую версию GP из HTTP-заголовка X-Coin-Resolved-Version.
 *
 * @param headerText сырой текст response headers от curl -D
 * @return версия GP или null, если заголовок отсутствует
 */
@NonCPS
def coinResolvedVersion(String headerText) {
    def matcher = (headerText =~ /(?m)^X-Coin-Resolved-Version:\s*(.+)$/)
    return matcher.find() ? matcher.group(1).trim() : null
}

/**
 * URL-кодирует спецсимволы pin для пути Nexus metadata (Maven-style).
 *
 * @param pin значение coin.version из config (может содержать =, ~, ^)
 */
@NonCPS
def encodePin(String pin) {
    return pin.replace('=', '%3D').replace('~', '%7E').replace('^', '%5E')
}

/** Рекурсивно приводит Map/List к CPS-сериализуемым структурам. */
@NonCPS
def toPlainValue(def value) {
    if (value instanceof Map) {
        def out = new LinkedHashMap()
        value.each { key, entry -> out[key] = toPlainValue(entry) }
        return out
    }
    if (value instanceof List) {
        return value.collect { toPlainValue(it) }
    }
    return value
}

/** Безопасно приводит значение к plain Map; null → пустая карта. */
@NonCPS
def toPlainMap(def value) {
    if (value == null) {
        return [:]
    }
    return toPlainValue(value)
}

/**
 * Точка входа: resolve manifest по goldenPath + pin.
 *
 * Primary: GET coin-api /v1/golden-paths/{gp}/resolve?pin=...
 * Fallback: Nexus pointer JSON → blob URL, с проверкой manifestHash.
 * Устанавливает env.COIN_GP_VERSION при успехе.
 *
 * @param bootstrapCfg merged config до resolve (lib + project, manifest ещё пуст)
 * @return plain Map — resolved manifest
 */
def call(Map bootstrapCfg) {
    def gp = bootstrapCfg.coin?.goldenPath?.toString()
    def pin = bootstrapCfg.coin?.version?.toString()
    def coinApi = bootstrapCfg.coin?.apiUrl?.toString() ?: 'http://coin-api:8090'
    def nexusSnapshots = bootstrapCfg.coin?.manifestCacheBase?.toString() ?: 'http://nexus:8081/repository/maven-snapshots'
    if (!gp || !pin) {
        error 'Coin: coin.goldenPath and coin.version are required in project config'
    }

    def project = bootstrapCfg.project?.name?.toString() ?: ''
    def resolveUrl = "${coinApi}/v1/golden-paths/${gp}/resolve?pin=${pin}"
    if (project) {
        resolveUrl += "&project=${project}"
    }
    def pinEnc = encodePin(pin)
    def pointerUrl = "${nexusSnapshots}/coin/manifest/${gp}/metadata/${gp}-metadata-pin-${pinEnc}.json"

    coinLog.step("Resolve manifest: ${resolveUrl}")

    def manifest = null
    def resolvedVer = null

    dir('.coin-resolve') {
        sh 'mkdir -p .'

        def apiTokenCred = bootstrapCfg.jenkins?.credentials?.apiToken?.toString() ?: 'coin-api-token'
        def status = ''
        withCredentials([string(credentialsId: apiTokenCred, variable: 'COIN_API_TOKEN')]) {
            status = sh(
                script: """
                    set +e
                    if [ -n "\${COIN_API_TOKEN}" ]; then
                      code=\$(curl -s -o manifest.json -w '%{http_code}' \\
                        -H "Authorization: Bearer \${COIN_API_TOKEN}" \\
                        '${resolveUrl}' \\
                        -D resolve.headers)
                    else
                      code=\$(curl -s -o manifest.json -w '%{http_code}' \\
                        '${resolveUrl}' \\
                        -D resolve.headers)
                    fi
                    echo "\${code}"
                """,
                returnStdout: true,
            ).trim()
        }

        if (status == '200') {
            manifest = readJSON file: 'manifest.json'
            def headerText = fileExists('resolve.headers') ? readFile('resolve.headers') : ''
            resolvedVer = coinResolvedVersion(headerText)
            coinLog.ok("coin-api resolve HTTP ${status}")
        } else {
            coinLog.warn("coin-api resolve HTTP ${status}")
            coinLog.step('Fallback: Nexus pointer → manifest blob')
            coinLog.kv('🔗', 'Pointer', pointerUrl)
            sh "curl -fsS '${pointerUrl}' -o pointer.json"
            def pointer = readJSON file: 'pointer.json'
            sh "curl -fsS '${pointer.blobUrl}' -o manifest.json"
            manifest = readJSON file: 'manifest.json'
            if (pointer.manifestHash != manifest.manifestHash) {
                error "Coin: manifest hash mismatch (expected ${pointer.manifestHash}, got ${manifest.manifestHash})"
            }
            coinLog.ok('Nexus fallback manifest verified (hash match)')
        }
    }

    if (!resolvedVer) {
        resolvedVer = manifest.goldenPath?.version?.toString()
    }
    if (resolvedVer) {
        env.COIN_GP_VERSION = resolvedVer
    }
    coinLog.kv('🏷️', 'GP version', resolvedVer ?: 'unknown')

    return toPlainMap(manifest)
}
