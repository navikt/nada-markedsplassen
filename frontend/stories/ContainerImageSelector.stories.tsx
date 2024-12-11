// ContainerImageSelector.stories.tsx
import React from 'react';
import { ComponentMeta, ComponentStory } from '@storybook/react';
import ContainerImageSelector from '../components/workstation/formElements/containerImageSelector';
import { WorkstationStateProvider } from '../components/workstation/WorkstationStateProvider';
import { WorkstationContainer } from '../lib/rest/generatedDto';

export default {
    title: 'Components/ContainerImageSelector',
    component: ContainerImageSelector,
} as ComponentMeta<typeof ContainerImageSelector>;

const Template: ComponentStory<typeof ContainerImageSelector> = (args) => (
    <WorkstationStateProvider initialState={args}>
        <ContainerImageSelector {...args} />
    </WorkstationStateProvider>
);

export const Default = Template.bind({});
Default.args = {
    workstation: {
        config: {
            image: 'image1',
        },
    },
    workstationOptions: {
        containerImages: [
            {
                image: 'image1',
                description: 'Description 1',
                labels: {
                    'org.opencontainers.image.title': 'Super cool data science image',
                    'org.opencontainers.image.description': 'This is a super cool data science image',
                    'org.opencontainers.image.source': 'https://example.com',
                },
                documentation: "# Documentation\n\nThis is the documentation for image 1",
            } as WorkstationContainer,
            {
                image: 'image2',
                description: 'Description 2',
                labels: {
                    'org.opencontainers.image.title': 'A pretty good image',
                    'org.opencontainers.image.description': 'This is a pretty good image',
                    'org.opencontainers.image.source': 'https://example.com',
                },
                documentation: "# Documentation\n\nThis is the documentation for image 2",
            } as WorkstationContainer,
        ],
    },
    onDocumentationLinkClick: () => alert('Documentation link clicked'),
};

export const NoLabels = Template.bind({});
NoLabels.args = {
    workstation: {
        config: {
            image: 'image1',
        },
    },
    workstationOptions: {
        containerImages: [
            {
                image: 'image1',
                description: 'Description 1',
                labels: {},
                documentation: "# Documentation\n\nThis is the documentation for image 1",
            } as WorkstationContainer,
            {
                image: 'image2',
                description: 'Description 2',
                labels: {},
                documentation: "# Documentation\n\nThis is the documentation for image 2",
            } as WorkstationContainer,
        ],
    },
    onDocumentationLinkClick: () => alert('Documentation link clicked'),
};
