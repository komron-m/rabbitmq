export MAIN_CLUSTER_HOST=$(docker ps -aqf 'name=rabbitmq1')

docker-up:
	docker-compose -f docker/docker-compose.yml up -d

docker-down:
	docker-compose -f docker/docker-compose.yml down -v

cluster:
	./docker/join_clusters.sh

.PHONY: docker-up docker-down cluster
