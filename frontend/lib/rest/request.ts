export interface HttpError {
    kind?: string
    statusCode?: number
    code?: string
    param?: string
    message: string
    requestId?: string
}

const makeError = async (res: Response): Promise<HttpError> => {
    const body = await res.json()
    return {
        kind: body.error.kind,
        statusCode: body.error.statusCode || res.status,
        code: body.error.code,
        param: body.error.param,
        message: body.error.message || 'An error occurred',
        requestId: body.error.requestId || res.headers.get('X-Request-Id'),
    }
}
export const apiTemplate = (url: string, method: string, body?: any) => fetch(url, {
    method: method,
    credentials: 'include',
    headers: body ? (body instanceof FormData ? undefined : {'Content-Type': 'application/json'}) : undefined,
    body: body ? (body instanceof FormData ? body : JSON.stringify(body)) : undefined,
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

export const deleteTemplate = curriedApiTemplate('DELETE')()
