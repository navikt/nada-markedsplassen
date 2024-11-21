import {Textarea, RadioGroup, Radio, Stack} from "@navikt/ds-react";

interface GlobalAllowURLListInputProps {
    urlList: string[];
    defaultValue: string;
    onDisableGlobalURLAllowList: (event: any) => void;
}

const GlobalAllowUrlListInput: React.FC<GlobalAllowURLListInputProps> = ({ urlList, defaultValue, onDisableGlobalURLAllowList }) => {
    return (
        <div className="flex gap-2 flex-col">
            <RadioGroup
                legend="Behold globale åpninger"
                defaultValue={defaultValue == "false" ? "false" : defaultValue == "true" ? "true" : "false"}
                onChange={onDisableGlobalURLAllowList}
                description="Vi har lagt til en liste over URL-er som er administrert sentralt, og tilgjengelig for alle brukere. Dette er åpninger som vil gi deg en bedre brukeropplevelse.
                Hvis du har behov, så kan du melde deg av disse URL-ene, men vi anbefaler at du ikke gjør det."
            >
                <Stack gap="0 6" direction={{ xs: "column", sm: "row" }} wrap={false}>
                    <Radio value="false">Ja (anbefalt)</Radio>
                    <Radio value="true">Nei</Radio>
                </Stack>
            </RadioGroup>
            <Textarea
                label="URL-er som er lagt til globalt"
                defaultValue={urlList.join("\n")}
                size="small"
                maxRows={2500}
                readOnly
                resize
            />
        </div>
    );
};

export default GlobalAllowUrlListInput;
