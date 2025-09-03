pipeline {
  agent any
  
  environment {
    API_PORT='1313'
    WS_PORT='9002' 
    
  }

  stages {
	  stage('Docker Kill') {
		  steps {
			  sh 'docker kill ash_container || echo "NO ALIVE CONTAINER"'
			  sh 'docker rm ash_container || echo "NO CONTAINER NAMED ash_container"'
		  }
	  }

	  stage('Docker Build') {
		  steps {
		  	sh 'docker compose build -f compose.yaml'
     	 }
	  }
   	stage('Docker Run') {
		  steps {
				sh 'docker compose up'
      }
	  }
  }
}
