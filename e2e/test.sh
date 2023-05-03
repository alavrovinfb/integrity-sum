#!/bin/bash

# new file test
echo "running  new file test"
POD=$(kubectl get pod -l app=nginx-app -o jsonpath="{.items[0].metadata.name}")
kubectl -n default exec -it "$POD" -- touch /usr/bin/newfile

sleep 4
# file deleted test
echo "running  removing file"
POD=$(kubectl get pod -l app=nginx-app -o jsonpath="{.items[0].metadata.name}")
kubectl -n default exec -it "$POD" -- rm -f /usr/bin/tr


sleep 4
# file changed test
echo "running  file changed test"
POD=$(kubectl get pod -l app=nginx-app -o jsonpath="{.items[0].metadata.name}")
kubectl -n default exec -it "$POD" -- cp /usr/bin/cut /usr/bin/tr
