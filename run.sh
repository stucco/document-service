# run within docker container

etcd_ip=$(/sbin/ifconfig docker0 | sed -n '2 p' | awk '{print $2}' | cut -d":" -f2)

docker run -d -p 4001:4001 -p 7001:7001 flynn/etcd

docker run -d \
  -e ETCD_HOST=$etcd_ip
  -e ETCD_PORT=4001
  stucco/doc-service