// UrlListInput.stories.tsx
import React from 'react';
import { ComponentMeta, ComponentStory } from '@storybook/react';
import UrlListInput from '../components/workstation/formElements/urlListInput';
import { WorkstationStateProvider } from '../components/workstation/WorkstationStateProvider';

export default {
    title: 'Components/UrlListInput',
    component: UrlListInput,
} as ComponentMeta<typeof UrlListInput>;

const Template: ComponentStory<typeof UrlListInput> = (args) => (
    <WorkstationStateProvider initialState={args}>
        <UrlListInput {...args} />
    </WorkstationStateProvider>
);

export const Default = Template.bind({});
Default.args = {
    workstation: {
        urlAllowList: ["https://example.com", "https://another-example.com"],
    },
};

export const Empty = Template.bind({});
Empty.args = {
    workstation: {
        urlAllowList: [],
    },
};

export const Undefined = Template.bind({});
Undefined.args = {
    workstation: {
    },
};
