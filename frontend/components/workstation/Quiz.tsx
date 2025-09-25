import { useState, useEffect } from "react";
import { Button, RadioGroup, Radio, VStack, Alert, Heading } from "@navikt/ds-react";
import { List, Link } from '@navikt/ds-react';
import { PadlockLockedIcon, PersonGroupIcon, EyeClosedIcon, HardHatIcon, HourglassIcon, DatabaseIcon, MigrationIcon, BroadcastMinusCircleIcon, ThumbUpIcon, ShieldCheckmarkIcon, ExclamationmarkTriangleIcon, KeyHorizontalIcon, FileTextIcon, ArrowsCirclepathIcon, GlobeIcon, BranchingIcon, CloudDownIcon } from '@navikt/aksel-icons';

interface Question {
  question: string;
  options: string[];
  correct: number;
  correct_answer_details: React.ReactNode;
}

const questions: Question[] = [
  {
    question: "I hvilke sammenhenger kan man få tilgang til data?",
    options: ["Ved tjenestelige behov", "Alltid", "Aldri"],
    correct: 0,
    correct_answer_details: "Tjenestelig behov betyr at du kun skal ha tilgang til data du faktisk trenger for å utføre jobben din.",
  },
  {
    question: "Hvis man bruker et bibilotek, hva er viktig å huske på?",
    options: ["At det har et kult navn", "Holde det oppdatert"],
    correct: 1,
      correct_answer_details: (
        <>
          Hvis du har koden din i Github, så kan du få Dependabot til å hjelpe deg med å holde tredjepartsbiblioteker oppdatert.{" "}
          <Link href="https://docs.knada.io/analyse/notebook/generelt/#bruk-av-github-advanced-security-og-dependabot" target="_blank" rel="noopener noreferrer">Les mer om Dependabot her.</Link>
        </>
      ),
  },
];

interface QuizProps {
  onQuizResult?: (isAllCorrect: boolean) => void;
}

