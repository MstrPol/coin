package org.coin.ci

import org.coin.ci.Semver

class Config implements Serializable {

    private static final long serialVersionUID = 1L

    private final def steps

    Config(def steps) {
        this.steps = steps
    }

    Map load(String configPath) {
        if (!steps.fileExists(configPath)) {
            steps.error("Coin config not found: ${configPath}")
        }
        def cfg = steps.readYaml(file: configPath)
        validate(cfg)
        return cfg
    }

    private void validate(Map cfg) {
        if (cfg.version != 1) {
            steps.error("Unsupported .coin config version: ${cfg.version} (expected 1)")
        }
        if (!cfg.project?.name || !cfg.project?.stack) {
            steps.error("project.name and project.stack are required in .coin/config.yaml")
        }

        enforceMinimumTemplateVersion(cfg)
    }

    private void enforceMinimumTemplateVersion(Map cfg) {
        def min = steps.env.COIN_MIN_TEMPLATE_VERSION ?: ''
        if (!min) {
            return
        }
        def tpl = cfg.coin?.template ?: ''
        def cur = cfg.coin?.templateVersion ?: ''
        if (!tpl || !cur) {
            steps.error(
                "Coin policy requires template pinning. Add coin.template and coin.templateVersion to .coin/config.yaml " +
                "(min required: ${min})."
            )
        }
        if (Semver.compare(cur, min) < 0) {
            steps.error(
                "Template ${tpl}@${cur} is ниже минимально допустимого (${min}). " +
                "Обновите golden path и поднимите coin.templateVersion."
            )
        }
    }

    boolean stageEnabled(Map cfg, String stage) {
        return cfg.pipeline?."${stage}"?.enabled != false
    }

    boolean shouldPublish(Map cfg) {
        def pub = cfg.pipeline?.publish
        if (!pub?.enabled) {
            return false
        }
        def when = pub.when ?: 'tag'
        def branch = steps.env.BRANCH_NAME ?: ''
        def tag = steps.env.TAG_NAME ?: ''
        switch (when) {
            case 'never':
                return false
            case 'always':
                return true
            case 'tag':
                return tag ==~ /v.*/ || branch ==~ /v.*/
            case 'branch':
                return branch in ['main', 'master']
            default:
                return false
        }
    }

    static String runtimeVersion(Map cfg, String key, String defaultVersion) {
        return cfg.runtime?."${key}" ?: defaultVersion
    }
}
