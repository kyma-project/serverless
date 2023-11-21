# Sample Functions

Functions support multiple languages through the use of runtimes. To use a chosen runtime, add its name and version as a value in the **spec.runtime** field of the [Function custom resource (CR)](../resources/06-10-function-cr.md). If this value is not specified, it defaults to `nodejs18`. Dependencies for a Node.js Function must be specified using the [`package.json`](https://docs.npmjs.com/creating-a-package-json-file) file format. Dependencies for a Python Function must follow the format used by [pip](https://packaging.python.org/key_projects/#pip).

>**TIP:** If you are interested in the Function's signature, `event` and `context` objects, and custom HTTP responses the Function returns, read about [Function’s specification](07-70-function-specification.md).

See sample Functions for all available runtimes:

<!-- tabs:start -->

#### **Node.js**

```bash
cat <<EOF | kubectl apply -f -
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: test-function-nodejs
spec:
  runtime: nodejs18
  source:
    inline:
      dependencies: |
        {
          "name": "test-function-nodejs",
          "version": "1.0.0",
          "dependencies": {
            "lodash":"^4.17.20"
          }
        }
      source: |
        const _ = require('lodash')
        module.exports = {
          main: function(event, context) {
            return _.kebabCase('Hello World from Node.js 16 Function');
          }
        }
EOF
```

#### **Python 3.9**

```bash
cat <<EOF | kubectl apply -f -
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: test-function-python39
spec:
  runtime: python39
  source:
    inline:
      dependencies: |
        requests==2.24.0
      source: |
        import requests
        def main(event, context):
            r = requests.get('https://swapi.dev/api/people/13')
            return r.json()
EOF
```

<!-- tabs:end -->