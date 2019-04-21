#!/bin/bash

docker stop mysql mysql1 mysql2
docker rm mysql mysql1 mysql2

docker stop zookeeper kafka1 
docker rm zookeeper kafka1 

docker run -itd \
  -e MYSQL_ALLOW_EMPTY_PASSWORD=true \
  --name mysql \
  -v $(pwd)/binlog-1:/var/log/mysql \
  -v $(pwd)/schema:/docker-entrypoint-initdb.d \
  -p 0.0.0.0:33061:3306 \
  mysql:5.7 \
  mysqld \
  --datadir=/var/lib/mysql \
  --user=mysql \
  --server-id=1 \
  --log-bin=/var/log/mysql/mysql-bin.log \
  --gtid-mode=ON \
  --enforce-gtid-consistency=true

docker run -itd \
  -e MYSQL_ALLOW_EMPTY_PASSWORD=true \
  --name mysql1 \
  -v $(pwd)/schema:/docker-entrypoint-initdb.d \
  -p 0.0.0.0:33062:3306 \
  mysql:5.7 \
  mysqld \
  --datadir=/var/lib/mysql \
  --user=mysql \
  --server-id=2 

docker run -itd \
  -e MYSQL_ALLOW_EMPTY_PASSWORD=true \
  --name mysql2 \
  -v $(pwd)/binlog-1:/var/log/mysql \
  -v $(pwd)/schema:/docker-entrypoint-initdb.d \
  -p 0.0.0.0:33063:3306 \
  mysql:5.7 \
  mysqld \
  --datadir=/var/lib/mysql \
  --user=mysql \
  --server-id=3 

docker run --name zookeeper \
  -p 2181:2181 \
  -itd \
  -e ALLOW_ANONYMOUS_LOGIN=yes \
   bitnami/zookeeper:latest

docker run --name kafka1 \
  -itd \
  --link zookeeper:zookeeper \
  -e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
  -e ALLOW_PLAINTEXT_LISTENER=yes \
  -e KAFKA_ADVERTISED_HOST_NAME=192.168.0.26 \
  -e KAFKA_CREATE_TOPICS=test:1:2 \
  -p 0.0.0.0:9092:9092 \
   wurstmeister/kafka
