import { CaptionsIcon, CogRotationIcon, GlobeIcon, LaptopIcon } from '@navikt/aksel-icons';
import {
    Heading,
    Loader,
    Tabs,
} from "@navikt/ds-react";
import { useEffect, useState } from "react";
import {
    JobStateRunning,
    Workstation_STATE_RUNNING,
    WorkstationJob,
    WorkstationResyncJob,
} from '../../lib/rest/generatedDto'
import { useWorkstationExists, useWorkstationJobs, useWorkstationMine, useWorkstationResyncJobs } from './queries';
import WorkstationAdministrate from "./WorkstationAdministrate";
import WorkstationPythonSetup from "./WorkstationPythonSetup";
import WorkstationSetupPage from "./WorkstationSetupPage";
import WorkstationStatus from "./WorkstationStatus";
import FirewallTagSelector from './formElements/firewallTagSelector';
import WorkstationLogState from './WorkstationLogState';
import WorkstationConnectivity from './WorkstationConnectivity'

export const Workstation = () => {
    const workstation = useWorkstationMine()
    const workstationExists = useWorkstationExists()
    const workstationJobs = useWorkstationJobs()
    const workstationResyncJobs = useWorkstationResyncJobs()

    const [startedGuide, setStartedGuide] = useState(false)
    const [activeTab, setActiveTab] = useState("internal_services");

    const workstationIsRunning = workstation.data?.state === Workstation_STATE_RUNNING;

    const haveRunningJob: boolean = (workstationJobs.data?.jobs?.filter((job):
    job is WorkstationJob => job !== undefined && job.state === JobStateRunning).length || 0) > 0;

    const haveRunningResyncJob: boolean = (workstationResyncJobs.data?.jobs?.filter((job):
    job is WorkstationResyncJob => job !== undefined && job.state === JobStateRunning).length || 0) > 0;

    const workstationsIsUpdating = haveRunningJob || haveRunningResyncJob;

    useEffect(() => {
        workstationExists.refetch()
    }, [haveRunningJob])
    
    if (workstationExists.isLoading || workstationJobs.isLoading) {
        return <Loader size="large" title="Laster.."/>
    }

    if (workstationExists.isError || workstationJobs.isError) {
        return <div>Det skjedde en feil under lasting av data :(</div>
    }

    const shouldShowSetupPage = workstationExists.data === false && !workstationsIsUpdating
    const shouldShowCreatingPage = workstationExists.data === false && workstationsIsUpdating

    if (shouldShowSetupPage) {
        return <WorkstationSetupPage setStartedGuide={setStartedGuide} startedGuide={startedGuide}/>
    }

    if (shouldShowCreatingPage) {
        return <div>Oppretter din Knast... <Loader/></div>
    }

    return (
        <div className="flex flex-col gap-4 w-full">
                    Her kan du gjøre endringer på din personlige Knast.
                <div className="flex flex-wrap gap-2">
                    <a
                        href="https://docs.knada.io/analyse/knast/generelt/"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-blue-600 hover:text-blue-800 underline"
                    >📖 Dokumentasjon
                    </a>
                    <a
                        href="https://docs.knada.io/analyse/knast/kom-i-gang/#kjr-knast-som-en-pwa"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-blue-600 hover:text-blue-800 underline"
                    >📱Installere Knast som en PWA
                    </a>
            </div>
            <div className="flex flex-col gap-4">
                <div>
                    <Heading level="1" size="medium">Status</Heading>
                    <div className="mt-4">
                        <WorkstationStatus hasRunningJob={workstationsIsUpdating}/>
                    </div>
                </div>
                <div className="flex flex-row gap-4">
                    <div className="flex flex-col">
                        <Tabs value={activeTab} onChange={setActiveTab}>
                            <Tabs.List>
                                <Tabs.Tab
                                    value="internal_services"
                                    label="Interne tjenester"
                                    icon={<CogRotationIcon aria-hidden/>}
                                />
                                <Tabs.Tab
                                    value="administrer"
                                    label="Administrer"
                                    icon={<LaptopIcon aria-hidden/>}
                                />
                                <Tabs.Tab
                                    value="logger"
                                    label="Internettåpninger"
                                    icon={<CaptionsIcon aria-hidden/>}
                                />
                                <Tabs.Tab
                                    value="python"
                                    label="Oppsett av Python"
                                    icon={<GlobeIcon aria-hidden/>}
                                />
                            </Tabs.List>
                            <Tabs.Panel value="internal_services" className="p-4">
                                <div className="flex flex-col gap-4">
                                    <WorkstationConnectivity />
                                    <FirewallTagSelector enabled={workstationIsRunning}/>
                                </div>
                            </Tabs.Panel>
                            <Tabs.Panel value="administrer" className="p-4">
                                <WorkstationAdministrate/>
                            </Tabs.Panel>
                            <Tabs.Panel value="logger" className="p-4">
                                <WorkstationLogState/>
                            </Tabs.Panel>
                            <Tabs.Panel value="python" className="p-4">
                                <WorkstationPythonSetup/>
                            </Tabs.Panel>
                        </Tabs>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default Workstation;
