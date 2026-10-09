// pipeline {
//     agent none

//     options {
//         timestamps()
//         disableConcurrentBuilds()
//     }

//     stages {
//         stage('Checkout') {
//             agent any
//             steps {
//                 checkout scm
//                 stash name: 'src', includes: '**/*'
//             }
//         }

//         stage('Verify') {
//             agent {
//                 docker { image 'golang:1.22'; args '-u root' }
//             }
//             steps {
//                 unstash 'src'
//                 sh 'go vet ./...'
//                 sh 'go test -race ./...'
//             }
//         }

//         stage('Build Binary') {
//             agent {
//                 docker { image 'golang:1.22'; args '-u root' }
//             }
//             steps {
//                 unstash 'src'
//                 sh 'CGO_ENABLED=0 go build -o bin/passport ./cmd/server'
//                 archiveArtifacts artifacts: 'bin/passport'
//             }
//         }

//         stage('Build Image') {
//             agent any
//             steps {
//                 unstash 'src'
//                 sh 'docker build -t passport-app:${BUILD_NUMBER} .'
//             }
//         }
//     }
// }


pipeline {
    agent any

    stages {
        stage('Check Environment') {
            steps {
                echo "Menjalankan build untuk branch: ${env.BRANCH_NAME}"
                sh 'go version'
            }
        }
        stage('Build App') {
            steps {
                echo 'Memproses kompilasi kode Go...'
                sh 'go build -o myapp main.go'
            }
        }
    }
}