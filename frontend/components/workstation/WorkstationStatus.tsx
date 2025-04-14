import {
    Workstation_STATE_RUNNING,
    Workstation_STATE_STARTING,
    Workstation_STATE_STOPPED,
    Workstation_STATE_STOPPING,
} from "../../lib/rest/generatedDto";
import { Alert, Button, BodyLong, Modal, Loader, CopyButton, List, Link, Popover } from "@navikt/ds-react";
import { InformationSquareFillIcon, PlayIcon, RocketIcon, StopIcon } from "@navikt/aksel-icons";
import {
    useStartWorkstation,
    useStopWorkstation,
    useWorkstationMine,
} from "./queries";
import { useRef, useState } from "react";
import { NaisdeviceGreen } from "../lib/icons/naisdeviceGreen";
import { buildUrl } from "../../lib/rest/apiUrl";

interface WorkstationStatusProps {
    hasRunningJob: boolean;
}

const NaisdevicePopoverContent = () => {
    return (
        <Popover.Content>
            <div className="flex flex-row items-end">
                <InformationSquareFillIcon color="#66a5f4" width="1.5em" height="1.5em" />
                Må sørge for at naisdevice er &nbsp;
                <p className="text-[#78c010]">grønn</p>
                <NaisdeviceGreen />
            </div>
        </Popover.Content>
    )
}

