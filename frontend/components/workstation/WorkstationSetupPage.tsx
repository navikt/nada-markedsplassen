import {useState, useContext} from 'react';
import {
    BodyLong,
    Button,
    FormProgress,
    GuidePanel,
    Heading, HelpText,
    Link,
    List,
    VStack
} from '@navikt/ds-react';
import MachineTypeSelector from './formElements/machineTypeSelector';
import ContainerImageSelector from './formElements/containerImageSelector';
import FirewallTagSelector from './formElements/firewallTagSelector';
import UrlListInput from './formElements/urlListInput';
import GlobalAllowUrlListInput from './formElements/globalAllowURLListInput';
import {useCreateWorkstationJob, useWorkstationOptions} from './queries';
import {ArrowRightIcon} from "@navikt/aksel-icons";
import {UserState} from "../../lib/context";
import {ExternalLink} from "@navikt/ds-icons";
import {WorkstationContainer} from "../../lib/rest/generatedDto";

const WorkstationSetupPage = () => {
    const userInfo = useContext(UserState)
    const options = useWorkstationOptions();
    const createWorkstationJob = useCreateWorkstationJob();

    const [disableGlobalURLAllowList, setDisableGlobalURLAllowList] = useState(false);
    const [selectedUrlList, setSelectedUrlList] = useState<string[]>([]);
    const [selectedContainerImage, setSelectedContainerImage] = useState<WorkstationContainer>();
    const [selectedMachineType, setSelectedMachineType] = useState<string>('');
    const [selectedFirewallTags, setSelectedFirewallTags] = useState<string[]>([]);
    const [activeStep, setActiveStep] = useState(1);
    const [started, setStarted] = useState(false);

    const handleSubmit = (event: React.FormEvent) => {
        event.preventDefault();

        const input = {
            machineType: selectedMachineType,
            containerImage: selectedContainerImage ? selectedContainerImage.image : '',
            onPremAllowList: selectedFirewallTags,
            urlAllowList: selectedUrlList,
            disableGlobalURLAllowList: disableGlobalURLAllowList,
        };

        console.log(input);

        createWorkstationJob.mutate(input);
    };

    if (options.isLoading) {
        return <p>Loading...</p>;
    }

    if (!started) {
        return (
            <VStack as="main" gap="12">
                <GuidePanel poster>
                    <Heading level="2" size="medium" spacing>
                        Hei, {userInfo?.name}!
                    </Heading>
                    <BodyLong spacing>
                        <div className="flex gap-1">
                            Jeg er her for å hjelpe deg å komme i gang med din nye arbeidsmaskin i skyen, også
                            kalt <strong>Knast</strong>
                            <HelpText title="right" placement="right" inlist>
                                Kudos til Beate Sildnes for å ha vunnet navnekonkurransen, og tenkt ut dette navnet!
                            </HelpText>
                        </div>
                    </BodyLong>
                    <BodyLong spacing>
                        På de påfølgende sidene vil du få muligheten til å velge:
                        <List>
                            <List.Item>
                                <strong>Maskintype</strong>: hvor mye minne og prosessorkraft du trenger
                            </List.Item>
                            <List.Item>
                                <strong>Utviklingsmiljø</strong>: programvaren, verktøyene og bibliotekene du ønsker å
                                starte med
                            </List.Item>
                            <List.Item>
                                <strong>Brannmuråpninger</strong>: hvilke interne tjenester og porter du trenger å nå
                                fra maskinen
                            </List.Item>
                            <List.Item>
                                <strong>Tillate URL-er</strong>: hvilke URL-er, eller tjenester på internet, du trenger
                                å nå fra maskinen
                            </List.Item>
                            <List.Item>
                                <strong>Sentralt administrerte URL-er</strong>: om du vil beholde de URL-er som er åpnet
                                for alle Knaster, eller om du ønsker å administrere alt selv
                            </List.Item>
                        </List>
                    </BodyLong>
                    <BodyLong>
                        Du kan når som helst gjøre endringer på <strong>alle</strong> dine valg.
                    </BodyLong>
                </GuidePanel>
                <div>
                    Det eneste vi krever for øyeblikket er at du har tilgang til <strong>naisdevice</strong>, og dette
                    er et krav vi håper å kunne fjerne i fremtiden.
                </div>
                <div>
                    Hvis du har spørsmål til noe rundt opprettelsen av din Knast, eller er ellers usikker på noe så
                    kan du nå oss på Slack i{" "}
                    <Link target="_blank" href="https://nav-it.slack.com/archives/CGRMQHT50">
                        #nada <ExternalLink/>
                    </Link>
                </div>
                <div>
                    <Button
                        variant="primary"
                        icon={<ArrowRightIcon aria-hidden/>}
                        iconPosition="right"
                        onClick={() => setStarted(true)}
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
                <FormProgress.Step href="#step4">Personlig administrerte URL-er</FormProgress.Step>
                <FormProgress.Step href="#step5">Sentralt administrerte URL-er</FormProgress.Step>
            </FormProgress>
            <div>
                <form onSubmit={handleSubmit}>
                    {activeStep === 1 &&
                        <div className="flex flex-col gap-4">
                            <Heading size="large">
                                Maskintype
                            </Heading>
                            <BodyLong>
                                Maskintypen bestemmer hvor mye minne, prosessorkraft, disk, og bredbånd du har
                                tilgjengelig for din neste arbeidssesjon.
                                I parentes kan du se noe informasjon om de forskjellige maskintypene,
                                men du kan også lese dokumentasjonen om <Link
                                href="https://cloud.google.com/compute/docs/general-purpose-machines#n2d_machines">N2D
                                maskin familien</Link>.
                                Hvis du er usikker på hvilken maskin du skal velge, så anbefaler vi å starte med
                                en <strong>n2d-standard-2</strong> maskin.
                                Husk at du kan endre maskintype når som helst hvis du finner ut at du trenger mer
                                ressurser for en analyse eller et prosjekt.
                            </BodyLong>
                            <MachineTypeSelector initialMachineType={selectedMachineType}
                                                 handleSetMachineType={setSelectedMachineType}
                            />
                        </div>
                    }
                    {activeStep === 2 &&
                        <div className="flex flex-col gap-4">
                            <Heading size="large">
                                Utviklingsmiljø
                            </Heading>
                            <BodyLong>
                                Utviklingsmiljøet bestemmer hvilken programvare, verktøy og biblioteker som er
                                tilgjengelig for deg når du starter din Knast. Dette er for at du skal komme raskere i
                                gang
                                med ditt arbeid. Hvis du er usikker på hvilken du skal velge, så anbefaler vi å starte
                                med
                                en VSCode variasjon, siden de kan brukes i nettleseren fra din lokale maskin. Det er
                                også
                                mulig å installere andre verktøy og programmer etter at Knasten er startet med et
                                utviklingsmiljø.
                                Alt som installeres til /home/ vil overleve omstarten av en Knasten, bytte av
                                maskintype, utviklingsmiljø, eller andre endringer.
                            </BodyLong>
                            <ContainerImageSelector initialContainerImage={selectedContainerImage}
                                                    handleSetContainerImage={setSelectedContainerImage}/>
                        </div>
                    }
                    {activeStep === 3 &&
                        <div className="flex flex-col gap-4">
                            <Heading size="large">
                                Brannmuråpninger
                            </Heading>
                            <BodyLong>
                                Brannmuråpninger bestemmer hvilke on-prem tjenester og porter du trenger å nå fra din
                                Knast, f.eks., Postgres eller Oracle databaser. Hvis du ikke trenger å nå noen interne
                                tjenester, eller du er usikker på hvilke som er aktuelle,
                                så kan du <strong>hoppe over</strong> dette steget. Det er fullt mulig å legge til eller
                                fjerne brannmuråpninger når som helst.
                            </BodyLong>
                            <FirewallTagSelector initialFirewallTags={selectedFirewallTags}
                                                 onFirewallChange={setSelectedFirewallTags}/>
                        </div>
                    }
                    {activeStep === 4 &&
                        <div className="flex flex-col gap-4">
                            <Heading size="large">
                                Personlig administrerte URL-er
                            </Heading>
                            <BodyLong>
                                Tillate URL-er bestemmer hvilke internett-URL-er du vil åpne mot fra din Knast. Hvis du
                                ikke trenger å åpne noen URL-er, eller du er usikker på hvilke som er aktuelle, så kan
                                du
                                <strong>hoppe over</strong> dette steget. Det er fullt mulig å legge til eller fjerne
                                URL-er når som helst.
                            </BodyLong>
                            <UrlListInput initialUrlList={selectedUrlList} onUrlListChange={setSelectedUrlList}/>
                        </div>
                    }
                    {activeStep === 5 &&
                        <div className="flex flex-col gap-4">
                            <Heading size="large">
                                Sentralt administrerte URL-er
                            </Heading>
                            <div>
                                Disse sentralt administrerte URL-ene kommer i <strong>tillegg</strong> til de URL-ene du har valgt å åpne selv. Det er altså ikke en erstatning for de URL-ene du har valgt å åpne selv.
                            </div>
                            <div className="flex flex-col gap-4">
                            <BodyLong>
                                Vi anbefaler at du beholder globale URL-åpninger bestemmer om du vil beholde de URL-er som er åpnet for
                                alle Knaster, eller om du vil ha en tom liste. Hvis du er usikker på hva dette betyr,
                                så anbefaler vi å beholde de globale URL-ene. Det er fullt mulig å legge til eller fjerne
                                URL-er når som helst.
                            </BodyLong>
                            <GlobalAllowUrlListInput disabled={disableGlobalURLAllowList} setDisabled={setDisableGlobalURLAllowList}/>
                            </div>
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
        </div>
    );
};

export default WorkstationSetupPage;
