import {Textarea, RadioGroup, Radio, Stack} from "@navikt/ds-react";
import {useState} from "react";
import {useWorkstationMine, useWorkstationOptions} from "../queries";

export interface GlobalAllowUrlListInputProps {
    disabled?: boolean;
    setDisabled?: (disabled: boolean) => void;
}

export const GlobalAllowUrlListInput = (props: GlobalAllowUrlListInputProps) => {
    const options= useWorkstationOptions()
    const workstation= useWorkstationMine()

    const [defaultKeepGlobalOpenings, setDefaultKeepGlobalOpenings] = useState<boolean>(props.disabled ?? workstation.data?.config?.disableGlobalURLAllowList ?? false);

    const urlList = options.data?.globalURLAllowList ?? ["Klarte ikke å hente URL-er for fremvisning :("]

    const description = "En sentralt administrert liste av URL-er, tilgjengelig for alle brukere, for å gi en bedre brukeropplevelse."

    if (options.isLoading || workstation.isLoading) {
        return <Textarea label="Åpninger du vil få tilgang til" defaultValue="Laster..." size="small" maxRows={2500}
                         readOnly resize/>
    }

    function handleChange(val: boolean) {
        setDefaultKeepGlobalOpenings(val)

        if (props.setDisabled) {
            props.setDisabled(val)
        }
    }

    return (
        <div className="flex gap-2 flex-col">
            <RadioGroup legend="Sentralt administrerte åpninger" value={defaultKeepGlobalOpenings}
                        description={description} onChange={handleChange}>
                <Stack gap="0 6" direction={{xs: "column", sm: "row"}} wrap={false}>
                    <Radio value={false}>Behold (anbefalt)</Radio>
                    <Radio value={true}>Skru av</Radio>
                </Stack>
            </RadioGroup>
            <Textarea label="Åpninger du vil få tilgang til" defaultValue={urlList.join("\n")} size="small"
                      maxRows={2500} readOnly resize/>
        </div>
    );
}

export default GlobalAllowUrlListInput;
