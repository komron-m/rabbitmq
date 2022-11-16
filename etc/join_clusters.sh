# !/bin/sh
MAIN_CLUSTER_NAME=cluster_A
MAIN_CLUSTER_HOST=6a7a2ae5ce23

rabbitmqctl stop_app

rabbitmqctl reset

rabbitmqctl join_cluster ${MAIN_CLUSTER_NAME}@${MAIN_CLUSTER_HOST}

rabbitmqctl start_app
