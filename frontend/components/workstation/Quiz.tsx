import { useState, useEffect } from "react";
import { Button, RadioGroup, Radio, VStack, Alert, Heading } from "@navikt/ds-react";
import { List, Link } from '@navikt/ds-react';
import { PadlockLockedIcon, KeyHorizontalIcon, FileTextIcon, ArrowsCirclepathIcon, GlobeIcon, BranchingIcon, CloudDownIcon } from '@navikt/aksel-icons';

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
    <VStack gap="8" justify="space-between" align="start" paddingInline="space-16">
    <Heading size="large">Sikker jobbing med skarpe data</Heading>
    <p>Når du jobber med skarpe data, så er det viktig å tenke sikkerhet hele veien.</p>

    <List>
      <List.Item icon={<PadlockLockedIcon aria-hidden />} className={"flex items-center gap-4"}>
        <strong>Tjenestelig behov</strong>: Du skal kun ha tilgang til data du faktisk trenger for å utføre jobben din.
      </List.Item>
      <List.Item icon={<KeyHorizontalIcon aria-hidden/> } className={"flex items-center gap-4"}>
        <strong>Least privilege</strong>: Minimer rettigheter – det reduserer risiko.
      </List.Item>
      <List.Item icon={<KeyHorizontalIcon aria-hidden/> } className={"flex items-center gap-4"}>
        <strong>Dataminimering</strong>: Begrense mengden data du samler inn og behandler til det som er nødvendig for formålet.
      </List.Item>
      <List.Item icon={<FileTextIcon aria-hidden/>} className={"flex items-center gap-4"}>
        <strong>Logger</strong>: Vær obs på at sensitive data kan havne i logger.
      </List.Item>
      <List.Item icon={<ArrowsCirclepathIcon aria-hidden/>} className={"flex items-center gap-4"}>
        <strong>Tredjepartsbiblioteker</strong>: Hold dem oppdatert – gamle versjoner kan ha sårbarheter.
      </List.Item>
      <List.Item icon={<GlobeIcon aria-hidden/>} className={"flex items-center gap-4"}>
        <strong>Internett-tilgang</strong>: Begrens åpninger mot internett mest mulig.
      </List.Item>
      <List.Item icon={<BranchingIcon aria-hidden/>} className={"flex items-center gap-4"}>
        <strong>Git</strong>: API-nøkler, passord og personopplysninger må aldri pushes til GitHub.
      </List.Item>
      <List.Item icon={<CloudDownIcon aria-hidden/>} className={"flex items-center gap-4"}>
        <strong>Lokal lagring</strong>: Ikke last ned skarpe data til din PC, eller lagre dem permanent i Knast.
      </List.Item>
    </List>
    <Heading size="small">Relevante ressurser:</Heading>
    <VStack gap="1">
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
