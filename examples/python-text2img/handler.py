import requests
import os
from bottle import HTTPResponse

def main(event, context):
    text = event['extensions']['request'].params.get('text')

    if not text:
        return HTTPResponse(body={'"text" parameter is mandatory'}, status=400)

    url = os.getenv('URL')
    apiKey = os.getenv('API-KEY')
    print(f'Trying to query {url} with "{text}"')
    # Send request to the external service
    response = requests.post(url, files={'text': text}, headers={'api-key': apiKey})
    return response.json()
    
