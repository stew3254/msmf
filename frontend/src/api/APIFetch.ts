interface FallibleResponse {
    error?: string
}

function fetchAPI(method: string, body: object, path: string[]) {
    const url = "/api/" + path.join("/");
    return fetch(url, {method: method, body: JSON.stringify(body)})
        .then(response => response.json())
        .catch(error => Promise.reject(error.message))
        .then((json: FallibleResponse) => {
            if (json.error) {
                return Promise.reject(json.error);
            } else {
                return json
            }
        })
}

export function postAPI(body: object, ...path: string[]) {
    return fetchAPI("POST", body, path);
}
