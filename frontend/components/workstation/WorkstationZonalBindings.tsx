import React, {useEffect, useState} from 'react';
import {
    WorkstationZonalTagBindingJob,
    WorkstationJobStateRunning,
    WorkstationJobStateFailed,
    WorkstationZonalTagBindingJobActionRemove,
    Workstation_STATE_RUNNING,
} from '../../lib/rest/generatedDto';
import {Heading, Button, Loader, HStack, List, Table} from '@navikt/ds-react';
import {
    CheckmarkCircleIcon,
    XMarkOctagonIcon,
    CogRotationIcon,
    ArrowUpIcon, ArrowDownIcon
} from '@navikt/aksel-icons';
import {FaceCryIcon} from '@navikt/aksel-icons';
import {
    useCreateZonalTagBindingJob,
    useWorkstationEffectiveTags,
    useWorkstationMine,
    useWorkstationZonalTagBindingJobs
} from "../knast/queries";

const WorkstationZonalTagBindings = ({}) => {
    const workstation = useWorkstationMine()
    const bindingJobs = useWorkstationZonalTagBindingJobs()
    const effectiveTags = useWorkstationEffectiveTags()
    const createZonalTagBindingJob = useCreateZonalTagBindingJob()

    const workstationIsRunning = workstation.data?.state === Workstation_STATE_RUNNING;
    const expectedTags = workstation.data?.config?.firewallRulesAllowList;

    const runningJobs = bindingJobs.data?.jobs?.filter(job => job != undefined && job.state === WorkstationJobStateRunning) ?? [];
    const failedJobs = bindingJobs.data?.jobs?.filter(job => job != undefined && job.state === WorkstationJobStateFailed) ?? [];

    function handleCreateZonalTagBindingJobs() {
        try {
            createZonalTagBindingJob.mutate()
        } catch (error) {
            console.error('Failed to create zonal tag binding job:', error);
        }
    }

    const renderStatus = (tag: string) => {
        const isEffective = effectiveTags.data?.tags?.some(eTag => eTag?.namespacedTagValue?.split('/').pop() === tag);
        const hasRunningJob = runningJobs?.some(job => job?.tagNamespacedName.split('/').pop() === tag);
        const hasFailedJob = failedJobs?.some(job => job?.tagNamespacedName.split('/').pop() === tag);

        if (isEffective) {
            return (
                <Table.Row key={tag}>
                    <Table.HeaderCell scope="row">
                        {tag}
                    </Table.HeaderCell>
                    <Table.DataCell>
                        <HStack gap="1">
                            Aktiv
                        </HStack>
                    </Table.DataCell>
                    <Table.DataCell>
                        <CheckmarkCircleIcon/>
                    </Table.DataCell>
                </Table.Row>
            );
        }

        if (hasRunningJob) {
            return (
                <Table.Row key={tag}>
                    <Table.HeaderCell scope="row">
                        {tag}
                    </Table.HeaderCell>
                    <Table.DataCell>
                        <HStack gap="1">
                            Åpner <ArrowUpIcon/>
                        </HStack>
                    </Table.DataCell>
                    <Table.DataCell>
                        <Loader size="small"/>
                    </Table.DataCell>
                </Table.Row>
            );
        }

        if (hasFailedJob) {
            const failedJobForTag = failedJobs?.find(job => job?.tagNamespacedName.split('/').pop() === tag);

            return (
                <Table.Row key={tag}>
                    <Table.HeaderCell scope="row">{tag}</Table.HeaderCell>
                    <Table.DataCell>
                        <HStack gap="1">
                            Oppkobling feilet: {failedJobForTag?.errors.join(', ').substring(0, 50) ?? 'ukjent feil'} <XMarkOctagonIcon/>
                        </HStack>
                    </Table.DataCell>
                    <Table.DataCell>
                        <Button onClick={handleCreateZonalTagBindingJobs}>Prøv igjen</Button>
                    </Table.DataCell>
                </Table.Row>
            );
        }

        return (
            <Loader size="small"/>
        );
    };

    function renderRemove(job: WorkstationZonalTagBindingJob) {
        const tag = job.tagNamespacedName.split('/').pop()
        return (
            <Table.Row key={tag}>
                <Table.HeaderCell scope="row">{tag}</Table.HeaderCell>
                <Table.DataCell>
                    <HStack gap="1">
                        Fjerner åpning<ArrowDownIcon/>
                    </HStack>
                </Table.DataCell>
                <Table.DataCell><Loader size="small"/></Table.DataCell>
            </Table.Row>
        )
    }

    if (!workstationIsRunning) {
        return <div></div>;
    }
    if (runningJobs?.length === 0 && expectedTags?.length === 0) {
        return <div>Ingen åpninger</div>;
    }

    return (
        <>
            <Heading className="pt-8" level="2" size="medium">Brannmur åpninger</Heading>
            <Table size="small">
                <Table.Header>
                    <Table.Row>
                        <Table.HeaderCell scope="col">Åpning</Table.HeaderCell>
                        <Table.HeaderCell scope="col">Status</Table.HeaderCell>
                        <Table.HeaderCell scope="col"></Table.HeaderCell>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    {runningJobs?.filter((job): job is WorkstationZonalTagBindingJob => job?.action === WorkstationZonalTagBindingJobActionRemove).map(renderRemove)}
                    {expectedTags?.map(renderStatus)}
                </Table.Body>
            </Table>
        </>
    );
};

export default WorkstationZonalTagBindings;
