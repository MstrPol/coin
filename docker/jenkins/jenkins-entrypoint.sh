#!/usr/bin/env bash
set -euo pipefail

KUBECONFIG_PATH="${KUBECONFIG_PATH:-/kubeconfig/config}"
JENKINS_TOKEN_PATH="${JENKINS_TOKEN_PATH:-/kubeconfig/jenkins-token}"
CASC_DIR="/var/jenkins_home/casc-config"
INIT_DIR="/var/jenkins_home/init.groovy.d"

resolve_kubeconfig() {
  if [[ -f "${KUBECONFIG_PATH}" ]]; then
    echo "${KUBECONFIG_PATH}"
    return
  fi
  if [[ -f /kubeconfig/kubeconfig.yaml ]]; then
    echo /kubeconfig/kubeconfig.yaml
    return
  fi
  echo ""
}

for i in $(seq 1 120); do
  if [[ -n "$(resolve_kubeconfig)" && -f "${JENKINS_TOKEN_PATH}" ]]; then
    break
  fi
  echo "waiting for kubeconfig and jenkins token in /kubeconfig..."
  sleep 2
done

if [[ -z "$(resolve_kubeconfig)" ]]; then
  echo "kubeconfig not found in /kubeconfig (expected config or kubeconfig.yaml)" >&2
  exit 1
fi
if [[ ! -f "${JENKINS_TOKEN_PATH}" ]]; then
  echo "jenkins token not found: ${JENKINS_TOKEN_PATH} (run: make k8s-auth)" >&2
  exit 1
fi

mkdir -p "${CASC_DIR}" "${INIT_DIR}"
cp /usr/share/jenkins/ref/casc.yaml "${CASC_DIR}/00-base.yaml"
cp /usr/share/jenkins/ref/casc-jobs.yaml "${CASC_DIR}/10-jobs.yaml"

# k8s cloud — bearer token (k3s client cert = EC key, fabric8 не парсит через CertificateCredentials)
cat > "${INIT_DIR}/50-kubernetes.groovy" <<'GROOVY'
import com.cloudbees.plugins.credentials.CredentialsScope
import com.cloudbees.plugins.credentials.domains.Domain
import com.cloudbees.plugins.credentials.SystemCredentialsProvider
import hudson.util.Secret
import jenkins.model.Jenkins
import org.csanchez.jenkins.plugins.kubernetes.KubernetesCloud
import org.jenkinsci.plugins.plaincredentials.impl.StringCredentialsImpl

def credId = 'k3s-token'
def tokenText = new File('/kubeconfig/jenkins-token').text.trim()

def store = SystemCredentialsProvider.getInstance().getStore()
def domain = Domain.global()

['k3s-token', 'k3s-client-cert'].each { id ->
    def old = store.getCredentials(domain).find { it.id == id }
    if (old != null) {
        store.removeCredentials(domain, old)
    }
}

store.addCredentials(domain, new StringCredentialsImpl(
    CredentialsScope.GLOBAL,
    credId,
    'k3s API token (local stack)',
    Secret.fromString(tokenText)
))

def jenkins = Jenkins.getInstance()
def cloud = jenkins.getCloud('kubernetes') as KubernetesCloud
if (cloud == null) {
    cloud = new KubernetesCloud('kubernetes')
    jenkins.clouds.add(cloud)
}
cloud.setServerUrl('https://k3s:6443')
cloud.setSkipTlsVerify(true)
cloud.setCredentialsId(credId)
cloud.setJenkinsUrl('http://jenkins:8080')
cloud.setJenkinsTunnel('jenkins:50000')
cloud.setNamespace('default')
cloud.setContainerCapStr('10')
cloud.setConnectTimeout(300)
cloud.setReadTimeout(300)
cloud.setMaxRequestsPerHostStr('32')
jenkins.save()
GROOVY

export CASC_JENKINS_CONFIG="${CASC_DIR}"

exec /usr/local/bin/jenkins.sh
