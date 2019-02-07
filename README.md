CFAR Logging Acceptance Tests [![slack.cloudfoundry.org][slack-badge]][loggregator-slack] [![CI Badge][ci-badge]][ci-pipeline]
=============================

## Usage

To run the CFAR logging acceptance tests you must have a user with permissions
to create orgs and spaces.

```
export CF_ADMIN_USER=<username>
export CF_ADMIN_PASSWORD=<password>
export CF_DOMAIN=<system_domain>
export SKIP_SSL_VALIDATION=false

go get -t ./...
go install github.com/onsi/ginkgo/ginkgo
ginkgo -race -r
```

[slack-badge]:              https://slack.cloudfoundry.org/badge.svg
[loggregator-slack]:        https://cloudfoundry.slack.com/archives/loggregator
[ci-badge]:                 https://loggregator.ci.cf-app.com/api/v1/pipelines/loggregator/jobs/cfar-lats/badge
[ci-pipeline]:              https://loggregator.ci.cf-app.com/

