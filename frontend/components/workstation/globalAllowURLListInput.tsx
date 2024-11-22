import {Textarea, RadioGroup, Radio, Stack} from "@navikt/ds-react";
import {forwardRef} from "react";
import {useWorkstation} from "./WorkstationStateProvider";

export const GlobalAllowUrlListInput = forwardRef<HTMLFieldSetElement, {}>(({}, ref) => {
   const {workstation, workstationOptions} = useWorkstation()

    const defaultKeepGlobalOpenings = workstation?.config?.disableGlobalURLAllowList ?? false
    const urlList = workstationOptions?.globalURLAllowList ?? ["Klarte ikke å hente URL-er for fremvisning :("]

    const description = "En sentralt administrert liste av URL-er, tilgjengelig for alle brukere, for å gi en bedre brukeropplevelse."

    return (
        <div className="flex gap-2 flex-col">
            <RadioGroup ref={ref} legend="Sentralt administrerte åpninger" defaultValue={defaultKeepGlobalOpenings ? "true" : "false"} description={description}>
                <Stack gap="0 6" direction={{ xs: "column", sm: "row" }} wrap={false}>
                    <Radio value="false">Skru på (anbefalt)</Radio>
                    <Radio value="true">Skru av</Radio>
                </Stack>
            </RadioGroup>
            <Textarea label="Åpninger du vil få tilgang til" defaultValue={urlList.join("\n")} size="small" maxRows={2500} readOnly resize />
        </div>
    );
})

GlobalAllowUrlListInput.displayName = "GlobalAllowUrlListInput";

export default GlobalAllowUrlListInput;
