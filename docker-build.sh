#!/bin/bash
docker network create front
docker run -d -p  8080:80 -v $HOME/cluster-build/chapter_1/src/go/webapp:/go/src/webapp -v $HOME/cluster-build/chapter_1/src/sh:/sh --network front golang sh /sh/start.sh
docker run -d --name=mysql -p 3306:3306  --env MYSQL_ROOT_PASSWORD=123456 --network front mysql             
