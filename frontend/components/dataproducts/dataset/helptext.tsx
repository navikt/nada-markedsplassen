import { HelpText } from "@navikt/ds-react";

const Personopplysninger = () => {
    return (
        <HelpText title="Hvor kommer dette fra?">
            <p>Personopplysninger er alle opplysninger som kan identifisere en fysisk person,
            enten det er selve opplysningen eller kombinasjonen av opplysninger som gjør
            det mulig å finne ut hvem den fysiske personen er.</p>
            <p>Husk at det som regel ikke holder å fjerne direkte identifiserende opplysninger,
            uten å ta høyde for kombinasjonen av informasjon man sitter igjen med.</p>
            <a href="https://www.datatilsynet.no/rettigheter-og-plikter/personopplysninger/" target="_blank"
               rel="noopener noreferrer">
                Les mer om personopplysninger hos datatilsynet.

            </a>
        </HelpText>
    );
};

const PersonopplysningerDetaljert = () => {
    return (
        <HelpText title="Hvor kommer dette fra?">
            <p className="font-bold">Personopplysninger kan klassifiseres som direkte identifiserende eller av særlig kategori.</p>
            <a href="https://www.datatilsynet.no/rettigheter-og-plikter/personopplysninger/" target="_blank"
               rel="noopener noreferrer">
                Les mer om særlig kategori av personopplysninger hos datatilsynet.
            </a>
            <p>
                Husk at det som regel ikke holder å fjerne direkte identifiserende opplysninger,
                uten å ta høyde for kombinasjonen av informasjon man sitter igjen med.
            </p>
        </HelpText>
    );
};

const TilgangsstyringHelpText = () => {
    return <HelpText title="Hva betyr det å gi tilgang til alle i Nav?">
        <p className="mb-2">Det å gi tilgang til alle i Nav betyr at man ikke trenger å søke om tilgang for å lese datasettet fra BigQuery og i Metabase.</p>
        <p>Det er mulig å gjøre det om senere. Du vil kunne se gruppen som omfavner hele Nav i oversikten over aktive tilganger i datasettet ditt, og du har mulighet til å fjerne denne tilgangen.</p>
    </HelpText>
}

const PseudonymiseringsText = () => {
    return <HelpText title="Pseudonymisering er ikke det samme som anonymisering!">
        <p className="mb-2"> Selv om du pseudonymiserer enkelte kolonner i tabellen, vil det ofte være mulig å reidentifisere personer gjennom annen informasjon i tabellen. Derfor er pseudnoymisering noe annet anonymisering. </p>
    </HelpText>
}


export {Personopplysninger, PersonopplysningerDetaljert, TilgangsstyringHelpText, PseudonymiseringsText}
