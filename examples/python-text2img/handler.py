import requests
import os
from bottle import HTTPResponse

def main(event, context):

    text = event["extensions"]["request"].params.get("text")

    if text:
        url = os.getenv('URL')
        apiKey = os.getenv('API-KEY')

        print('Trying to query {} with "{}"'.format(url, text))

        # Send request to the external service
        response = requests.post(url, files=dict(text=text), headers={"api-key": apiKey})
    
        return response.json()

    else:
        return HTTPResponse(body={"'text' parameter is mandatory"}, status=400)

    
