# Jenkins CI

`ci/Jenkinsfile` defines the Jenkins CI pipeline for pull requests and the
`main` branch.

The pipeline performs the following checks:

- Builds and tests the Go project in Docker.
- Runs the shadow tests.
- Builds the application container image.
- Scans the image for HIGH and CRITICAL vulnerabilities with Trivy.
- Archives shadow-test logs.
