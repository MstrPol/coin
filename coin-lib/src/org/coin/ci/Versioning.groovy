package org.coin.ci

class Versioning implements Serializable {

    private static final long serialVersionUID = 1L

    private final def steps

    Versioning(def steps) {
        this.steps = steps
    }

    Map compute(Map cfg) {
        def mode = cfg.coin?.versioning?.mode ?: 'corporate'
        if (mode != 'corporate') {
            steps.error("Unsupported versioning mode '${mode}'. Coin supports only corporate versioning.")
        }

        def tagPrefix = cfg.coin?.versioning?.tagPrefix ?: 'v'
        def tag = steps.env.TAG_NAME ?: ''
        def branch = steps.env.BRANCH_NAME ?: 'detached'
        def sha = (steps.env.GIT_COMMIT ?: 'local').take(8)
        def build = steps.env.BUILD_NUMBER ?: '0'

        // Релизную версию даёт только тег vX.Y.Z (единственный источник релиза).
        if (tag && tag ==~ /^${java.util.regex.Pattern.quote(tagPrefix)}\d+\.\d+\.\d+([-.][0-9A-Za-z.-]+)?$/) {
            def version = tag.replaceFirst("^${java.util.regex.Pattern.quote(tagPrefix)}", '')
            return [version: version, imageTag: dockerTag(version), source: "tag:${tag}"]
        }

        // Стабилизационная ветка release/<JIRA-ID> → release candidate.
        if (branch ==~ /^release\/.+/) {
            def safeId = branch.replaceFirst('^release/', '')
                .toLowerCase()
                .replaceAll(/[^0-9a-z.-]+/, '-')
                .replaceAll(/^-+|-+$/, '')
            def version = "0.0.0-rc.${safeId}.${build}+${sha}"
            return [version: version, imageTag: dockerTag(version), source: "release-branch:${branch}"]
        }

        def safeBranch = branch
            .toLowerCase()
            .replaceAll(/[^0-9a-z.-]+/, '-')
            .replaceAll(/^-+|-+$/, '')
        def version = "0.0.0-${safeBranch}.${build}+${sha}"
        return [version: version, imageTag: dockerTag(version), source: "branch:${branch}:${sha}"]
    }

    private static String dockerTag(String version) {
        return version.replace('+', '-').replaceAll(/[^0-9A-Za-z_.-]+/, '-')
    }
}

