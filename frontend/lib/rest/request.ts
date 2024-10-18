export interface HttpError{
    message: string
    status: number
    id?: string | null
}

const makeError = async (res: Response): Promise<HttpError> => {
    const body = await res.text()
    return {
        message: body || 'An error occurred',
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
    const contentType = res.headers.get('Content-Type')
    
    if(!contentType || !contentType.includes('application/json')){
        console.log("Invalid response with content type: ", contentType)
        return null
    }

    const resText = await res.text()

    return resText.length ? JSON.parse(resText) : null
})

const curriedApiTemplate = (method: string) => (body?: string) => (url: string) => apiTemplate(url, method, body)

export const fetchTemplate = curriedApiTemplate('GET')()

export const postTemplate = (url: string, body?: any) => curriedApiTemplate('POST')(body)(url)

export const putTemplate = (url: string, body?: any) => curriedApiTemplate('PUT')(body)(url)

export const deleteTemplate = (url: string) => curriedApiTemplate('DELETE')()(url)