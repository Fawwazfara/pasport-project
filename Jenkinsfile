pipeline {
    agent any

    options {
        timestamps()
        disableConcurrentBuilds()
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        stage('Verify') {
            steps {
                sh 'go vet ./...'
                sh 'go test -race ./...'
            }
        }
        stage('Build Binary') {
            steps {
                sh 'go build -o app .'
            }
        }
    }
}