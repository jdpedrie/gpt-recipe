import mimetypes

import requests
from requests_toolbelt import MultipartEncoder


class TandoorAPI:

    def __init__(self, url, token, **kwargs):
        self.token = token
        self.url = url
        self.headers = {'Content-Type': 'application/json', 'Authorization': f'Bearer {self.token}'}
        self.session = requests.Session()

    def get_recipe_from_json(self, json):
        url = self.url + "recipe-from-source/"
        response = self.session.post(url, json=json, headers=self.headers)
        if response.status_code != 200:
            return False, response.json()
        return True, response.json()

    def create_recipe(self, recipe):
        url = self.url + "recipe/"
        response = self.session.post(url, json=recipe, headers=self.headers)
        if response.status_code != 201:
            if type(response) is dict:
                return False, response.json()
            else:
                return False, response
        return True, response.json()

    def create_file(self, name, image):
        url = self.url + "user-file/"
        headers = self.headers.copy()

        m = MultipartEncoder(fields={'name': name, 'file': (image[0], image[1], mimetypes.guess_type(image[0])[0])})
        headers['Content-Type'] = m.content_type
        response = self.session.post(url, data=m, headers=headers)
        if response.status_code != 201:
            return False, response.json()
        return True, response.json()

    def create_food(self, recipe):
        url = self.url + "food/"
        food = {'name': recipe['name'].lower(), 'recipe': {'id': recipe['id'], 'name': recipe['name']}}
        response = self.session.post(url, json=food, headers=self.headers)
        if response.status_code != 201:
            return False, response.json()
        return True, response.json()
