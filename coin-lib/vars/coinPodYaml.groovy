// Рендер K8s pod YAML из шаблона библиотеки и merged config.

/**
 * Подставляет образы, registry prefix и resource limits в coin-pod-template.yaml.
 * cfg передаётся как JSON-строка, чтобы не тащить LazyMap через CPS между нодами.
 *
 * @param cfgJson merged config, сериализованный в JSON
 * @return готовый YAML для podTemplate(yaml: ...)
 */
def call(String cfgJson) {
    def cfg = new groovy.json.JsonSlurper().parseText(cfgJson)
    def jnlpImage = cfg.jnlp?.image ?: cfg.jenkins?.jnlp?.image
    def stackImage = cfg.runtime?.image
    def registryPrefix = cfg.jenkins?.registry?.prefix ?: 'localhost:8082/coin-docker'
    def pod = cfg.pod ?: [:]
    def jnlpRes = pod.jnlp ?: [:]
    def stackRes = pod.stack ?: [:]
    def jnlpReq = jnlpRes.requests ?: [cpu: '100m', memory: '256Mi']
    def jnlpLim = jnlpRes.limits ?: [memory: '512Mi']
    def stackReq = stackRes.requests ?: [cpu: '500m', memory: '1Gi']
    def stackLim = stackRes.limits ?: [memory: '4Gi']

    def tpl = libraryResource('coin-pod-template.yaml')
    return tpl
        .replace('${JNLP_IMAGE}', jnlpImage.toString())
        .replace('${STACK_IMAGE}', stackImage.toString())
        .replace('${REGISTRY_PREFIX}', registryPrefix.toString())
        .replace('${JNLP_CPU_REQUEST}', jnlpReq.cpu.toString())
        .replace('${JNLP_MEMORY_REQUEST}', jnlpReq.memory.toString())
        .replace('${JNLP_MEMORY_LIMIT}', jnlpLim.memory.toString())
        .replace('${STACK_CPU_REQUEST}', stackReq.cpu.toString())
        .replace('${STACK_MEMORY_REQUEST}', stackReq.memory.toString())
        .replace('${STACK_MEMORY_LIMIT}', stackLim.memory.toString())
}
