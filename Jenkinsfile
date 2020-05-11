#!/usr/bin/groovy

pipeline {
    agent any

    stages {
        stage('Pull Dependencies') {
            steps {
                echo '下载依赖...'
                sh 'go mod tidy'
            }
        }

        stage('Build Binary') {
            steps {
                echo '构建可执行二进制文件...'
                sh 'go build -o bin/godfs main.go'
            }
        }

        stage('Build Docker Image') {
            steps {
                echo '构建Docker镜像...'
            }
            dir ('example') {
                /* 构建镜像 */
                def customImage = docker.build("hehety/godfs:${params.VERSION}-arm32v7")

                docker.withRegistry('https://index.docker.io', 'docker-registry') {
                    customImage.push()
                    customImage.push('latest')
                }
            }
        }
    }
}