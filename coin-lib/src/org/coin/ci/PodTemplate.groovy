package org.coin.ci

/**
 * Pod template для K8s dynamic agents.
 * Локальный стенд: docker.sock + COIN_REGISTRY_PREFIX для coin run build/publish.
 */
class PodTemplate implements Serializable {

    private static final long serialVersionUID = 1L

    String build(String jnlpImage, String stackImage) {
        return """\
apiVersion: v1
kind: Pod
spec:
  containers:
    - name: jnlp
      image: ${jnlpImage}
      resources:
        requests:
          cpu: "100m"
          memory: "256Mi"
        limits:
          memory: "512Mi"
    - name: stack
      image: ${stackImage}
      command: ["sleep"]
      args: ["infinity"]
      tty: true
      securityContext:
        runAsUser: 0
      env:
        - name: COIN_REGISTRY_PREFIX
          value: "nexus:8082/coin-docker"
      volumeMounts:
        - name: docker-sock
          mountPath: /var/run/docker.sock
      resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          memory: "4Gi"
  volumes:
    - name: docker-sock
      hostPath:
        path: /var/run/docker.sock
        type: Socket
""".stripIndent()
    }
}
