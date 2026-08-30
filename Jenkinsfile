def isReleaseTag() {
	return env.GIT_BRANCH ==~ /^refs\/tags\/\d+\.\d+\.\d+$/
}

pipeline {
	agent any

	options {
		disableConcurrentBuilds()
		timeout(time: 2, unit: 'HOURS')
	}

	stages {
		stage('Push Image to Docker Registry') {
			when {
				expression { isReleaseTag() }
			}

			steps {
				script {
					withDockerRegistry(
						url: 'https://registry.onlinedi.vision:5000',
						credentialsId: 'docker-registry'
					) {
						sh "docker buildx bake -f docker-bake.hcl --set release.output='type=registry'"
					}
				}
			}
		}
	}

	post {
		failure {
			emailext(
				from: 'jenkins@mail.onlinedi.vision',
				subject: "Build Failed: ${env.JOB_NAME} - ${env.BUILD_NUMBER}",
				body: "The build failed. Check the console output at ${env.BUILD_URL}.",
				to: 'TEAM@mail.onlinedi.vision'
			)
		}
	}
}
