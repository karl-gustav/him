export DOCKER_BUILDKIT=1
GPC_PROJECT_ID=my-cloud-collection
SERVICE_NAME=him
CONTAINER_NAME=europe-west1-docker.pkg.dev/$(GPC_PROJECT_ID)/cloud-run/$(SERVICE_NAME)

.PHONY: *

test:
	go test ./...

run:
	@echo 'Run this command:'
	@echo 'GOOGLE_APPLICATION_CREDENTIALS=~/<sa-credentials>.json\
	PORT=9991\
	go run .'


build: test
	docker build -t $(CONTAINER_NAME) .

push: build
	docker push $(CONTAINER_NAME)

deploy: push
	gcloud beta run deploy $(SERVICE_NAME)\
		--project $(GPC_PROJECT_ID)\
		--allow-unauthenticated\
		-q\
		--region europe-west1\
		--platform managed\
		--memory 128Mi\
		--image $(CONTAINER_NAME)
	
use-latest-version:
	gcloud alpha run services update-traffic $(SERVICE_NAME)\
		--to-latest\
		--project $(GPC_PROJECT_ID)\
		--region europe-west1\
		--platform managed
