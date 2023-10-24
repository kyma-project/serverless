# Overview

The function uses [Deep AI API](https://deepai.org/machine-learning-model/text2img) for image generation from text.
It's purpose is to demonstrate how to achieve basic development tasks as a python function developer:

 - How to consume request
 - How to customise response
 - How to read ENVs
 - How to configure libraries as dependencies
 - How to send requests from the function code
 - How to configure function in the Kyma context
   - function API exposure
   - function ENVs injection

## Deploy function

Copy the `resources/secrets/deepai-template.env` into `resources/secrets/deepai.env` and fill in `API-KEY`.

Then, deploy the application via
```
make deploy
```
## Test

```bash
  curl \
    -F 'text=Teddy bear' \
    https://function-text2img.{KYMA_RUNTIME_DOMAIN}
```