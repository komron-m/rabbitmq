rabbitmqctl stop_app

rabbitmqctl reset

rabbitmqctl join_cluster rabbitmq1@${MAIN_CLUSTER_HOST}

rabbitmqctl start_app
