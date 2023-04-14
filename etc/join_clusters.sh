# !/bin/sh
MAIN_CLUSTER_NAME=cluster_A
MAIN_CLUSTER_HOST=6a243d836813

rabbitmqctl stop_app

rabbitmqctl reset

rabbitmqctl join_cluster ${MAIN_CLUSTER_NAME}@${MAIN_CLUSTER_HOST}

rabbitmqctl start_app
