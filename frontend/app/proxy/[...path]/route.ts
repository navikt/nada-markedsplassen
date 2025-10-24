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
	const headers = request.headers
	const authorization = headers.get('authorization')

	if (!authorization) {
		return NextResponse.json(
			{ error: 'Unauthenticated' },
			{ status: 401 }
		)
	}

	const oboToken = await exchangeToken(authorization)
	if (!oboToken) {
		return NextResponse.json(
			{ error: 'Internal server error' },
			{ status: 500 }
		)
	}
	
	headers.set('authorization', oboToken)

	const path = request.nextUrl.pathname.replace('/proxy/', '')
	const searchParams = request.nextUrl.searchParams.toString()
	const url = `${BACKEND_URL}/${path}${searchParams ? `?${searchParams}` : ''}`

	try {
		const requestBody = await getBody(request)
		const response = await fetch(url, {
			method: request.method,
			headers: headers,
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

async function getBody(request: NextRequest) {
	if (request.method !== 'GET' && request.method !== 'HEAD') {
		if (isContentTypeJSON(request.headers)) {
			return JSON.stringify(await request.json())
		}
		return await request.text()
	}
}


async function exchangeToken(token: string) {
	const scope = process.env.NADA_BACKEND_SCOPE
	const tokenExchangeEndpoint = process.env.NAIS_TOKEN_EXCHANGE_ENDPOINT
	if (!scope || !tokenExchangeEndpoint) {
		console.error(
			`Missing required environments variable: NADA_BACKEND_SCOPE=${scope}, NAIS_TOKEN_EXCHANGE_ENDPOINT=${tokenExchangeEndpoint}`
		)
		return undefined
	}

	const body = {
		identity_provider: 'azuread',
		target: scope,
		user_token: token.replace('Bearer ', '')
	}

	const response = await fetch(tokenExchangeEndpoint, {
		method: 'POST',
		headers: {
			'content-type': 'application/json',
		},
		body: JSON.stringify(body)
	})
	const jsonResponse = await response.json()

	if (response.status !== 200) {
		console.error(`Failed to exchange token: ${jsonResponse}`)
		return undefined
	}

	return `${jsonResponse['token_type']} ${jsonResponse['access_token']}`
}
