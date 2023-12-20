#!/bin/sh

IMG_VERSION=${IMG_VERSION"Define IMG_VERSION env"}

yq ".protecode[] |= sub(\":main\", \":${IMG_VERSION}\")" sec-scanners-config.yaml
