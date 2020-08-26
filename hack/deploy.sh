#!/bin/bash

docker build -t davidtimm/anno .
docker push davidtimm/anno

ytt -f config/ -v system_namespace="$1" | kubectl apply -f -
kubectl -n $1 delete pods -l app=anno-proxy
