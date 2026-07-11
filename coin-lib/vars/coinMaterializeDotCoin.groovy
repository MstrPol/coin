// Запись resolved manifest и effective config в workspace pod-ноды.

/**
 * Создаёт каталог .coin/ и материализует артефакты для coin-executor.
 * manifest.json — полный resolved manifest; effective-config.yaml — merged Jenkins glue (debug).
 *
 * @param manifest resolved manifest из coinResolveManifest
 * @param cfg merged config из coinLoadConfig
 */
def call(Map manifest, Map cfg) {
    sh 'mkdir -p .coin'
    writeJSON file: '.coin/manifest.json', json: manifest
    writeYaml file: '.coin/effective-config.yaml', data: cfg
    coinLog.kv('💾', 'Manifest', '.coin/manifest.json')
    coinLog.kv('💾', 'Effective config', '.coin/effective-config.yaml')
    coinLog.ok('Workspace .coin/ materialized')
}
