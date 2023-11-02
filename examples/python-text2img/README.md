# Overview

The Function uses [Deep AI API](https://deepai.org/machine-learning-model/text2img) for image generation from the text.
Its purpose is to demonstrate how to achieve basic development tasks as a Python Function developer, such as:

 - How to consume a request
 - How to customize a response
 - How to read ENVs
 - How to configure libraries as dependencies
 - How to send requests from the Function code
 - How to configure Function in the Kyma context
   - function API exposure
   - function ENVs injection

## Deploy Function

1. Copy `resources/secrets/deepai-template.env` into `resources/secrets/deepai.env` and fill in `API-KEY`.

2. Deploy the application.
```
make deploy
```
## Test

```bash
  curl \
    -F 'text=Teddy bear' \
    https://function-text2img.{KYMA_RUNTIME_DOMAIN}
```