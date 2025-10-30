
docker_final:
	docker build -t ${DOCKER_IMAGE_NAME} .

registry_login:
	echo "${REGISTRY_PASSWORD}" | docker login -u ${REGISTRY_USER} --password-stdin 2>/dev/null || true

push_image:
	docker push ${DOCKER_IMAGE_NAME}

upload_docker_artifacts: registry_login docker_final push_image