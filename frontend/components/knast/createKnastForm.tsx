import { Button, Link, Loader, Popover, Table } from "@navikt/ds-react";
import React, { useState } from "react";
import { ColorInfoText } from "./designTokens";
import { ChevronDownDoubleIcon, HandFingerIcon, QuestionmarkCircleIcon } from "@navikt/aksel-icons";
import ContainerImageSelector from "./widgets/containerImageSelector";
import MachineTypeSelector from "./widgets/machineTypeSelector";
import useMachineTypeSelector from "./widgets/machineTypeSelector";
import { useAutoCloseAlert } from "./widgets/autoCloseAlert";
import { useWorkstationJobs } from "./queries";
import { JobStateRunning, WorkstationJob } from "../../lib/rest/generatedDto";

export interface CreateKnastFormProps {
    options?: any;
}

const CreateKnastForm = ({ options }: CreateKnastFormProps) => {
    const knastIntroRef = React.useRef<HTMLDivElement | null>(null);
    const envIntroRef = React.useRef<HTMLDivElement | null>(null);
    const machinetypeIntroRef = React.useRef<HTMLDivElement | null>(null);
    const [showKnastIntroBar, setShowKnastIntroBar] = React.useState(false);
    const [showEnvIntro, setShowEnvIntro] = React.useState(false);
    const [showMachineTypeIntro, setShowMachineTypeIntro] = React.useState(false);
    const [selectedContainerImage, setSelectedContainerImage] = useState<string>(options?.containerImages?.find((image: any) => image !== undefined)?.image || "");
    const { MachineTypeSelector, selectedMachineType } = useMachineTypeSelector(options?.machineTypes?.find((type: any) => type !== undefined)?.machineType || "", options)
    const { showAlert, AutoHideAlert } = useAutoCloseAlert(5000);
    const [error, setError] = useState<string | null>(null);
    const [saving, setSaving] = useState(false);
    const workstationJobs = useWorkstationJobs();
    const hasRunningJobs = !!workstationJobs.data?.jobs?.filter((job): job is WorkstationJob =>
        job !== undefined && job.state === JobStateRunning).length || false;

    const handleCreate = async () => {
        try {
            // Call API to create knast with selectedContainerImage and selectedMachineType
            // await createKnastAPI(selectedContainerImage, selectedMachineType);
            showAlert();
        } catch (e: any) {
            setError(e.message || "Noe gikk galt ved oppretting av Knast.");
        }
    }

    return (<div>
        <div className="flex flex-row mt-4 items-center">
            <Popover placement="top" content={"url format"} anchorEl={knastIntroRef.current} open={showKnastIntroBar} onClose={() => setShowKnastIntroBar(false)}>
                <p className="p-2 text-sm">Knast er ditt sikre og brukervennlige utviklingsmiljø i skyen, som er basert på Google Cloud Workstation og støtter tilgang til on-prem data.</p>
            </Popover>
            <p>Vi finner ikke din</p> <p className="font-bold ml-1">Knast</p>
            <div className="ml-0" ref={knastIntroRef}
                onMouseEnter={() => setShowKnastIntroBar(true)}
                onMouseLeave={() => setShowKnastIntroBar(false)}>
                <QuestionmarkCircleIcon color={ColorInfoText} width={20} height={20} /></div>
            <p>. Det er enkelt å lage en</p>
            <ChevronDownDoubleIcon className="ml-1 animate-bounce" width={22} height={22} />
            {/* Add form fields and logic for creating a new knast */}
        </div>
        <div className="w-180 border-blue-100 border rounded mt-10 p-6">
            <div className="grid grid-cols-[20%_70%] gap-6">
                <div className="flex flex-row items-center mb-2">
                    <Popover placement="top" content={"url format"} anchorEl={envIntroRef.current} open={showEnvIntro} onClose={() => setShowEnvIntro(false)}>
                        <p className="p-2 text-sm">Utviklingsmiljø er programvaren, verktøyene og bibliotekene du ønsker.</p>
                    </Popover>
                    <p>Utviklingsmiljø</p>
                    <div className="ml-0" ref={envIntroRef}
                        onMouseEnter={() => setShowEnvIntro(true)}
                        onMouseLeave={() => setShowEnvIntro(false)}>
                        <QuestionmarkCircleIcon color={ColorInfoText} width={20} height={20} /></div>
                </div>
                <ContainerImageSelector initialContainerImage={selectedContainerImage} handleSetContainerImage={setSelectedContainerImage} />

                <div className="flex flex-row items-center mb-2">
                    <Popover placement="top" content={"url format"} anchorEl={machinetypeIntroRef.current} open={showMachineTypeIntro} onClose={() => setShowMachineTypeIntro(false)}>
                        <p className="p-2 text-sm">Maskintype handler om hvor mye minne og prosessorkraft du trenger.</p>
                    </Popover>
                    <p>Maskintype</p>
                    <div className="ml-0" ref={machinetypeIntroRef}
                        onMouseEnter={() => setShowMachineTypeIntro(true)}
                        onMouseLeave={() => setShowMachineTypeIntro(false)}>
                        <QuestionmarkCircleIcon color={ColorInfoText} width={20} height={20} /></div>
                </div>
                <MachineTypeSelector />
            </div>
            <div className="mt-10">
            <div className="flex mt-2 flex-rol gap-4">
                <Button variant="primary" disabled={!selectedContainerImage || !selectedMachineType} onClick={handleCreate}>Opprett{(saving || hasRunningJobs) && <Loader className="ml-2" />}</Button>
            </div>
            {error && <div className="mt-4 text-red-600">{error}</div>}
            <AutoHideAlert variant="success" className="mt-4">Endringer lagret.</AutoHideAlert>
            </div>
        </div>
    </div>
    )
}

export default CreateKnastForm;