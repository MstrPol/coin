import org.coin.ci.PodTemplate

/** Pod YAML для dynamic agent (jnlp + stack). */
def call(String jnlpImage, String stackImage) {
    return new PodTemplate().build(jnlpImage, stackImage)
}
