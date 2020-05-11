pipeline {

    agent {
      docker {
        customWorkspace '/tmp'
        image 'hehety/golang:1.12.13-alpine3.9-arm32v7'
        args '-v godfs_build_cache:/root/go'
      }
    }

    environment {
      GOPROXY = "https://goproxy.io"
    }

    stages {
      stage('Checkout') {
        steps {
          sh 'env'
          echo 'checkout repository...'
          checkout([$class: 'GitSCM', branches: [[name: '*/jenkins-pipeline']], doGenerateSubmoduleConfigurations: false, extensions: [], submoduleCfg: [], userRemoteConfigs: [[url: 'https://github.com/hetianyi/godfs.git']]])
        }
      }

      stage('Pull Dependencies') {
        steps {
          echo 'pull dependencies...'
          sh 'go mod tidy'
        }
      }

      stage('Build') {
        steps {
          echo 'build binary file...'
          sh 'go build -o bin/godfs main.go'
        }
      }

      stage('Clean') {
        steps {
          cleanWs()
        }
      }

    }
}