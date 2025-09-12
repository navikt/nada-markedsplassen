import { useState, useEffect } from "react";
import { Button, RadioGroup, Radio, VStack, Alert, Heading } from "@navikt/ds-react";

interface Question {
  question: string;
  options: string[];
  correct: number;
}

const questions: Question[] = [
  {
    question: "I hvilke sammenhenger kan man få tilgang til data?",
    options: ["Ved tjenestelige behov", "Alltid", "Aldri"],
    correct: 0,
  },
  {
    question: "Hvis man bruker et bibilotek, hva er viktig å huske på?",
    options: ["At det har et kult navn", "Holde det oppdatert"],
    correct: 1,
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
    <VStack gap="8">
      <Heading size="medium">Quiz</Heading>
      {questions.map((q, idx) => (
        <div key={idx}>
          <RadioGroup
            legend={q.question}
            value={answers[idx] !== null ? String(answers[idx]) : ""}
            onChange={(val) => handleChange(idx, val)}
          >
            {q.options.map((opt, optIdx) => (
              <Radio key={optIdx} value={String(optIdx)}>{opt}</Radio>
            ))}
          </RadioGroup>
          {answers[idx] !== null && (
            answers[idx] === q.correct ? (
              <Alert variant="success" size="small">Riktig!</Alert>
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
