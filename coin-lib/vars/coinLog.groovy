// Структурированное логирование coin-lib в консоль Jenkins.
// Без static-полей — CPS sandbox не даёт читать их из vars-скриптов.

/**
 * Открывает визуальный блок стейджа: разделители + заголовок с emoji.
 */
def section(String emoji, String title) {
    echo ''
    echo '════════════════════════════════════════════════════'
    echo "${emoji}  Coin │ ${title}"
    echo '────────────────────────────────────────────────────'
}

/** Закрывает блок стейджа нижним разделителем. */
def sectionEnd() {
    echo '────────────────────────────────────────────────────'
    echo ''
}

/**
 * Строка лога с отступом и emoji.
 *
 * @param emoji символ-маркер уровня/типа сообщения
 * @param message текст сообщения
 */
def line(String emoji, String message) {
    echo "   ${emoji}  ${message}"
}

/**
 * Строка «ключ: значение» с отступом.
 */
def kv(String emoji, String key, String value) {
    echo "   ${emoji}  ${key}: ${value}"
}

/** Информационное сообщение. */
def info(String message) {
    line('ℹ️', message)
}

/** Успешное завершение шага или стейджа. */
def ok(String message) {
    line('✅', message)
}

/** Предупреждение (fallback, деградация). */
def warn(String message) {
    line('⚠️', message)
}

/** Пропуск шага (например, publish на non-tag build). */
def skip(String message) {
    line('⏭️', message)
}

/** Начало выполняемого действия. */
def step(String message) {
    line('▶️', message)
}

/** Шапка пайплайна в начале coinPipeline(). */
def banner() {
    echo ''
    echo '════════════════════════════════════════════════════'
    echo '🪙  Coin pipeline'
    echo '════════════════════════════════════════════════════'
}
