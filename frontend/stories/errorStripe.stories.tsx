import React from 'react';
import { ComponentStory, ComponentMeta } from '@storybook/react';
import ErrorStripe, { ErrorStripeProps } from '../components/lib/errorStripe';

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
        message: 'An error occurred',
        code: 'INVALID_REQUEST',
        param: 'dataset',
    },
};

export const NoError = Template.bind({});
NoError.args = {
    error: null,
};
