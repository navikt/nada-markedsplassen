import React from 'react';
import { Meta, Story } from '@storybook/react';
import DiffViewerComponent, { DiffViewerProps } from '../components/workstation/DiffViewerComponent';
import { Diff } from '../../lib/rest/generatedDto';
import {
    WorkstationDiffContainerImage,
    WorkstationDiffDisableGlobalURLAllowList,
    WorkstationDiffMachineType, WorkstationDiffOnPremAllowList, WorkstationDiffURLAllowList
} from "../lib/rest/generatedDto";

export default {
    title: 'Components/DiffViewerComponent',
    component: DiffViewerComponent,
} as Meta;

const Template: Story<DiffViewerProps> = (args) => <DiffViewerComponent {...args} />;

export const Default = Template.bind({});
Default.args = {
    diff: {
        [WorkstationDiffContainerImage]: {
            added: ['New Image 1'],
            removed: ['Old Image 1'],
        },
        [WorkstationDiffDisableGlobalURLAllowList]: {
            added: ['true'],
            removed: ['false'],
        },
        [WorkstationDiffURLAllowList]: {
            added: ['vg.no'],
            removed: ['dagbladet.no', 'aftenposten.no'],
        },
        [WorkstationDiffMachineType]: {
            added: ['n1-standard-1'],
            removed: ['n1-standard-2'],
        },
        [WorkstationDiffOnPremAllowList]: {
            added: ['onprem1', 'onprem-super.adeo.no'],
            removed: ['onprem2'],
        }
    } as Record<string, Diff>,
};

export const Empty = Template.bind({});
Empty.args = {
    diff: {},
};
