// FirewallTagSelector.stories.tsx
import React from 'react';
import { ComponentMeta, ComponentStory } from '@storybook/react';
import FirewallTagSelector from '../components/workstation/formElements/firewallTagSelector';
import { WorkstationStateProvider } from '../components/workstation/WorkstationStateProvider';

export default {
    title: 'Components/FirewallTagSelector',
    component: FirewallTagSelector,
} as ComponentMeta<typeof FirewallTagSelector>;

const Template: ComponentStory<typeof FirewallTagSelector> = (args) => (
    <WorkstationStateProvider initialState={args}>
        <FirewallTagSelector {...args} />
    </WorkstationStateProvider>
);

export const Default = Template.bind({});
Default.args = {
    workstation: {
        config: {
            firewallRulesAllowList: ["tag1", "tag2"],
        },
    },
    workstationOptions: {
        firewallTags: [
            { name: "tag1" },
            { name: "tag2" },
            { name: "tag3" },
        ],
    },
};
