import {
    Heading,
    Loader,
    Tabs,
} from "@navikt/ds-react"
import {
    WorkstationJob,
    WorkstationJobStateRunning,
} from "../../lib/rest/generatedDto";
import WorkstationJobsState from "./jobs";
import {CaptionsIcon, CogRotationIcon, GlobeIcon, LaptopIcon} from "@navikt/aksel-icons";
import {useState} from "react";
import {useWorkstationExists, useWorkstationJobs} from "./queries";
import WorkstationStatus from "./WorkstationStatus";
import WorkstationAdministrate from "./WorkstationAdministrate";
import WorkstationZonalTagBindings from "./WorkstationZonalBindings";
import WorkstationLogState from "./WorkstationLogState";
import WorkstationPythonSetup from "./WorkstationPythonSetup";
import WorkstationSetupPage from "./WorkstationSetupPage";

export const Workstation = () => {
    const workstationExists = useWorkstationExists()
    const workstationJobs = useWorkstationJobs()

    const [unreadJobsCounter, setUnreadJobsCounter] = useState(0);
    const [startedGuide, setStartedGuide] = useState(false)
    const [activeTab, setActiveTab] = useState("administrer");

    const incrementUnreadJobsCounter = () => {
        setUnreadJobsCounter(prevCounter => prevCounter + 1);
    };

    const haveRunningJob: boolean = (workstationJobs.data?.jobs?.filter((job):
    job is WorkstationJob => job !== undefined && job.state === WorkstationJobStateRunning).length || 0) > 0;

    const shouldShowSetupPage = workstationExists.data === false && !haveRunningJob

    if (workstationExists.isLoading || workstationJobs.isLoading) {
        return <Loader size="large" title="Laster.."/>
    }

    if (workstationExists.isError || workstationJobs.isError) {
        return <div>Det skjedde en feil under lasting av data :(</div>
    }

    if (shouldShowSetupPage) {
        return <WorkstationSetupPage setStartedGuide={setStartedGuide} startedGuide={startedGuide}/>
    }

    return (
        <div className="flex flex-col gap-4">
            <div/>
            <div>
                Her kan du gjøre endringer på din personlige Knast
            </div>
            <div className="flex">
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
                            value="logger"
                            label="URL håndtering"
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
