#!/bin/sh

IMG_VERSION=${IMG_VERSION?"Define IMG_VERSION env"}

yq eval-all --inplace '
    select(fileIndex == 0).protecode=[
        select(fileIndex == 1)
        | .global.containerRegistry.path as $registryPath
        | (
            {
                "serverless_operator" : {
                    "name" : "serverless-operator",
                    "directory" : "prod",
                    "version" : env(IMG_VERSION)
                }
            }
            + .global.images
          )[]
        | $registryPath + "/" + .directory + "/" + .name + ":" + .version
    ]
    | select(fileIndex == 0)
    ' sec-scanners-config.yaml config/serverless/values.yaml