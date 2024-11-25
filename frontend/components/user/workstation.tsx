import {
    Loader,
    Tabs,
} from "@navikt/ds-react"
import LoaderSpinner from "../lib/spinner"
import {
    Workstation_STATE_RUNNING, WorkstationJob,
    WorkstationJobStateRunning,
} from "../../lib/rest/generatedDto";
import {
    useConditionalEffectiveTags,
    useConditionalWorkstationLogs, useConditionalWorkstationZonalTagBindingJobs,
    useGetWorkstation,
    useGetWorkstationJobs,
    useGetWorkstationOptions, useGetWorkstationStartJobs
} from "../../lib/rest/workstation";
import WorkstationLogState from "../workstation/logState";
import WorkstationJobsState from "../workstation/jobs";
import {CaptionsIcon, CogRotationIcon, GlobeIcon, LaptopIcon} from "@navikt/aksel-icons";
import PythonSetup from "../workstation/pythonSetup";
import WorkstationInputForm from "../workstation/form";
import WorkstationStatus from "../workstation/status";
import {useEffect, useState} from "react";
import {useWorkstation, useWorkstationDispatch} from "../workstation/WorkstationStateProvider";

const WorkstationContainer = () => {
    const {
        workstationJobs,
    } = useWorkstation()

    const [unreadJobsCounter, setUnreadJobsCounter] = useState(0);
    const [activeTab, setActiveTab] = useState("administrer");

    const incrementUnreadJobsCounter = () => {
        setUnreadJobsCounter(prevCounter => prevCounter + 1);
    };

    const haveRunningJob: boolean = (workstationJobs?.filter((job): job is WorkstationJob => job !== undefined && job.state === WorkstationJobStateRunning).length ?? 0) > 0;

    return (
        <div className="flex flex-col gap-8">
            <p>Her kan du opprette og gjøre endringer på din personlige Knast</p>
            <div className="flex">
                <div className="flex flex-col">
                    <WorkstationStatus/>
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
                        <WorkstationInputForm incrementUnreadJobsCounter={incrementUnreadJobsCounter}/>
                    </Tabs.Panel>
                    <Tabs.Panel value="endringer" className="p-4">
                        <WorkstationJobsState/>
                    </Tabs.Panel>
                    <Tabs.Panel value="logger" className="p-4">
                        <WorkstationLogState/>
                    </Tabs.Panel>
                    <Tabs.Panel value="python" className="p-4">
                        <PythonSetup/>
                    </Tabs.Panel>
                </Tabs>
            </div>
        </div>
    )
}

export const Workstation: React.FC = () => {
    const {data: workstation, isLoading: loading} = useGetWorkstation();
    const {data: workstationOptions, isLoading: loadingOptions} = useGetWorkstationOptions();
    const {data: workstationJobs, isLoading: loadingJobs} = useGetWorkstationJobs();
    const isRunning = workstation?.state === Workstation_STATE_RUNNING;
    const {data: workstationLogs} = useConditionalWorkstationLogs(isRunning);
    const {data: workstationZonalTagBindingJobs} = useConditionalWorkstationZonalTagBindingJobs(isRunning);
    const {data: effectiveTags} = useConditionalEffectiveTags(isRunning);
    const {data: workstationStartJobs} = useGetWorkstationStartJobs();

    const dispatch = useWorkstationDispatch();

    useEffect(() => {
        if (workstation) dispatch({type: 'SET_WORKSTATION', payload: workstation});
        if (workstationOptions) dispatch({type: 'SET_WORKSTATION_OPTIONS', payload: workstationOptions});
        if (workstationLogs) dispatch({type: 'SET_WORKSTATION_LOGS', payload: workstationLogs ?? []});
        if (workstationJobs) dispatch({type: 'SET_WORKSTATION_JOBS', payload: workstationJobs ?? []});
        if (workstationStartJobs) dispatch({type: 'SET_WORKSTATION_START_JOBS', payload: workstationStartJobs ?? []});
        if (workstationZonalTagBindingJobs) dispatch({
            type: 'SET_WORKSTATION_ZONAL_TAG_BINDING_JOBS',
            payload: workstationZonalTagBindingJobs ?? []
        });
        if (effectiveTags) dispatch({type: 'SET_EFFECTIVE_TAGS', payload: effectiveTags ?? []});
    }, [workstation, workstationOptions, workstationLogs, workstationJobs, workstationStartJobs, workstationZonalTagBindingJobs, effectiveTags]);

    if (loading || loadingOptions || loadingJobs) return <LoaderSpinner/>;

    return <WorkstationContainer/>;
};
