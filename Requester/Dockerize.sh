#!/bin/bash
docker build -t dolittle/k8s_certificate_manager_requester .
echo "Remember to tag the image with a version as well as :latest"