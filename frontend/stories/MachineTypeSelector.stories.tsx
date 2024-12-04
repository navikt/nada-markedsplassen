import React from 'react';
import { ComponentMeta, ComponentStory } from '@storybook/react';
import { MachineTypeSelector } from '../components/workstation/formElements/machineTypeSelector';
import { WorkstationStateProvider } from '../components/workstation/WorkstationStateProvider';
import {WorkstationMachineType, WorkstationState} from "../lib/rest/generatedDto";

export default {
    title: 'Components/MachineTypeSelector',
    component: MachineTypeSelector,
} as ComponentMeta<typeof MachineTypeSelector>;

const Template: ComponentStory<typeof MachineTypeSelector> = (args: WorkstationState) => (
    <WorkstationStateProvider initialState={args}>
        <MachineTypeSelector />
    </WorkstationStateProvider>
);

export const Default = Template.bind({});
Default.args = {
    workstation: {
        config: {
            machineType: 'type1',
        },
    },
    workstationOptions: {
        machineTypes: [
            { machineType: 'some-kind-of-type', vCPU: 2, memoryGB: 4 } as WorkstationMachineType,
            { machineType: 'some-other-kind-of-type', vCPU: 4, memoryGB: 8 } as WorkstationMachineType,
        ],
    },
};

export const NoMachineType = Template.bind({});
NoMachineType.args = {
    workstation: {
        config: {
            machineType: '',
        },
    },
    workstationOptions: {
        machineTypes: [
            { machineType: 'some-kind-of-type', vCPU: 2, memoryGB: 4 } as WorkstationMachineType,
            { machineType: 'some-other-kind-of-type', vCPU: 4, memoryGB: 8 } as WorkstationMachineType,
        ],
    },
};

export const NoMachineTypeNoOptions = Template.bind({});
NoMachineType.args = {
    workstation: {
        config: {
            machineType: '',
        },
    },
    workstationOptions: {
        machineTypes: [],
    },
};
