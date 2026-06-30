// Слойная конфигурация Jenkins glue: defaults библиотеки → GP/manifest → project.

/**
 * Рекурсивно преобразует Map/List в CPS-сериализуемые LinkedHashMap/ArrayList.
 * Убирает LazyMap, который Jenkins не переносит между нодами.
 */
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

/**
 * Безопасно приводит значение к plain Map; null → пустая карта.
 */
@NonCPS
def toPlainMap(def value) {
    if (value == null) {
        return [:]
    }
    return toPlainValue(value)
}

/**
 * Глубокое слияние двух Map: вложенные Map объединяются, скаляры перезаписываются.
 * Пустые строки, пустые Map и null в override игнорируются.
 */
@NonCPS
def deepMerge(Map base, Map override) {
    if (override == null) {
        return base == null ? [:] : new LinkedHashMap(base)
    }
    def result = base == null ? [:] : new LinkedHashMap(base)
    override.each { key, value ->
        if (value == null) {
            return
        }
        if (value instanceof String && value.trim().isEmpty()) {
            return
        }
        if (value instanceof Map && value.isEmpty()) {
            return
        }
        if (value instanceof Map && result[key] instanceof Map) {
            result[key] = deepMerge((Map) result[key], (Map) value)
        } else {
            result[key] = value
        }
    }
    return result
}

/**
 * Извлекает из resolved manifest поля, нужные только для Jenkins glue (образы, executor, stages, creds).
 * Не копирует весь manifest — только runtime/executor/pipeline/credentials.
 */
@NonCPS
def manifestToConfig(Map manifest) {
    def layer = [:]
    if (manifest.build?.engine) {
        layer.build = [engine: manifest.build.engine.toString()]
    }
    if (manifest.runtime?.image) {
        layer.runtime = [image: manifest.runtime.image.toString()]
    }
    if (manifest.pipeline?.stages) {
        layer.pipeline = [stages: manifest.pipeline.stages]
    }
    if (manifest.credentials?.docker) {
        layer.jenkins = [credentials: [docker: manifest.credentials.docker.toString()]]
    }
    return layer
}

/**
 * Загружает defaults из resources/coin-lib-defaults.yaml и применяет env-override
 * (COIN_API_URL, COIN_MANIFEST_CACHE_BASE, COIN_RUNTIME_IMAGE).
 */
def loadLibDefaults() {
    def lib = readYaml text: libraryResource('coin-lib-defaults.yaml')
    lib = toPlainMap(lib)
    if (env.COIN_API_URL?.trim()) {
        lib.coin = (lib.coin ?: [:]) + [apiUrl: env.COIN_API_URL.trim()]
    }
    if (env.COIN_MANIFEST_CACHE_BASE?.trim()) {
        lib.coin = (lib.coin ?: [:]) + [manifestCacheBase: env.COIN_MANIFEST_CACHE_BASE.trim()]
    }
    if (env.COIN_RUNTIME_IMAGE?.trim()) {
        lib.runtime = (lib.runtime ?: [:]) + [image: env.COIN_RUNTIME_IMAGE.trim()]
    }
    return lib
}

/**
 * Точка входа: собирает effective config из трёх слоёв (lib < GP < project).
 *
 * @param manifest resolved manifest (слой GP)
 * @param projectCfg содержимое .coin/config.yaml (слой project)
 * @return merged Map, готовый к сериализации и materialize в effective-config.yaml
 */
def call(Map manifest, Map projectCfg) {
    def lib = loadLibDefaults()
    def gp = manifestToConfig(toPlainMap(manifest))
    def project = toPlainMap(projectCfg)
    def merged = deepMerge(lib, gp)
    merged = deepMerge(merged, project)
    return toPlainMap(merged)
}
