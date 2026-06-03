import groovy.json.JsonOutput

// Парсит stacks.<stack>.<runtime> из catalog.yaml (минимальный YAML-парсер).
def parseCatalogStacks(String text) {
    def stacks = [:].withDefault { [] }
    def idx = text.indexOf('stacks:')
    if (idx < 0) {
        return stacks
    }
    String current = null
    text.substring(idx).eachLine { line ->
        def sm = (line =~ /^  ([a-z0-9-]+):\s*$/)
        if (sm.matches()) {
            current = sm.group(1)
            stacks[current] = []
            return
        }
        def rm = (line =~ /^    "([^"]+)":/)
        if (rm.find() && current) {
            stacks[current] << rm.group(1)
        }
    }
    return stacks
}

def defaultStacks() {
    return [
        'go'         : ['1.22'],
        'python-uv'  : ['3.13'],
        'python-pip' : ['3.13'],
        'java-maven' : ['17'],
        'java-gradle': ['17'],
    ]
}

def loadCatalogStacks() {
    def path = '/var/jenkins_home/coin-agents-catalog.yaml'
    def file = new File(path)
    if (!file.exists()) {
        return defaultStacks()
    }
    def parsed = parseCatalogStacks(file.text)
    return parsed.isEmpty() ? defaultStacks() : parsed
}

def stacks = loadCatalogStacks()
def stackChoices = JsonOutput.toJson(stacks.keySet().sort())

def runtimeScript = stacks.collect { stack, runtimes ->
    "if (STACK == '${stack}') return ${JsonOutput.toJson(runtimes.sort())}"
}.join('\nelse ') + '\nelse return []'

pipelineJob('coin-agents') {
    description('Platform: CI agent images (catalog.yaml → registry)')
    parameters {
        booleanParam('BUILD_ALL', false, 'Собрать все stack/runtime из catalog.yaml')
        choiceParameter {
            name('STACK')
            description('Toolchain stack (из catalog.yaml)')
            filterable(false)
            choiceType('PT_SINGLE_SELECT')
            script {
                groovyScript {
                    script {
                        script("return ${stackChoices}")
                        sandbox(true)
                    }
                    fallbackScript {
                        script('return ["go"]')
                        sandbox(true)
                    }
                }
            }
            randomName('')
            filterLength(0)
        }
        cascadeChoiceParameter {
            name('RUNTIME')
            description('Runtime version (зависит от STACK)')
            filterable(false)
            choiceType('PT_SINGLE_SELECT')
            referencedParameters('STACK')
            script {
                groovyScript {
                    script {
                        script(runtimeScript)
                        sandbox(true)
                    }
                    fallbackScript {
                        script('return []')
                        sandbox(true)
                    }
                }
            }
            randomName('')
            filterLength(0)
        }
    }
    definition {
        cpsScm {
            scm {
                git {
                    remote {
                        url('http://gitea:3000/coin/coin.git')
                        credentials('gitea-git')
                    }
                    branch('main')
                }
            }
            scriptPath('coin-jenkins-agents/Jenkinsfile')
        }
    }
}
