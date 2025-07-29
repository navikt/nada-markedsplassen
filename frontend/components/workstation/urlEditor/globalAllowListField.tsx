import {ChevronDownIcon, ChevronUpIcon, TasklistIcon } from "@navikt/aksel-icons";
import { Button, ExpansionCard, HStack, Link, List, Popover, Radio, RadioGroup, Stack } from "@navikt/ds-react";
import { useRef, useState } from "react";

interface GlobalAllowListSelectorProps {
    urls: string[];
    optIn: boolean;
    onChange: (value: boolean) => void;
}

const GlobalAllowListField = ({ onChange, urls, optIn }: GlobalAllowListSelectorProps) => {
    const description = "Noen åpninger mot internett har mange nytte av og vi har derfor valgt å åpne disse som standard for alle brukere. Men, du står fritt til å ikke åpne for disse."
    const [showList, setShowList] = useState(false);
    return (
        <div className="flex gap-2 flex-col">
            <RadioGroup legend="Sentralt administrerte åpninger mot internett" defaultValue={optIn} value={optIn}
                description={description} onChange={onChange}>
                <Stack gap="0 6" direction={{ xs: "column", sm: "row" }} wrap={false}>
                    <Radio value={true}>Behold åpninger (anbefalt)</Radio>
                    <Radio value={false}>Ikke beholde åpninger</Radio>
                    <Link href="#" onClick={() => setShowList(!showList)}><TasklistIcon title="a11y-title" fontSize="1.5rem" />
                        {showList ? (
                            <>Skjul URL-listen<ChevronUpIcon/></>
                            ) : (
                            <>Vis URL-listen<ChevronDownIcon/></>
                            )}
                    </Link>
                </Stack>
                <div
                    hidden={!showList}
                    className="border border-orange-200 bg-orange-50 p-2 rounded-md  mt-2"
                >
                    <List as="ul" size="small">
                        {urls.map((url, index) => (
                            <List.Item key={index}>{url}</List.Item>
                        ))}
                    </List>
                </div>

            </RadioGroup>



        </div>
    );

}

export default GlobalAllowListField
