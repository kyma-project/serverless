#!/bin/sh

IMG_VERSION=${IMG_VERSION?"Define IMG_VERSION env"}

yq eval-all --inplace "
    select(fileIndex == 0).protecode=[
        select(fileIndex == 1)
        | (
            .global.images
            + {
                  \"serverless_operator\":{
                      \"name\":\"serverless-operator\",
                      \"directory\":\"prod\",
                      \"version\":\"${IMG_VERSION}\"
                  }
              }
          )[]
        | \"europe-docker.pkg.dev/kyma-project/\" + .directory + \"/\" + .name + \":\" + .version
    ]
    | select(fileIndex == 0)
    " sec-scanners-config.yaml config/serverless/values.yaml