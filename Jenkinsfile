pipeline {
  agent any
  
  environment {
    API_PORT='1313'
    WS_PORT='9002' 
    
  }

  stages {
		stage('Docker Shadow Build') {
			steps {
				sh 'docker compose -f shadow-compose.yaml build'
			}
		}

		stage('Docker Shadow Run') {
			steps {
				sh 'docker compose -f shadow-compose.yaml up'
			}
		}

		stage('Shadow Boxing') {
			parallel {
				stage('First Shadow') {
					steps{
						sh 'file_sufix=$(curl -i -X POST -H "Content-Type: multipart/form-data" -F "data=@shadows/1.png" https://onlinedi.vision:7377/upload | tail -n1) \
							curl https://127.0.0.1:7377/$file_sufix -o test1 \
							count=$(wc -l shadows/1.png | cut -f1 -d' ') \
							diffr=$(sdiff -B -b -s test screen.png | wc -l) \
							exit $(( $diffr*100/$count )) \
						'
					}
				}
				stage('Second Shadow') {
					steps{
						sh 'file_sufix=$(curl -i -X POST -H "Content-Type: multipart/form-data" -F "data=@shadows/2.png" https://onlinedi.vision:7377/upload | tail -n1) \
							curl https://127.0.0.1:7377/$file_sufix -o test2 \
							count=$(wc -l shadows/2.png | cut -f1 -d' ') \
							diffr=$(sdiff -B -b -s test screen.png | wc -l) \
							exit $(( $diffr*100/$count )) \
						'
					}
				}
				stage('Third Shadow') {
					steps {
						sh 'file_sufix=$(curl -i -X POST -H "Content-Type: multipart/form-data" -F "data=@shadows/3.png" https://onlinedi.vision:7377/upload | tail -n1) \
							curl https://127.0.0.1:7377/$file_sufix -o test3 \
							count=$(wc -l shadows/3.png | cut -f1 -d' ') \
							diffr=$(sdiff -B -b -s test screen.png | wc -l) \
							exit $(( $diffr*100/$count )) \
						'
					}
				}


			}
		}

		stage('Cleanup Shadow Arena') {
			steps {
				sh 'rm -rf ~/del_cdn'
			}
		}
			
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
