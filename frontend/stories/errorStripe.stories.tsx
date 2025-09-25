import React from 'react'
import { StoryFn, Meta } from '@storybook/nextjs'
import ErrorStripe, { ErrorStripeProps } from '../components/lib/errorStripe'
import {
  CodeGCPArtifactRegistry,
  ParamGroupEmail,
} from '../lib/rest/generatedDto'

export default {
  title: 'Components/ErrorStripe',
  component: ErrorStripe,
} as Meta<typeof ErrorStripe>

export const Default = {
  args: {
    error: {
      message: 'An error occurred',
    },
  },
}

export const WithRequestID = {
  args: {
    error: {
      message: 'An error occurred',
      requestID: '12345',
    },
  },
}

export const WithCodeAndParam = {
  args: {
    error: {
      message:
        'gapi: Error 403: Permission denied on "projects/navikt-internal/locations/europe-west1/repositories/nada-images/packages/*".',
      code: CodeGCPArtifactRegistry,
      statusCode: 403,
      param: ParamGroupEmail,
      requestId: '12345abc',
    },
  },
}

export const NoError = {
  args: {
    error: null,
  },
}
