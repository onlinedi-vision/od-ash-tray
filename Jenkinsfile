pipeline {
  agent any
  
  environment {
    API_PORT='1313'
    WS_PORT='9002' 
    
  }

  stages {
		stage('DEL_CDN Setup') {
			steps {
				sh 'stat ~/del_cdn 2> /dev/null > /dev/null || mkdir ~/del_cdn'
			}
		}
		stage('Docker Shadow Build') {
			steps {
				sh 'docker compose -f shadow-compose.yaml build'
			}
		}

		stage('Docker Shadow Run') {
			steps {
				withCredentials([vaultString(credentialsId:'vault-ash-key',variable:'ASH_TRAY_KEY')]){
					sh 'docker compose -f shadow-compose.yaml up -d'
				}
			}
		}

		stage('Shadow Boxing') {
			parallel {
				stage('First Shadow') {
					steps{
						sh 'file_sufix=$(curl -i -X POST -H "Content-Type: multipart/form-data" -F "data=@shadows/1.png" -k http://127.0.0.1:7377/ash/upload | tail -n1); \
							curl "http://127.0.0.1:7377/${file_sufix}" -k -o test1 ;\
							diff test1 shadows/1.png \
						'
					}
				}
				stage('Second Shadow') {
					steps{
						sh 'file_sufix=$(curl -i -X POST -H "Content-Type: multipart/form-data" -F "data=@shadows/2.png" -k http://127.0.0.1:7377/ash/upload | tail -n1); \
							curl -k "http://127.0.0.1:7377/${file_sufix}" -o test2; \
							diff test2 shadows/2.png \
						'
					}
				}
				stage('Third Shadow') {
					steps {
						sh 'file_sufix=$(curl -i -X POST -H "Content-Type: multipart/form-data" -F "data=@shadows/3.png" -k http://127.0.0.1:7377/ash/upload | tail -n1); \
							curl -k "http://127.0.0.1:7377/${file_sufix}" -o test3; \
							diff test3 shadows/3.png \
						'
					}
				}

				stage('4th Shadow') {
					steps {
						sh 'file_sufix=$(curl -i -X POST -H "Content-Type: multipart/form-data" -F "data=@shadows/4.png" -k http://127.0.0.1:7377/ash/upload | tail -n1); \
							curl -k "http://127.0.0.1:7377/${file_sufix}" -o test4; \
							diff test4 shadows/4.png \
						'
					}
				}

				stage('5th Shadow') {
					steps {
						sh 'file_sufix=$(curl -i -X POST -H "Content-Type: multipart/form-data" -F "data=@shadows/5.png" -k http://127.0.0.1:7377/ash/upload | tail -n1); \
							curl -k "http://127.0.0.1:7377/${file_sufix}" -o test5; \
							diff test5 shadows/5.png \
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
  }
}
