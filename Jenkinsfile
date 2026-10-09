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
                sh 'CGO_ENABLED=0 go build -o app ./cmd/server'
                archiveArtifacts artifacts: 'app'
            }
        }
    }
}