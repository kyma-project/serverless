#!/bin/sh

IMG_VERSION=${IMG_VERSION?"Define IMG_VERSION env"}

yq eval-all --inplace '
    select(fileIndex == 0).bdba=[
        (
            select(fileIndex == 1)
            | (
                .global.images + {
                    "serverless_operator" : "europe-docker.pkg.dev/kyma-project/prod/serverless-operator:"+ env(IMG_VERSION)
                }
            )[]
        )
        +
        (
            select(fileIndex == 2)
            | .global.images[]
        )
    ] 
    | select(fileIndex == 0).bdba = (select(fileIndex == 0).bdba | unique)
    | select(fileIndex == 0)
    ' sec-scanners-config.yaml config/serverless/values.yaml config/buildless-serverless/values.yaml
