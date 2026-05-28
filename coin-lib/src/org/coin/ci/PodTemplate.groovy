package org.coin.ci

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
      resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          memory: "4Gi"
""".stripIndent()
    }
}