const Quiz = ({ onQuizResult }: QuizProps) => {
  const [answers, setAnswers] = useState<(number | null)[]>(Array(questions.length).fill(null));

  useEffect(() => {
    if (onQuizResult) {
      const allCorrect = answers.every((ans, idx) => ans === questions[idx].correct);
      onQuizResult(allCorrect);
    }
  }, [answers, onQuizResult]);
  const handleChange = (questionIdx: number, value: string) => {
    const newAnswers = [...answers];
    newAnswers[questionIdx] = parseInt(value, 10);
    setAnswers(newAnswers);
  };

  return (
    <VStack gap="1" justify="space-between" align="start" paddingInline="space-16">
    <Heading size="large">Sikkerhetstips til jobbing med skarpe data</Heading>
    <p>Når du jobber med skarpe data, så er det viktig å tenke sikkerhet hele veien.</p>

    <List>
    <List.Item icon={<PadlockLockedIcon aria-hidden />} className={"flex text-start items-center gap-4"}>
    <strong>Tjenestelig behov</strong>: Du (og koden din) skal kun ha tilgang til data og ressurser som er nødvendige for å utføre jobben.
        </List.Item>
      <List.Item icon={<KeyHorizontalIcon aria-hidden/> }  className={"flex text-start items-center gap-4"}>
    <strong>Least privilege</strong>: Minimer rettigheter for både brukere og systemer. Jo færre rettigheter, jo mindre skade hvis noe går galt.

        </List.Item>
      <List.Item icon={<DatabaseIcon aria-hidden/> }  className={"flex text-start items-center gap-4"}>
    <strong>Dataminimering</strong>: Samle inn og behandle kun den mengden data som er nødvendig for formålet. Unngå å lagre detaljer på data som ikke trengs.

        </List.Item>
      <List.Item icon={<FileTextIcon aria-hidden/>}  className={"flex text-start items-center gap-4"}>
    <strong>Logger</strong>: Unngå at sensitive data havner i logger. Masker eller fjern personopplysninger.

      </List.Item>
    <List.Item icon={<ThumbUpIcon aria-hidden/>}  className={"flex text-start items-center gap-4"}>
    <strong>Redusere konsekvens</strong>:Tenk gjennom hva som skjer hvis passord, nøkler eller data kommer på avveie, eller hvis koden manipuleres.
      </List.Item>
    <div className="flex text-start pl-10 pb-4 flex-col items-start">
    <List.Item icon={<HourglassIcon aria-hidden/>}  className={"flex text-start items-center gap-4"}>
    Begrense oppbevaringstid.
      </List.Item>
    <List.Item icon={<MigrationIcon aria-hidden/>}  className={"flex text-start items-center gap-4"}>
    Aggreger eller anonymiser data før de forlater godkjent behandlingsflate.
      </List.Item>
    <List.Item icon={<BroadcastMinusCircleIcon aria-hidden/>}  className={"flex text-start items-center gap-4"}>
    Ikke dele data videre uten godkjenning.
      </List.Item>
    </div>
    <List.Item icon={<ShieldCheckmarkIcon aria-hidden/>}  className={"flex text-start items-center gap-4"}>
    <strong>Redusere sannsynliget</strong>: Gjør det vanskelig for angrep å lykkes:
        </List.Item>
      <div className="flex text-start pl-10 pb-4 flex-col items-start">
    <List.Item icon={<GlobeIcon aria-hidden/>}  className={"flex text-start items-center gap-4"}>
    Begrens åpninger mot internett mest mulig.
      </List.Item>
    <List.Item icon={<ArrowsCirclepathIcon aria-hidden/>}  className={"flex text-start items-center gap-4"}>
    Hold kode og tredjepartsbiblioteker oppdatert.
      </List.Item>
    <List.Item icon={<ExclamationmarkTriangleIcon aria-hidden/>}  className={"flex text-start items-center gap-4"}>
    <strong>Vær obs på fremmed kode</strong>: Sjekk kilde, sikkerhetsstatus og oppdateringer før bruk.
      </List.Item>
    </div>
    <List.Item icon={<BranchingIcon aria-hidden/>}  className={"flex text-start items-center gap-4"}>
    <strong>Git</strong>: API-nøkler, passord og personopplysninger må aldri pushes til GitHub.
      </List.Item>
    <List.Item icon={<CloudDownIcon aria-hidden/>}  className={"flex text-start items-center gap-4"}>
    <strong>Lagring</strong>: Ikke last ned skarpe data til din PC, ikke lagre dem permanent i Knast, eller på uautoriserte flater utenfor Navs kontroll.
      </List.Item>
    <List.Item icon={<HardHatIcon aria-hidden/>}  className={"flex text-start items-center gap-4"}>
    <strong>Arbeidsform</strong>:
        </List.Item>
      <div className="flex text-start pl-10 pb-4 flex-col items-start">
    <List.Item icon={<PersonGroupIcon aria-hidden/>}  className={"flex text-start items-center gap-4"}>
    Unngå mob programmering med skarpe data, særlig i åpne landskap der mange kan se skjermen.
      </List.Item>
    <List.Item icon={<EyeClosedIcon aria-hidden/>}  className={"flex text-start items-center gap-4"}>
    Bare de som har tjenestelig behov skal ha tilgang, og visning av data må skje på en måte som hindrer innsyn fra uvedkommende.
      </List.Item>
    </div>
    </List>
    <VStack gap="2" className="pb-8">
    <Heading size="small" align="start">Relevante ressurser:</Heading>
    <Link href="https://navno.sharepoint.com/sites/intranett-personvern/SitePages/Start.aspx" target="_blank" rel="noopener noreferrer">
    Personvern i Nav
    </Link>
    <Link href="https://navno.sharepoint.com/sites/intranett-personvern/SitePages/Kurs%20og%20oppl%C3%A6ring.aspx#oppl%C3%A6ringsforslag-per-kompetanseomr%C3%A5de" target="_blank" rel="noopener noreferrer">
    Kurs og opplæring rundt personvern og informasjonssikkerhet
    </Link>
    <Link href="https://sikkerhet.nav.no" target="_blank" rel="noopener noreferrer">
    Security Playbook for Nav
      </Link>
    <Link href="https://www.datatilsynet.no/rettigheter-og-plikter/personvernprinsippene/grunnleggende-personvernprinsipper/" target="_blank" rel="noopener noreferrer">
    Datatilsynets grunnleggende personvernprinsipper
    </Link>
    </VStack>

<Heading size="medium">Dataquiz</Heading>
    {questions.map((q, idx) => (
      <div key={idx}>
      <RadioGroup
        legend={q.question}
        value={answers[idx] !== null ? String(answers[idx]) : ""}
        onChange={(val) => handleChange(idx, val)}
        style={{ textAlign: "left" }}
      >
      {q.options.map((opt, optIdx) => (
        <Radio key={optIdx} value={String(optIdx)}>{opt}</Radio>
      ))}
      </RadioGroup>
      {answers[idx] !== null && (
        answers[idx] === q.correct ? (
          <Alert variant="success" size="small">Riktig! {q.correct_answer_details}</Alert>
        ) : (
        <Alert variant="error" size="small">Dessverre feil.</Alert>
        )
      )}
      </div>
    ))}
    </VStack>
  );
};

export default Quiz;
