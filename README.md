## Online Division Ash Tray

### Short Description

Online Division Ash Tray (od-ash-tray) is the Content Delivery Network Node Software of the OD project.

### Quick Start

```
Missing :(
```

### Jenkins

GitHub Actions have been replaced by two Jenkins pipelines:

- `ci/Jenkinsfile` is the multibranch CI pipeline for pull requests and `main`.
- `Jenkinsfile` publishes images for numeric semantic-version tags such as
  `1.2.3`. It requires the Jenkins credential `docker-registry`.

CI agents require Go 1.24.5, Docker with Compose and Buildx, and access to the
Docker daemon for the Trivy scan and shadow tests.
