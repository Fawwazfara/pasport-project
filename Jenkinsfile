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
                echo 'Menjalankan verifikasi kode...'
                sh 'go vet ./...'
                sh 'go test -race ./...'
            }
        }
        stage('Build Binary') {
            steps {
                echo 'Membagikan/Membangun binary...'
                sh 'go build -o app .'
            }
        }
    }
}