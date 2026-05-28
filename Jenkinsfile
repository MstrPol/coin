// Сборка образов Coin CI (опциональный job в монорепо coin)
pipeline {
    agent any

    parameters {
        string(name: 'REGISTRY', defaultValue: 'coin', description: 'Docker registry prefix')
    }

    stages {
        stage('Build images') {
            steps {
                dir('coin-images') {
                    sh "make build REGISTRY=${params.REGISTRY}"
                }
            }
        }
        stage('Push images') {
            when {
                branch pattern: 'main|master|release/.*', comparator: 'REGEXP'
            }
            steps {
                dir('coin-images') {
                    sh "make push REGISTRY=${params.REGISTRY}"
                }
            }
        }
    }
}
