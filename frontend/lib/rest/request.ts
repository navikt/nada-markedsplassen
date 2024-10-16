const makeError = async (res: Response) => {
    const body = await res.text()
    return {
      message: body,
      status: res.status,
      id: res.headers.get('X-Correlation-Id'),
    }
  }

export const apiTemplate = (url: string, method: string, body?: any) => fetch(url, {
    method: method,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  }).then(async res => {
    if (!res.ok) {
      throw await makeError(res)
    }
    //if response is 2xx, the response body should be json object
    return res.json()
  })
  
const curriedApiTemplate = (method: string) => (body?: string) => (url:string) => apiTemplate(url, method, body)

export const fetchTemplate = curriedApiTemplate('GET')()
  
export const postTemplate = (url: string, body?: any) => curriedApiTemplate('POST')(body)(url)

export const putTemplate = (url: string, body?: any) => curriedApiTemplate('PUT')(body)(url)
  
export const deleteTemplate = (url: string) => curriedApiTemplate('DELETE')()(url)