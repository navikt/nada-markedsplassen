// WorkstationInputForm.stories.tsx
import React from 'react';
import { ComponentMeta, ComponentStory } from '@storybook/react';
import WorkstationInputForm from '../components/workstation/form';
import { WorkstationStateProvider } from '../components/workstation/WorkstationStateProvider';
import {WorkstationMachineType} from "../lib/rest/generatedDto";

export default {
    title: 'Components/WorkstationInputForm',
    component: WorkstationInputForm,
} as ComponentMeta<typeof WorkstationInputForm>;

const Template: ComponentStory<typeof WorkstationInputForm> = (args) => (
    <WorkstationStateProvider initialState={args}>
        <WorkstationInputForm {...args} />
    </WorkstationStateProvider>
);

export const Default = Template.bind({});
Default.args = {
    workstation: {
        config: {
            machineType: 'n1-standard-1',
            image: 'image1',
            firewallRulesAllowList: ["tag1", "tag2"],
            urlAllowList: ["https://example.com", "https://another-example.com"],
            disableGlobalURLAllowList: false,
        },
    },
    workstationOptions: {
        containerImages: [
            { image: 'image1', description: 'Description 1', labels: { 'org.opencontainers.image.title': 'Title 1' } },
            { image: 'image2', description: 'Description 2', labels: { 'org.opencontainers.image.title': 'Title 2' } },
        ],
        firewallTags: [
            { name: "tag1" },
            { name: "tag2" },
            { name: "tag3" },
        ],
        machineTypes: [
            { machineType: 'n1-standard-1', memoryGB: 3.75, vCPU: 1 } as WorkstationMachineType,
            { machineType: 'n1-standard-2', memoryGB: 7.5, vCPU: 2 } as WorkstationMachineType,
        ],
    },
    workstationJobs: [],
    workstationLogs: [],
    workstationStartJobs: [],
    workstationZonalTagBindingJobs: [],
    effectiveTags: [],
    refetchWorkstationJobs: () => alert('Refetch Workstation Jobs'),
    incrementUnreadJobsCounter: () => alert('Increment Unread Jobs Counter'),
};
