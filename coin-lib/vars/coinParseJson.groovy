// Парсинг JSON-строки в CPS-сериализуемые структуры (между built-in и pod-нодой).

/**
 * Рекурсивно заменяет LazyMap/LazyList на LinkedHashMap/ArrayList.
 */
@NonCPS
def toPlainValue(def value) {
    if (value instanceof Map) {
        def out = new LinkedHashMap()
        value.each { key, entry -> out[key] = toPlainValue(entry) }
        return out
    }
    if (value instanceof List) {
        return value.collect { toPlainValue(it) }
    }
    return value
}

/**
 * Парсит JSON-текст в plain Map/List.
 * Используется после передачи manifestJson/cfgJson как String между нодами.
 *
 * @param jsonText сериализованный JSON (из JsonOutput.toJson)
 * @return LinkedHashMap или List, безопасный для CPS
 */
def call(String jsonText) {
    def parsed = new groovy.json.JsonSlurper().parseText(jsonText)
    return toPlainValue(parsed)
}
