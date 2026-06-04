import org.coin.ci.Platform

/**
 * Checkout coin-platform → COIN_PLATFORM_DIR.
 * Вызов из Jenkinsfile platform/product jobs (классы src/ напрямую не import'ятся).
 */
def call(Map opts = [:]) {
    return new Platform(this).checkout(opts)
}
