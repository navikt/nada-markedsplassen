import { NextRequest, NextResponse } from "next/server";

const BACKEND_URL = process.env.NEXT_PUBLIC_ENV === 'development' ?
	'http://localhost:8080/api' : 'http://nada-backend/api'

export async function GET(request: NextRequest) {
	return handleRequest(request)
}

export async function POST(request: NextRequest) {
	return handleRequest(request)
}

export async function PUT(request: NextRequest) {
	return handleRequest(request)
}

export async function DELETE(request: NextRequest) {
	return handleRequest(request)
}

export async function PATCH(request: NextRequest) {
	return handleRequest(request)
}

async function handleRequest(request: NextRequest) {
	const authorization = request.headers.get('authorization')
	console.log(request.headers)


	if (!authorization) {
		// TODO: Redirect?
		return NextResponse.json(
			{ error: 'Unauthorized' },
			{ status: 401 },
		)
	}

	const path = request.nextUrl.pathname.replace('/api/proxy/', '')
	const searchParams = request.nextUrl.searchParams.toString()
	const url = `${BACKEND_URL}/${path}${searchParams ? `?${searchParams}` : ''}`

	try {
		const requestBody = await body(request)
		const response = await fetch(url, {
			method: request.method,
			// TODO: Send med alle headers?
			headers: {
				'authorization': authorization,
				'content-type': request.headers.get('content-type') || 'application/json',
			},
			body: requestBody,
		})

		const data = await responseData(response)
		return new NextResponse(JSON.stringify(data), {
			status: response.status,
			headers: response.headers,
		})
	} catch (error) {
		console.error('Proxy error:', error)
		return NextResponse.json(
			{ error: 'Failed to fetch from backend' },
			{ status: 500 }
		)
	}

}

async function responseData(response: Response) {
	if (isContentTypeJSON(response.headers)) {
		return await response.json()
	}
	return await response.text()
}

function isContentTypeJSON(headers: Headers) {
	return headers.get('content-type')?.includes('application/json')
}

async function body(request: NextRequest) {
	if (request.method !== 'GET' && request.method !== 'HEAD') {
		if (isContentTypeJSON(request.headers)) {
			return JSON.stringify(await request.json())
		}
		return await request.text()
	}
}
