// Чтение .coin/config.yaml продукта без полного checkout, когда возможно.

/**
 * Загружает project config: сначала readTrusted (без checkout), иначе checkout scm.
 * readTrusted доступен в multibranch после первой синхронизации SCM metadata.
 *
 * @return Map — распарсенный .coin/config.yaml
 */
def call() {
    def projectCfg = null
    try {
        def text = readTrusted('.coin/config.yaml')
        projectCfg = readYaml text: text
        coinLog.ok('Project config loaded via readTrusted (no checkout)')
    } catch (Exception e) {
        coinLog.warn("readTrusted failed: ${e.message}")
        coinLog.step('Fallback: checkout scm → .coin/config.yaml')
    }
    if (!projectCfg) {
        checkout scm
        projectCfg = readYaml file: '.coin/config.yaml'
        coinLog.ok('Project config loaded from workspace checkout')
    }
    return projectCfg
}
