import { CaptionsIcon, CogRotationIcon, GlobeIcon, LaptopIcon, RouterIcon } from '@navikt/aksel-icons'
import {
    Heading,
    Loader,
    Tabs,
} from "@navikt/ds-react";
import { useState } from "react";
import {
    Workstation_STATE_RUNNING,
    WorkstationJob,
    WorkstationJobStateRunning,
} from '../../lib/rest/generatedDto'
import WorkstationJobsState from "./jobs";
import { useWorkstationExists, useWorkstationJobs, useWorkstationMine } from './queries'
import WorkstationAdministrate from "./WorkstationAdministrate";
import WorkstationPythonSetup from "./WorkstationPythonSetup";
import WorkstationSetupPage from "./WorkstationSetupPage";
import WorkstationStatus from "./WorkstationStatus";
import WorkstationZonalTagBindings from "./WorkstationZonalBindings";
import FirewallTagSelector from './formElements/firewallTagSelector'
import WorkstationLogState from './WorkstationLogState'

export const Workstation = () => {
    const workstation = useWorkstationMine()
    const workstationExists = useWorkstationExists()
    const workstationJobs = useWorkstationJobs()

    const [unreadJobsCounter, setUnreadJobsCounter] = useState(0);
    const [startedGuide, setStartedGuide] = useState(false)
    const [activeTab, setActiveTab] = useState("administrer");

    const incrementUnreadJobsCounter = () => {
        setUnreadJobsCounter(prevCounter => prevCounter + 1);
    };

    const workstationIsRunning = workstation.data?.state === Workstation_STATE_RUNNING;

    const haveRunningJob: boolean = (workstationJobs.data?.jobs?.filter((job):
    job is WorkstationJob => job !== undefined && job.state === WorkstationJobStateRunning).length || 0) > 0;


    if (workstationExists.isLoading || workstationJobs.isLoading) {
        return <Loader size="large" title="Laster.."/>
    }

    if (workstationExists.isError || workstationJobs.isError) {
        return <div>Det skjedde en feil under lasting av data :(</div>
    }

    const shouldShowSetupPage = workstationExists.data === false && !haveRunningJob
    const shouldShowCreatingPage = workstationExists.data === false && haveRunningJob

    if (shouldShowSetupPage) {
        return <WorkstationSetupPage setStartedGuide={setStartedGuide} startedGuide={startedGuide}/>
    }

    if (shouldShowCreatingPage) {
        return <div>Oppretter din Knast... <Loader/></div>
    }

    return (
        <div className="flex flex-col gap-4 w-full">
            <div/>
            <div>
                Her kan du gjøre endringer på din personlige Knast
            </div>
            <div className="flex flex-row gap-4">
                <div className="flex flex-col gap-4">
                    <Heading level="1" size="medium">Status</Heading>
                    <WorkstationStatus/>
                    <WorkstationZonalTagBindings/>
                </div>
            </div>
            <div className="flex flex-col">
                <Tabs value={activeTab} onChange={setActiveTab}>
                    <Tabs.List>
                        <Tabs.Tab
                            value="administrer"
                            label="Administrer"
                            icon={<LaptopIcon aria-hidden/>}
                        />
                        <Tabs.Tab
                          value="onprem"
                          label="Endre nettverk"
                          icon={<RouterIcon aria-hidden/>}
                        />
                        <Tabs.Tab
                            value="logger"
                            label="Endre internettåpninger"
                            icon={<CaptionsIcon aria-hidden/>}
                        />
                        <Tabs.Tab
                            value="python"
                            label="Oppsett av Python"
                            icon={<GlobeIcon aria-hidden/>}
                        />
                        <Tabs.Tab
                            value="endringer"
                            label={unreadJobsCounter > 0 ? `Endringslogg (${unreadJobsCounter})` : "Endringslogg"}
                            icon={haveRunningJob ? <Loader size="small"/> : <CogRotationIcon aria-hidden/>}
                            onClick={() => setUnreadJobsCounter(0)}
                        />
                    </Tabs.List>
                    <Tabs.Panel value="administrer" className="p-4">
                        <WorkstationAdministrate incrementUnreadJobsCounter={incrementUnreadJobsCounter}/>
                    </Tabs.Panel>
                    <Tabs.Panel value="onprem" className="p-4">
                        <FirewallTagSelector enabled={workstationIsRunning}/>
                    </Tabs.Panel>
                    <Tabs.Panel value="logger" className="p-4">
                        <WorkstationLogState/>
                    </Tabs.Panel>
                    <Tabs.Panel value="python" className="p-4">
                        <WorkstationPythonSetup/>
                    </Tabs.Panel>
                    <Tabs.Panel value="endringer" className="p-4">
                        <WorkstationJobsState/>
                    </Tabs.Panel>
                </Tabs>
            </div>
        </div>
    )
}

export default Workstation;
