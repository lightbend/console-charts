: "${VERSION:=latest}"
docker_registry=registry.lightbend.com
image_name=lightbend-console-operator
full_docker_name="${docker_registry}/${image_name}:${VERSION}"
