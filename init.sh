#!/bin/bash

docker stop mysql mysql1 mysql2
docker rm mysql mysql1 mysql2

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
  --binlog_do_db=test \
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
  --server-id=1 

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
  --server-id=1 \
  --log-bin=/var/log/mysql/mysql-bin.log \
  --binlog_do_db=test \
  --gtid-mode=ON \
  --enforce-gtid-consistency=true
