pipeline {
  agent {
    label 'go'
  }
  stages {
    stage('Build') {
      sh 'go test -v'
    }
  }
}