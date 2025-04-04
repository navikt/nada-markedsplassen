import { BodyShort, Heading, VStack } from "@navikt/ds-react";

const WorkstationPythonSetup= () => {
    return (
        <div className="basis-1/2">
            <VStack gap="5">
                <Heading size="medium">Autentisering mot pypi-proxy</Heading>
                <BodyShort size="medium">Vi har laget en global deny-regel mot <strong>pypi.org</strong>. For å kunne laste ned Python pakker ønsker vi at dere går via vår pypi-proxy.</BodyShort>
                <BodyShort size="medium">Vi har konfigurert en global pip config <strong>/etc/pip.conf</strong>, som ikke tar stilling til hvilken package manager (uv, pip, poetry) du bruker, så lenge den er kompatibel med pip.</BodyShort>
                <BodyShort size="medium">For å kunne bruke denne configen må du kjøre følgende kommandoer:</BodyShort>
                <code>
                    $ gcloud auth login --update-adc
                    <br />
                    $ pypi-auth
                </code>
                <Heading size="small">gcloud auth login --update-adc</Heading>
                <BodyShort size="medium">Med denne kommandoen logger du inn på Google Cloud med din personlige bruker og oppdaterer Application Default Credentials (ADC).</BodyShort>
                <Heading size="small">pypi-auth</Heading>
                <BodyShort size="medium">Denne kommandoen vil konfigurere <strong>.netrc</strong> filen i ditt <strong>$HOME</strong> directory for pypi artifact registry ved å bruke en autentiseringsnøkkel som er hentet fra <code>gcloud auth login --update-adc</code>. Denne nøkkelen er kortlevd, 1 time, og du må oppdatere den etter denne perioden ved å kjøre <code>pypi-auth</code> på nytt.</BodyShort>

                <Heading size="medium">Oppsett av Poetry</Heading>
                <BodyShort size="medium">For å bruke Poetry med pypi-proxy må du oppdatere <strong>pyproject.toml</strong> filen din med følgende innhold:</BodyShort>
                <code>
                    [[tool.poetry.source]]
                    <br />
                    name = "private"
                    <br />
                    url = "https://europe-north1-python.pkg.dev/knada-gcp/pypiproxy/simple/"
                    <br />
                    priority = "supplemental"
                </code>
                <BodyShort size="medium">Etter dette kan du kjøre <code>poetry add ...</code> som vanlig.</BodyShort>
            </VStack>
        </div>
    );
};

export default WorkstationPythonSetup;
