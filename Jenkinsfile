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
				sh 'cp /etc/letsencrypt/live/onlinedi.vision/fullchain.pem fullchain.pem'
				sh 'cp /etc/letsencrypt/live/onlinedi.vision/privkey.pem privkey.pem'
		  	sh 'docker compose -f compose.yaml build'
     	 }
	  }
   	stage('Docker Run') {
		  steps {
				 withCredentials([vaultString(credentialsId:'vault-ash-key',variable:'ASH_TRAY_KEY')]){
						sh 'docker compose up -d'
				}
      }
	  }
		stage('Certs Cleanup') {
			steps {
				sh 'rm privkey.pem fullchain.pem'
			}
		}
  }
}
