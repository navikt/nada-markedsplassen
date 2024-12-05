// GlobalAllowUrlListInput.stories.tsx
import React from 'react';
import { ComponentMeta, ComponentStory } from '@storybook/react';
import GlobalAllowUrlListInput from '../components/workstation/formElements/globalAllowURLListInput';
import { WorkstationStateProvider } from '../components/workstation/WorkstationStateProvider';

export default {
    title: 'Components/GlobalAllowUrlListInput',
    component: GlobalAllowUrlListInput,
} as ComponentMeta<typeof GlobalAllowUrlListInput>;

const Template: ComponentStory<typeof GlobalAllowUrlListInput> = (args) => (
    <WorkstationStateProvider initialState={args}>
        <GlobalAllowUrlListInput />
    </WorkstationStateProvider>
);

export const Default = Template.bind({});
Default.args = {
    workstation: {
        config: {
            disableGlobalURLAllowList: false,
        },
    },
    workstationOptions: {
        globalURLAllowList: ["https://example.com", "https://another-example.com"],
    },
};

export const Empty = Template.bind({});
Empty.args = {
    workstation: {
        config: {
            disableGlobalURLAllowList: false,
        },
    },
    workstationOptions: {
        globalURLAllowList: [],
    },
};

export const Error = Template.bind({});
Error.args = {
    workstation: {
        config: {
            disableGlobalURLAllowList: false,
        },
    },
    workstationOptions: {
        globalURLAllowList: undefined,
    },
};

export const Disabled = Template.bind({});
Disabled.args = {
    workstation: {
        config: {
            disableGlobalURLAllowList: true,
        },
    },
    workstationOptions: {
        globalURLAllowList: ["https://example.com", "https://another-example.com"],
    },
};