const WorkstationStatus = ({ hasRunningJob }: WorkstationStatusProps) => {
    const [showNaisdeviceInfo, setShowNaisdeviceInfo] = useState(false);
    const [currentRef, setCurrentRef] = useState<HTMLButtonElement | null>(null);
    const button1Ref = useRef<HTMLButtonElement>(null);
    const button2Ref = useRef<HTMLButtonElement>(null);

    const workstation = useWorkstationMine()

    const startWorkstation = useStartWorkstation()
    const stopWorkstation = useStopWorkstation()

    const modalRef = useRef<HTMLDialogElement>(null);

    const handleOnStart = () => {
        startWorkstation.mutate()
    };

    const handleOnStop = () => {
        stopWorkstation.mutate()
    };

    const handleOpenWorkstationWindow = () => {
        if (workstation.data?.host) {

            window.open(buildUrl("googleOauth2")("login")({redirect: `https://${workstation.data?.host}/`}), "_blank");
        }
    };

    const startStopButtons = (startButtonDisabled: boolean, stopButtonDisabled: boolean) => {
        return (
            <>
                <Button style={{ backgroundColor: 'var(--a-green-500)' }} disabled={startButtonDisabled} onMouseOver={() => setShowNaisdeviceInfo(true)} onClick={handleOnStart}>
                    <div className="flex"><PlayIcon title="Start din Knast" fontSize="1.5rem" />Start</div>
                </Button>
                <Button style={{ backgroundColor: 'var(--a-red-500)' }} ref={button2Ref} disabled={stopButtonDisabled} onClick={handleOnStop}>
                    <div className="flex"><StopIcon title="Stopp din Knast" fontSize="1.5rem" />Stopp</div>
                </Button>
            </>
        )
    }

    if (!workstation) {
        return (
            <div className="flex flex-col gap-4 pt-4">
                <Alert variant={'warning'}>Du har ikke opprettet en Knast</Alert>
                <div className="flex gap-2">
                    {startStopButtons(true, true)}
                </div>
            </div>
        )
    }

    switch (workstation.data?.state) {
        case Workstation_STATE_RUNNING:
            return (
                <div className="flex gap-2">
                    {startStopButtons(true, false)}
                    <Button ref={button1Ref} onMouseOver={() => {
                        setCurrentRef(button1Ref.current)
                        setShowNaisdeviceInfo(true)
                    }}
                        onMouseLeave={() => setShowNaisdeviceInfo(false)}
                        onClick={handleOpenWorkstationWindow}>
                        <div className="flex"><RocketIcon title="a11y-title" fontSize="1.5rem" />Åpne din Knast i nytt
                            vindu
                        </div>
                    </Button>
                    <Button ref={button2Ref} onMouseOver={() => {
                        setCurrentRef(button2Ref.current)
                        setShowNaisdeviceInfo(true)
                    }}
                        onMouseLeave={() => setShowNaisdeviceInfo(false)} onClick={() => modalRef?.current?.showModal()}>Bruk Knast via lokal IDE</Button>
                    <Popover
                        open={showNaisdeviceInfo}
                        onClose={() => setShowNaisdeviceInfo(false)}
                        anchorEl={currentRef}
                    >
                        <NaisdevicePopoverContent />
                    </Popover>

                    <Modal width="medium" ref={modalRef} header={{ heading: "Bruk av Knast via lokal IDE" }} closeOnBackdropClick>
                        <Modal.Body>
                            <BodyLong>
                                <List as="ol" title="Følgende må gjøres på lokal maskin for å koble til Knast:">
                                    <List.Item title={"Logg inn i Google Cloud (kjøres lokalt)"}>
                                        <div className="flex">
                                            <code
                                                className="rounded-xs bg-surface-neutral-subtle font-mono text-sm font-semibold">gcloud
                                                auth login</code><CopyButton size="xsmall" copyText="gcloud auth login" />
                                        </div>
                                    </List.Item>
                                    <List.Item title={"Opprette SSH-tunnel (kjøres lokalt)"}>
                                        <div className="flex">
                                            <code
                                                className="rounded-xs bg-surface-neutral-subtle px-1 py-05 font-mono text-sm font-semibold">
                                                gcloud workstations start-tcp-tunnel --cluster=knada --config={workstation.data.slug} --region=europe-north1 --project knada-gcp --local-host-port=:33649 {workstation.data.slug} 22</code>
                                            <CopyButton size="xsmall" copyText={`gcloud workstations start-tcp-tunnel --cluster=knada --config=${workstation.data.slug} --region=europe-north1 --project knada-gcp --local-host-port=:33649 ${workstation.data.slug} 22`} />
                                        </div>
                                    </List.Item>
                                    <List.Item title={"Opprette SSH-nøkkel (kjøres lokalt, hopp over om du allerede har gjort dette)"}>
                                        <em>Sett et passord på SSH-nøkkelen. Du vil aldri bli bedt om å bytte dette.</em>
                                        <div className="flex">
                                            <code className="rounded-xs bg-surface-neutral-subtle font-mono text-sm font-semibold">{`ssh-keygen -t ed25519 -C "din_epost_email@nav.no"`}</code><CopyButton size="xsmall" copyText={`ssh-keygen -t ed25519 -C "din_epost_email@nav.no"`}></CopyButton>
                                        </div>
                                    </List.Item>
                                    <List.Item title={"Få Knast til å stole på din SSH-nøkkel"}>
                                        <div className="bg-red-50"><b> NB! Dette steget utføres <em>på Knasten</em>. Trykk "Åpne din Knast i nytt vindu"</b></div>
                                        <ul>
                                            <li>Opprette directory <strong>~/.ssh/</strong> hvis det ikke allerede finnes</li>
                                            <li>Opprett filen <strong>authorized_keys</strong> i <strong>~/.ssh/</strong></li>
                                            <li>Lime inn innholdet fra public-delen av SSH-nøkkelen fra <strong>.ssh/id_ed25519.pub</strong> eller tilsvarende på lokal maskin inn i <strong>authorized_keys</strong> på Knast</li>
                                        </ul>

                                    </List.Item>
                                    <List.Item title={"Legg til knast i ssh-configen (kjøres lokalt, hopp over om du allerede har gjort dette)"}>
                                        <div className="flex">
                                            <code
                                                className="rounded-xs bg-surface-neutral-subtle font-mono text-sm font-semibold">{`echo -e "\\nHost knast\\n\\tHostName localhost\\n\\tPort 33649\\n\\tUser user\\n\\tUserKnownHostsFile /dev/null\\n\\tStrictHostKeyChecking no">>~/.ssh/config`}</code><CopyButton
                                                    size="xsmall"
                                                    copyText={`echo -e "\\nHost knast\\n\\tHostName localhost\\n\\tPort 33649\\n\\tUser user\\n\\tUserKnownHostsFile /dev/null\\n\\tStrictHostKeyChecking no">>~/.ssh/config`}></CopyButton>
                                        </div>
                                    </List.Item>
                                    {workstation.data?.config?.image?.includes("intellij") || workstation.data?.config?.image?.includes("pycharm") ?
                                        (
                                            <div>
                                                <List.Item title={"Sørg for at Remote Development Gateway pluginen er installert i IntelliJ/PyCharm"}>
                                                    For installasjon av pluginen se <Link href="https://www.jetbrains.com/help/idea/jetbrains-gateway.html">her</Link>.
                                                </List.Item>
                                                <List.Item title={"Åpne Remote Development i IntelliJ/PyCharm (File | Remote Development...)"}>
                                                    <ul>
                                                        <li>Følg så denne <Link href="https://www.jetbrains.com/help/idea/remote-development-starting-page.html#connect_to_rd_ij">guiden</Link></li>
                                                    </ul>
                                                </List.Item>
                                            </div>
                                        ) : (
                                            <div>
                                                <List.Item>
                                                    Installere extension <code>Remote - SSH</code> i VS Code
                                                </List.Item>
                                                <List.Item title={"Åpne Command Palette i VS Code (⇧⌘P / Ctrl+Shift+P)"}>
                                                    <ul>
                                                        <li>Velg/Skriv inn: <code>Remote - SSH: Connect to host...</code></li>
                                                        <li>Skriv inn: <strong>knast</strong></li>
                                                    </ul>
                                                </List.Item>
                                                <p>Dette er også beskrevet med skjermbilder i <Link
                                                    href="https://cloud.google.com/workstations/docs/develop-code-using-local-vscode-editor">dokumentasjonen
                                                    til Google Workstations.</Link></p>
                                            </div>
                                        )}
                                </List>
                            </BodyLong>
                        </Modal.Body>
                    </Modal>
                </div>
            )
        case Workstation_STATE_STOPPING:
            return (
                <div className="flex flex-col gap-4">
                    <p>Stopper din Knast <Loader size="small" transparent /></p>
                    <div className="flex gap-2">
                        {startStopButtons(true, true)}
                    </div>
                </div>
            )
        case Workstation_STATE_STARTING:
            return (
                <div className="flex flex-col gap-4">
                    <p>Starter din Knast <Loader size="small" transparent /></p>
                    <div className="flex gap-2">
                        {startStopButtons(true, true)}
                    </div>
                </div>
            )
        case Workstation_STATE_STOPPED:
            return (
                <div className="flex flex-col gap-4">
                    {hasRunningJob && (<Alert variant={'info'}>Det er endring pågar, da kan Knast ikke starte</Alert>)}
                    <div className="flex gap-2">
                        {startStopButtons(hasRunningJob, true)}
                    </div>
                </div>
            )
    }
}

export default WorkstationStatus;
