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
    const workstationJobs= useWorkstationJobs()

    const [unreadJobsCounter, setUnreadJobsCounter] = useState(0);
    const [activeTab, setActiveTab] = useState("administrer");

    const incrementUnreadJobsCounter = () => {
        setUnreadJobsCounter(prevCounter => prevCounter + 1);
    };

    const haveRunningJob: boolean = (workstationJobs.data?.jobs?.filter((job):
    job is WorkstationJob => job !== undefined && job.state === WorkstationJobStateRunning).length ?? 0) > 0;

    if (workstationExists.isLoading || workstationJobs.isLoading) {
        return <Loader size="large" title="Laster.."/>
    }

    if (workstationExists.data === false && !haveRunningJob) {
        return <WorkstationSetupPage/>
    }

    return (
        <div className="flex flex-col gap-8">
            <p>Her kan du opprette og gjøre endringer på din personlige Knast</p>
            <div className="flex">
                <div className="flex flex-col">
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
                            value="endringer"
                            label={unreadJobsCounter > 0 ? `Endringslogg (${unreadJobsCounter})` : "Endringslogg"}
                            icon={haveRunningJob ? <Loader size="small"/> : <CogRotationIcon aria-hidden/>}
                            onClick={() => setUnreadJobsCounter(0)}
                        />
                        <Tabs.Tab
                            value="logger"
                            label="Blokkerte URLer"
                            icon={<CaptionsIcon aria-hidden/>}
                        />
                        <Tabs.Tab
                            value="python"
                            label="Oppsett av Python"
                            icon={<GlobeIcon aria-hidden/>}
                        />
                    </Tabs.List>
                    <Tabs.Panel value="administrer" className="p-4">
                        <WorkstationAdministrate incrementUnreadJobsCounter={incrementUnreadJobsCounter}/>
                    </Tabs.Panel>
                    <Tabs.Panel value="endringer" className="p-4">
                        <WorkstationJobsState/>
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
    )
}

export default Workstation;
