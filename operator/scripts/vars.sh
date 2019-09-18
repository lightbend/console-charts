: "${VERSION:=latest}"
docker_registry=lightbend-docker-registry.bintray.io
image_name=lightbend/console-operator
full_docker_name="${docker_registry}/${image_name}:${VERSION}"
