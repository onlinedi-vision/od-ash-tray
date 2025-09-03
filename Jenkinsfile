pipeline {
  agent any
  
  environment {
    API_PORT='1313'
    WS_PORT='9002' 
    
  }

  stages {
	  stage('Docker Kill') {
		  steps {
				sh 'docker compose down'
		  }
	  }

	  stage('Docker Build') {
		  steps {
		  	sh 'docker compose -f compose.yaml build'
     	 }
	  }
   	stage('Docker Run') {
		  steps {
				sh 'docker compose up -d'
      }
	  }
  }
}
