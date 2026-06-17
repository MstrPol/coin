// Рендер K8s pod YAML из шаблона библиотеки и merged config.

/**
 * Подставляет runtime image, registry prefix и resource limits в coin-pod-template.yaml.
 * cfg передаётся как JSON-строка, чтобы не тащить LazyMap через CPS между нодами.
 *
 * @param cfgJson merged config, сериализованный в JSON
 * @return готовый YAML для podTemplate(yaml: ...)
 */
def call(String cfgJson) {
    def cfg = new groovy.json.JsonSlurper().parseText(cfgJson)
    def runtimeImage = cfg.runtime?.image
    def registryPrefix = cfg.jenkins?.registry?.prefix ?: 'localhost:8082/coin-docker'
    def buildEngine = cfg.build?.engine?.toString() ?: 'buildkit'
    def pod = cfg.pod ?: [:]
    def jnlpRes = pod.jnlp ?: [:]
    def jnlpReq = jnlpRes.requests ?: [cpu: '500m', memory: '1Gi']
    def jnlpLim = jnlpRes.limits ?: [memory: '4Gi']

    def procMount = '        procMount: Unmasked'
    def podmanVolumesBlock = '''  volumes:
    - name: podman-storage
      emptyDir:
        sizeLimit: 12Gi
'''
    def podmanVolumeMountsBlock = '''      volumeMounts:
        - name: podman-storage
          mountPath: /var/lib/containers/storage
'''

    def tpl = libraryResource('coin-pod-template.yaml')
    return tpl
        .replace('${RUNTIME_IMAGE}', runtimeImage.toString())
        .replace('${REGISTRY_PREFIX}', registryPrefix.toString())
        .replace('${COIN_BUILD_ENGINE}', buildEngine)
        .replace('${JNLP_CPU_REQUEST}', jnlpReq.cpu.toString())
        .replace('${JNLP_MEMORY_REQUEST}', jnlpReq.memory.toString())
        .replace('${JNLP_MEMORY_LIMIT}', jnlpLim.memory.toString())
        .replace('${PODMAN_VOLUMES_BLOCK}', podmanVolumesBlock)
        .replace('${PODMAN_VOLUME_MOUNTS_BLOCK}', podmanVolumeMountsBlock)
        .replace('${PROC_MOUNT}', procMount)
}
