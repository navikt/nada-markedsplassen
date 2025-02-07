import { ArrowRightIcon, ExternalLinkIcon } from "@navikt/aksel-icons";
import {
    Button,
    FormProgress,
    GuidePanel,
    Heading, HelpText,
    Link,
    List, Loader,
    VStack
} from '@navikt/ds-react';
import { useContext, useEffect, useState } from 'react';
import { UserState } from "../../lib/context";
import ContainerImageSelector from './formElements/containerImageSelector';
import MachineTypeSelector from './formElements/machineTypeSelector';
import { useCreateWorkstationJob, useWorkstationOptions } from './queries';

export interface WorkstationSetupPageProps {
    startedGuide: boolean;
    setStartedGuide: (value: boolean) => void;
}

const WorkstationSetupPage = (props: WorkstationSetupPageProps) => {
    const userInfo = useContext(UserState)
    const options = useWorkstationOptions();
    const createWorkstationJob = useCreateWorkstationJob();

    const [selectedContainerImage, setSelectedContainerImage] = useState<string>(options.data?.containerImages[0]?.image || '')
    const [selectedMachineType, setSelectedMachineType] = useState<string>(options.data?.machineTypes[0]?.machineType || '')
    const [activeStep, setActiveStep] = useState(1);

    useEffect(() => {
        console.log(selectedMachineType)
        console.log(selectedContainerImage)
    }, [selectedContainerImage, selectedMachineType]);

    const handleSubmit = (event: React.FormEvent) => {
        event.preventDefault();

        const input = {
            machineType: selectedMachineType,
            containerImage: selectedContainerImage,
        };

        createWorkstationJob.mutate(input)
        props.setStartedGuide(false)
    };

    if (options.isLoading) {
        return <Loader size="large" title="Laster..."/>;
    }

    if (!props.startedGuide) {
        return (
            <VStack as="main" gap="12">
                <GuidePanel poster>
                    <div className="flex flex-col gap-8">
                        <div>
                            <Heading level="2" size="medium" spacing>
                                Hei, {userInfo?.name}!
                            </Heading>
                        </div>
                        <div>
                            <div className="flex gap-1">
                                Jeg er her for å hjelpe deg å komme i gang med din nye arbeidsmaskin i skyen, også
                                kalt <strong>Knast</strong>
                                <HelpText title="right" placement="right">
                                    Takk til Beate Sildnes for navnet!
                                </HelpText>
                            </div>
                        </div>
                        <div>
                            På de påfølgende sidene vil du få muligheten til å velge:
                            <List>
                                <List.Item>
                                    <strong>Maskintype</strong>: hvor mye minne og prosessorkraft du trenger
                                </List.Item>
                                <List.Item>
                                    <strong>Utviklingsmiljø</strong>: programvaren, verktøyene og bibliotekene du ønsker
                                    å
                                    starte med
                                </List.Item>
                            </List>
                        </div>
                        <div>
                            Du kan når som helst gjøre endringer på <strong>alle</strong> dine valg.
                        </div>
                    </div>
                </GuidePanel>
                <div>
                    Det eneste vi krever for øyeblikket er at du har tilgang til <strong>naisdevice</strong>, og dette
                    er et krav vi håper å kunne fjerne i fremtiden.
                </div>
                <div>
                    En kjørende Knast vil <strong>stenges etter 2 timer uten aktivitet</strong>. Den vil også ha en hard
                    grense på <strong>12 timer</strong> for hver økt. Dette er for å sikre at ressursene i skyen ikke
                    blir brukt unødvendig, og ha muligheten til å kjøre sikkerthetsoppdateringer.
                </div>
                <div>
                    Hvis du har spørsmål til noe rundt opprettelsen av din Knast, eller er ellers usikker på noe så
                    kan du nå oss på Slack i{" "}
                    <Link target="_blank" href="https://nav-it.slack.com/archives/CGRMQHT50">
                        #nada <ExternalLinkIcon/>
                    </Link>
                </div>
                <div>
                    <Button
                        variant="primary"
                        icon={<ArrowRightIcon aria-hidden/>}
                        iconPosition="right"
                        onClick={() => props.setStartedGuide(true)}
                    >
                        Start opprettelse av Knast
                    </Button>
                </div>
                <div></div>
            </VStack>
        );
    }

    return (
        <div className="flex flex-col gap-8">
            <div></div>
            <div>
                <Heading size="medium">Konfigurer oppsettet av din Knast!</Heading>
            </div>
            <FormProgress
                totalSteps={5}
                activeStep={activeStep}
                onStepChange={setActiveStep}
            >
                <FormProgress.Step href="#step1">Maskintype</FormProgress.Step>
                <FormProgress.Step href="#step2">Utviklingsmiljø</FormProgress.Step>
                <FormProgress.Step href="#step3">Brannmuråpninger</FormProgress.Step>
                <FormProgress.Step href="#step4">Personlig administrerte URLer</FormProgress.Step>
                <FormProgress.Step href="#step5">Sentralt administrerte URLer</FormProgress.Step>
            </FormProgress>
            <div>
                <form onSubmit={handleSubmit}>
                    {activeStep === 1 &&
                        <div className="flex flex-col gap-8">
                            <Heading size="large">
                                Maskintype
                            </Heading>
                            <div>
                                Maskintypen bestemmer hvor mye minne, prosessorkraft, disk, og båndbredde du har
                                tilgjengelig for din neste arbeidssesjon. Du får noe enkel informasjon om hver
                                maskintype her,
                                men det er også mulig å lese den detaljerte dokumentasjonen om <Link
                                target="_blank"
                                href="https://cloud.google.com/compute/docs/general-purpose-machines#n2d_machines">N2D
                                maskin familien <ExternalLinkIcon/></Link>.
                            </div>
                            <div>
                                Hvis du er usikker på hvilken maskin du skal velge, så anbefaler vi å starte
                                med <strong>n2d-standard-2</strong>.
                                Husk at du kan endre maskintype når som helst hvis du finner ut at du trenger mer
                                ressurser for en analyse eller et prosjekt.
                            </div>
                            <MachineTypeSelector initialMachineType={selectedMachineType}
                                                 handleSetMachineType={setSelectedMachineType}
                            />
                        </div>
                    }
                    {activeStep === 2 &&
                        <div className="flex flex-col gap-8">
                            <Heading size="large">
                                Utviklingsmiljø
                            </Heading>
                            <div>
                                Utviklingsmiljøet bestemmer hvilken programvare, verktøy og biblioteker som er
                                tilgjengelig for deg når du starter din Knast. Dette er for at du skal komme raskere i
                                gang med ditt arbeid.
                            </div>
                            <div>
                                Hvis du er usikker på hvilken du skal velge, så anbefaler vi å starte
                                med VSCode, siden den kan brukes i nettleseren fra din lokale maskin.
                            </div>
                            <div>
                                Noen ting som er verdt å nevne:
                                <List>
                                    <List.Item>
                                        Du har <strong>root</strong> på din Knast, og kan installere hva du vil
                                        via <strong>apt</strong> eller <strong>pip</strong>, f.eks.
                                    </List.Item>
                                    <List.Item>
                                        All data som blir lagt under <strong>/home</strong> lagres permanent, dvs., at
                                        det vil overleve en omstart, bytte av maskintype, utviklingsmiljø, eller andre
                                        endringer.
                                    </List.Item>
                                    <List.Item>
                                        Det er mulig å lage utviklingsmiljø tilpasset deg eller ditt team, f.eks. med
                                        spesielt installerte verktøy, biblioteker, etc. Dette håndteres via <Link
                                        target="_blank"
                                        href="https://github.com/navikt/knast-images">knast-images <ExternalLinkIcon/></Link>.
                                    </List.Item>
                                </List>
                            </div>
                            <ContainerImageSelector initialContainerImage={selectedContainerImage}
                                                    handleSetContainerImage={setSelectedContainerImage}/>
                        </div>
                    }
                </form>
            </div>
            <div className="flex flex-row gap-4">
                {activeStep > 1 && activeStep <= 5 &&
                    <Button variant="secondary" onClick={() => setActiveStep(activeStep - 1)}>Forrige</Button>}
                {activeStep < 5 && <Button onClick={() => setActiveStep(activeStep + 1)}>Neste</Button>}
                {activeStep === 5 && <Button type="submit" onClick={handleSubmit}>Opprett din Knast</Button>}
            </div>
            <div/>
        </div>
    );
};

export default WorkstationSetupPage;
