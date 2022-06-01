# All these commands must run from repository root.

run-prom2kafka:
	APP_NAME=prom2kafka \
	# ARGS='-storageNode=localhost:8400' \
	$(MAKE) run-via-docker

prom2kafka:
	APP_NAME=prom2kafka $(MAKE) app-local

prom2kafka-race:
	APP_NAME=prom2kafka RACE=-race $(MAKE) app-local

prom2kafka-prod:
	APP_NAME=prom2kafka $(MAKE) app-via-docker

prom2kafka-prod-race:
	APP_NAME=prom2kafka RACE=-race $(MAKE) app-via-docker

prom2kafka-pure:
	APP_NAME=prom2kafka $(MAKE) app-local-pure

prom2kafka-pure-prod:
	APP_NAME=prom2kafka APP_SUFFIX='-pure' DOCKER_OPTS='--env CGO_ENABLED=0' $(MAKE) app-via-docker

package-prom2kafka:
	APP_NAME=prom2kafka $(MAKE) package-via-docker

package-prom2kafka-race:
	APP_NAME=prom2kafka RACE=-race $(MAKE) package-via-docker

publish-prom2kafka:
	APP_NAME=prom2kafka $(MAKE) publish-via-docker

publish-prom2kafka-race:
	APP_NAME=prom2kafka RACE=-race $(MAKE) publish-via-docker
