import {ExpansionCard, HStack, List, Loader, Radio, RadioGroup, Stack} from "@navikt/ds-react";
import {useWorkstationOptions} from "../queries";
import {TasklistIcon} from "@navikt/aksel-icons";
import { useImperativeHandle, useState } from 'react'

export interface GlobalAllowUrlListInputProps {
    disabledGlobal: boolean;
    ref: React.Ref<{ getDisabled: () => boolean }>;
}

export const GlobalAllowUrlListInput = (props: GlobalAllowUrlListInputProps) => {
    const options = useWorkstationOptions()

    const [disabled, setDisabled] = useState<boolean>(props.disabledGlobal)

    useImperativeHandle(props.ref, () => ({
        getDisabled: () => {
            return disabled
        },
    }))

    const description = "En sentralt administrert liste av URLer, tilgjengelig for alle brukere, for å gi en bedre brukeropplevelse."

    function handleChange(val: boolean) {
        setDisabled(val)
    }

    return (
        <div className="flex gap-2 flex-col">
            <RadioGroup legend="Sentralt administrerte åpninger" defaultValue={props.disabledGlobal}
                        description={description} onChange={handleChange} disabled={options.isLoading}>
                <Stack gap="0 6" direction={{xs: "column", sm: "row"}} wrap={false}>
                    <Radio value={false}>Behold (anbefalt)</Radio>
                    <Radio value={true}>Skru av</Radio>
                </Stack>
            </RadioGroup>
            {options.isLoading && (
                <Loader size="small"/>
            )}
            {!options.isLoading && (
            <ExpansionCard size="small" aria-label="åpninger du vil få mot internett">
                <ExpansionCard.Header>
                    <HStack wrap={false} gap="4" align="center">
                        <div>
                            <TasklistIcon aria-hidden fontSize="3rem" />
                        </div>
                        <div>
                            <ExpansionCard.Description>
                                Åpninger du vil få tilgang til mot internett
                            </ExpansionCard.Description>
                        </div>
                    </HStack>
                </ExpansionCard.Header>
                <ExpansionCard.Content>
                    <List as="ul" size="small">
                        {options.data?.globalURLAllowList.map((url, index) => (
                            <List.Item key={index}>{url}</List.Item>
                        ))}
                    </List>
                </ExpansionCard.Content>
            </ExpansionCard>
            )}
        </div>
    );
}

export default GlobalAllowUrlListInput;
