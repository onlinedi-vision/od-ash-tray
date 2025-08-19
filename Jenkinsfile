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
		  	sh 'docker build -t ash .'
     	 }
	  }
   	stage('Docker Run') {
		  steps {
				 sh 'docker run -d -p 127.0.0.1:1313:1313 --name backend_container --env SCYLLA_CASSANDRA_PASSWORD=$SCYLLA_CASSANDRA_PASSWORD --env WS_PORT="9002" --env API_PORT="1313" ash:latest'
      }
	  }
  }
}
