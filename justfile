set positional-arguments

initdb:
    #!/usr/bin/env sh
    docker run -d --name distance_db \
        --network distance \
        -p 52342:3306 \
        -v /var/distance_mysql_dev:/var/lib/mysql \
        -v /home/ty001/distance-back/scripts/mysql:/docker-entrypoint-initdb.d \
        -e  MYSQL_ROOT_PASSWORD=my-secret-pw \
        mysql:latest

redis:
    #!/usr/bin/env sh
    docker run -d  --name distance_redis \
        --network=distance \
        -p 52343:6379 \
        -v /var/distance_redis_dev:/usr/local/etc/redis \
        redis redis-server /usr/local/etc/redis/redis.conf


install:
    #!/usr/bin/env sh
    GOBIN=$(pwd)/install go install github.com/chiliososada/distance-back/cmd/app

run: install
    $(pwd)/install/app

redisui:
    docker run -d --name  redisinsight --network=distance -p 52344:5540 redis/redisinsight:latest


push message:
    git add --all
    git commit -m "{{message}}"
    git push origin lzy