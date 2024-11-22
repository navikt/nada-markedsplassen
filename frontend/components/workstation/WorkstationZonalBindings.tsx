import React, {useEffect, useState} from 'react';
import {
    WorkstationZonalTagBindingJob,
    WorkstationJobStateRunning, WorkstationJobStateFailed, WorkstationZonalTagBindingJobActionRemove, Workstation_STATE_RUNNING,
} from '../../lib/rest/generatedDto';
import {Heading, Button, Loader, HStack, List} from '@navikt/ds-react';
import {
    CheckmarkCircleIcon,
    XMarkOctagonIcon,
    CogRotationIcon,
    ArrowUpIcon, ArrowDownIcon
} from '@navikt/aksel-icons';
import {FaceCryIcon} from '@navikt/aksel-icons';
import {createWorkstationZonalTagBindingJob} from "../../lib/rest/workstation";
import { useWorkstation } from './WorkstationStateProvider';

const WorkstationZonalTagBindings= ({
                                                                                 }) => {
    const {workstation, workstationZonalTagBindingJobs, effectiveTags } = useWorkstation()

    const workstationIsRunning = workstation?.state === Workstation_STATE_RUNNING;
    const expectedTags = workstation?.config?.firewallRulesAllowList;
    const jobs = workstationZonalTagBindingJobs;

    const runningJobs = jobs?.filter(job => job?.state === WorkstationJobStateRunning);
    const failedJobs = jobs?.filter(job => job?.state === WorkstationJobStateFailed);

    const [showButton, setShowButton] = useState(false);

    useEffect(() => {
        const hasFailedOrInactiveTags = expectedTags?.some(tag => {
            const isEffective = effectiveTags?.some(eTag => eTag?.namespacedTagValue?.split('/').pop() === tag);
            const hasRunningJob = runningJobs?.some(job => job?.tagNamespacedName.split('/').pop() === tag);
            const hasFailedJob = failedJobs?.some(job => job?.tagNamespacedName.split('/').pop() === tag);
            return !isEffective && (!hasRunningJob || hasFailedJob);
        });

        setShowButton(!!hasFailedOrInactiveTags);
    }, [expectedTags, effectiveTags, runningJobs, failedJobs]);

    async function handleCreateZonalTagBindingJobs() {
        try {
            await createWorkstationZonalTagBindingJob();
            setShowButton(false);
        } catch (error) {
            console.error('Failed to create zonal tag binding job:', error);
        }
    }

    const renderStatus = (tag: string) => {
        const isEffective = effectiveTags?.some(eTag => eTag?.namespacedTagValue?.split('/').pop() === tag);
        const hasRunningJob = runningJobs?.some(job => job.tagNamespacedName.split('/').pop() === tag);
        const hasFailedJob = failedJobs?.some(job => job.tagNamespacedName.split('/').pop() === tag);

        if (isEffective) {
            return (
                <List.Item icon={<CheckmarkCircleIcon/>} key={tag}>
                    {tag}
                </List.Item>
            );
        }

        if (hasRunningJob) {
            return (
                <List.Item icon={<Loader size="small"/>} key={tag}>
                    <HStack gap="1">Oppretter åpning mot <strong>{tag}</strong> <ArrowUpIcon/></HStack>
                </List.Item>
            );
        }

        if (hasFailedJob) {
            const failedJobForTag = failedJobs?.find(job => job.tagNamespacedName.split('/').pop() === tag);

            return (
                <List.Item icon={<XMarkOctagonIcon/>} key={tag}>
                    Kobling mot <strong>{tag}</strong> feilet: {failedJobForTag?.errors.join(', ').substring(0, 50) ?? 'ukjent feil'}
                </List.Item>
            );
        }

        return (
            <List.Item icon={<FaceCryIcon/>} key={tag}>
                Kobling mot <strong>{tag}</strong> er ikke aktiv, og vi vet ikke hvorfor det feilet
            </List.Item>
        );
    };

    function renderRemove(job: WorkstationZonalTagBindingJob, index: number) {
        return (
            <List.Item icon={<Loader size="small"/>} key={index}>
                <HStack gap="1">Fjerner åpning
                    mot <strong>{job.tagNamespacedName.split('/').pop()}</strong><ArrowDownIcon/></HStack>
            </List.Item>
        );
    }

    if (!workstationIsRunning) return <div></div>;

    return (
        <div>
            <Heading level="2" size="small">Brannmur åpninger</Heading>
            <List>
                {runningJobs?.filter(job => job.action === WorkstationZonalTagBindingJobActionRemove).map(renderRemove)}
                {expectedTags?.map(renderStatus)}
                {(!runningJobs?.length && !expectedTags?.length) && (
                    <List.Item>
                        Ingen åpninger
                    </List.Item>
                )}
            </List>
            {showButton &&
                <Button icon={<CogRotationIcon/>} onClick={handleCreateZonalTagBindingJobs}>Prøv å opprette koblinger på
                    nytt</Button>}
        </div>
    );
};

export default WorkstationZonalTagBindings;
