# grab address of MAIN_CLUSTER
MAIN_CLUSTER_HOST=$(docker ps -aqf 'name=rabbitmq_container_1')

# print it's address for debugging purpose
echo "MAIN CLUSTER HOST:" ${MAIN_CLUSTER_HOST}

# join node in cloud1 into self-hosted a.k.a MAIN
cat ./docker/join_script.sh | docker exec -e MAIN_CLUSTER_HOST=${MAIN_CLUSTER_HOST} -i rabbitmq_cloud1 sh

# join node in cloud2 into self-hosted a.k.a MAIN
cat ./docker/join_script.sh | docker exec -e MAIN_CLUSTER_HOST=${MAIN_CLUSTER_HOST} -i rabbitmq_cloud2 sh
