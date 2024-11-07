import React from 'react';
import { ComponentStory, ComponentMeta } from '@storybook/react';
import ErrorStripe, { ErrorStripeProps } from '../components/lib/errorStripe';
import {CodeGCPArtifactRegistry, ParamGroupEmail} from "../lib/rest/generatedDto";

export default {
    title: 'Components/ErrorStripe',
    component: ErrorStripe,
} as ComponentMeta<typeof ErrorStripe>;

const Template: ComponentStory<typeof ErrorStripe> = (args) => <ErrorStripe {...args} />;

export const Default = Template.bind({});
Default.args = {
    error: {
        message: 'An error occurred',
    },
};

export const WithRequestID = Template.bind({});
WithRequestID.args = {
    error: {
        message: 'An error occurred',
        requestID: '12345',
    },
};

export const WithCodeAndParam = Template.bind({});
WithCodeAndParam.args = {
    error: {
        message: 'gapi: Error 403: Permission denied on "projects/navikt-internal/locations/europe-west1/repositories/nada-images/packages/*".',
        code: CodeGCPArtifactRegistry,
        statusCode: 403,
        param: ParamGroupEmail,
        requestId: '12345abc',
    },
};

export const NoError = Template.bind({});
NoError.args = {
    error: null,
};
