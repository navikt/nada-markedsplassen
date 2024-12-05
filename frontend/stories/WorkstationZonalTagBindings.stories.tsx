import React from 'react';
import { Meta, Story } from '@storybook/react';
import WorkstationZonalTagBindings, { WorkstationZonalTagBindingsProps } from '../components/workstation/WorkstationZonalBindings';
import {
    EffectiveTag,
    WorkstationZonalTagBindingJob,
    WorkstationZonalTagBindingJobActionAdd,
    WorkstationZonalTagBindingJobActionRemove
} from '../lib/rest/generatedDto';
import {
    WorkstationJobStateCompleted,
    WorkstationJobStateFailed,
    WorkstationJobStateRunning
} from "../lib/rest/generatedDto";

export default {
    title: 'Components/WorkstationZonalTagBindings',
    component: WorkstationZonalTagBindings,
} as Meta;

const Template: Story<WorkstationZonalTagBindingsProps> = (args) => <WorkstationZonalTagBindings {...args} />;

export const Default = Template.bind({});
Default.args = {
    workstationIsRunning: true,
    expectedTags: [
        "something",
        "adeo.no",
        "vg.no",
        "dagbladet.no",
        "chess.no"
    ] as string[],
    effectiveTags: [
        {
            tagValue: "tagValues/281479855594020",
            tagKey: "tagKeys/23092989382",
            namespacedTagKey: "project/something/something",
            namespacedTagValue: "project/something/something",
            tagKeyParentName: "project/092314034"
        },
        {
            tagValue: "tagValues/281479855594020",
            tagKey: "tagKeys/23092989382",
            namespacedTagKey: "project/something/adeo.no",
            namespacedTagValue: "project/something/adeo.no",
            tagKeyParentName: "project/092314034"
        },
        {
            tagValue: "tagValues/281479855594020",
            tagKey: "tagKeys/23092989382",
            namespacedTagKey: "project/something/dagbladet.no",
            namespacedTagValue: "project/something/bsky.app",
            tagKeyParentName: "project/092314034"
        }
    ] as EffectiveTag[],
    jobs: [
        {
            state: WorkstationJobStateRunning,
            id: 1,
            action: WorkstationZonalTagBindingJobActionAdd,
            duplicate: false,
            zone: "europe-west1-b",
            tagNamespacedName: "project/something/dagbladet.no"
        },
        {
            state: WorkstationJobStateCompleted,
            id: 2,
            action: WorkstationZonalTagBindingJobActionAdd,
            duplicate: false,
            zone: "europe-west1-b",
            tagNamespacedName: "project/something/something"
        },
        {
            state: WorkstationJobStateFailed,
            id: 3,
            action: WorkstationZonalTagBindingJobActionRemove,
            duplicate: false,
            errors: "a datacenter exploded",
            zone: "europe-west1-b",
            tagValue: "tagValues/281479855594020",
            tagNamespacedName: "project/something/adeo.no"
        },
        {
            state: WorkstationJobStateRunning,
            id: 4,
            action: WorkstationZonalTagBindingJobActionRemove,
            duplicate: false,
            zone: "europe-west1-b",
            tagNamespacedName: "project/something/bsky.app"
        },
        {
            state: WorkstationJobStateFailed,
            id: 5,
            action: WorkstationZonalTagBindingJobActionRemove,
            duplicate: false,
            zone: "europe-west1-b",
            errors: "something really bad went wrong",
            tagValue: "tagValues/281479855594020",
            tagNamespacedName: "project/something/chess.no"
        },
    ] as WorkstationZonalTagBindingJob[],
};

export const Empty = Template.bind({});
Empty.args = {
    jobs: [],
};
