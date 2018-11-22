#!/bin/bash
docker build -t dolittle/k8s_certificate_manager_requester:latest-rsa .
echo "Remember to tag the image with a version as well as :latest"