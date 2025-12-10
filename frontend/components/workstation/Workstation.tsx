import { CaptionsIcon, CogRotationIcon, GlobeIcon, LaptopIcon } from '@navikt/aksel-icons';
import {
    Button,
    Heading,
    Loader,
    Modal,
    Tabs,
} from "@navikt/ds-react";
import { useEffect, useState } from "react";
import {
    JobStateRunning,
    OnpremHostTypeTNS,
    Workstation_STATE_RUNNING,
    WorkstationJob,
    WorkstationResyncJob,
} from '../../lib/rest/generatedDto';
import FirewallTagSelector from './formElements/firewallTagSelector';
import { useConfigWorkstationSSHJobs, useWorkstationExists, useWorkstationJobs, useWorkstationMine, useWorkstationOnpremMapping, useWorkstationResyncJobs } from './queries';
import WorkstationAdministrate from "./WorkstationAdministrate";
import WorkstationConnectivity from './WorkstationConnectivity';
import WorkstationLogState from './WorkstationLogState';
import WorkstationPythonSetup from "./WorkstationPythonSetup";
import WorkstationSetupPage from "./WorkstationSetupPage";
import WorkstationStatus from "./WorkstationStatus";
import { configWorkstationSSH } from '../../lib/rest/workstation';
import { useOnpremMapping } from '../onpremmapping/queries';

export const Workstation = () => {
    const workstation = useWorkstationMine()
    const workstationExists = useWorkstationExists()
    const workstationJobs = useWorkstationJobs()
    const workstationResyncJobs = useWorkstationResyncJobs()
    const sshJob = useConfigWorkstationSSHJobs()

    const [startedGuide, setStartedGuide] = useState(false)
    const [activeTab, setActiveTab] = useState("internal_services");
    const [loggerSubTab, setLoggerSubTab] = useState<string | undefined>(undefined);
    const [openConnectivityAlert, setOpenConnectivityAlert] = useState(false);
    const workstationOnpremMapping = useWorkstationOnpremMapping()
    const onpremMapping = useOnpremMapping()

    const workstationIsRunning = workstation.data?.state === Workstation_STATE_RUNNING;

    // allowSSH is false when workstation data is not loaded, potentially issue?
    const allowSSH = !!workstation.data?.allowSSH;

    const updatingSSH = !!sshJob.data

    const haveRunningJob: boolean = (workstationJobs.data?.jobs?.filter((job):
        job is WorkstationJob => job !== undefined && job.state === JobStateRunning).length || 0) > 0;

    const haveRunningResyncJob: boolean = (workstationResyncJobs.data?.jobs?.filter((job):
        job is WorkstationResyncJob => job !== undefined && job.state === JobStateRunning).length || 0) > 0;

    const workstationsIsUpdating = haveRunningJob || haveRunningResyncJob || updatingSSH;

    const handleSetActiveTab = (tab: string, subTab?: string) => {
        setActiveTab(tab)
        if (tab === "logger" && subTab) {
            setLoggerSubTab(subTab)
        } else {
            setLoggerSubTab(undefined)
        }
    }

    const onSSHSwitch = (checked: boolean) => {
        if (checked && workstationOnpremMapping.data?.hosts?.some(
            h => Object.entries(onpremMapping.data?.hosts ?? {}).find(([type, _]) => type === OnpremHostTypeTNS)?.[1].some(host => host?.Host === h))) {
            setOpenConnectivityAlert(true);
            return;
        }
        configWorkstationSSH(checked);
    }

    useEffect(() => {
        workstationExists.refetch()
    }, [haveRunningJob, workstationExists])

    if (workstationExists.isLoading || workstationJobs.isLoading) {
        return <Loader size="large" title="Laster.." />
    }

    if (workstationExists.isError || workstationJobs.isError) {
        return <div>Det skjedde en feil under lasting av data :(</div>
    }

    const shouldShowSetupPage = workstationExists.data === false && !workstationsIsUpdating
    const shouldShowCreatingPage = workstationExists.data === false && workstationsIsUpdating

    if (shouldShowSetupPage) {
        return <WorkstationSetupPage setStartedGuide={setStartedGuide} startedGuide={startedGuide} />
    }

    if (shouldShowCreatingPage) {
        return <div>Oppretter din Knast... <Loader /></div>
    }

    return (
        <div className="flex flex-col gap-4 w-full">
            Her kan du gjøre endringer på din personlige Knast.
            <Modal open={openConnectivityAlert} onClose={() => setOpenConnectivityAlert(false)}
                aria-label="">
                <Modal.Body>
                    Av sikkerhetshensyn kan ikke Knast åpne DVH-kilder når SSH (lokal IDE-tilgang) er aktivert. Du vil ikke kunne koble til DVH-kilden hvis du fortsetter.
                </Modal.Body>
                <Modal.Footer>
                    <Button type="button" onClick={() => {
                        configWorkstationSSH(true)
                        setOpenConnectivityAlert(false)
                    }}>
                        Fortsett
                    </Button>
                    <Button type="button" onClick={() => setOpenConnectivityAlert(false)} variant='secondary'>
                        Avbryt
                    </Button>
                </Modal.Footer>
            </Modal>
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
                        <WorkstationStatus hasRunningJob={workstationsIsUpdating} setActiveTab={handleSetActiveTab} toggleSSHSwitch={onSSHSwitch} />
                    </div>
                </div>
                <div className="flex flex-row gap-4">
                    <div className="flex flex-col max-w-3xl">
                        <Tabs value={activeTab} onChange={setActiveTab}>
                            <Tabs.List>
                                <Tabs.Tab
                                    value="internal_services"
                                    label="Interne tjenester"
                                    icon={<CogRotationIcon aria-hidden />}
                                />
                                <Tabs.Tab
                                    value="administrer"
                                    label="Administrer"
                                    icon={<LaptopIcon aria-hidden />}
                                />
                                <Tabs.Tab
                                    value="logger"
                                    label="Internettåpninger"
                                    icon={<CaptionsIcon aria-hidden />}
                                />
                                <Tabs.Tab
                                    value="python"
                                    label="Oppsett av Python"
                                    icon={<GlobeIcon aria-hidden />}
                                />
                            </Tabs.List>
                            <Tabs.Panel value="internal_services" className="p-4">
                                <div className="flex flex-col gap-4">
                                    <WorkstationConnectivity />
                                    <FirewallTagSelector enabled={workstationIsRunning} allowSSH={allowSSH} updatingSSH={updatingSSH} />
                                </div>
                            </Tabs.Panel>
                            <Tabs.Panel value="administrer" className="p-4">
                                <WorkstationAdministrate />
                            </Tabs.Panel>
                            <Tabs.Panel value="logger" className="p-4">
                                <WorkstationLogState initialTab={loggerSubTab} />
                            </Tabs.Panel>
                            <Tabs.Panel value="python" className="p-4">
                                <WorkstationPythonSetup />
                            </Tabs.Panel>
                        </Tabs>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default Workstation;
