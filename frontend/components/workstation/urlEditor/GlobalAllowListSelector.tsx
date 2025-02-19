import { TasklistIcon } from "@navikt/aksel-icons";
import { Button, ExpansionCard, HStack, Link, List, Popover, Radio, RadioGroup, Stack } from "@navikt/ds-react";
import { useRef, useState } from "react";

interface GlobalAllowListSelectorProps {
    urls: string[];
    optIn: boolean;
    onChange: (value: boolean) => void;
}

const GlobalAllowListSelector = ({ onChange, urls, optIn }: GlobalAllowListSelectorProps) => {
    const description = "En sentralt administrert liste av URLer, tilgjengelig for alle brukere."
    const [showList, setShowList] = useState(false);
    return (
        <div className="flex gap-2 flex-col">
            <RadioGroup legend="Sentralt adm  inistrerte åpninger" defaultValue={optIn} value={optIn}
                description={description} onChange={onChange}>
                <Stack gap="0 6" direction={{ xs: "column", sm: "row" }} wrap={false}>
                    <Radio value={true}>Behold (anbefalt)</Radio>
                    <Radio value={false}>Skru av</Radio>
                    <Link href="#" onClick={() => setShowList(!showList)}><TasklistIcon title="a11y-title" fontSize="1.5rem" />
                        {showList ? "Skjul url-listen" : "Vis url-listen"}</Link>
                </Stack>
                <p
                    hidden={!showList}
                    className="border border-gray-300 p-2 rounded-md w-[30rem] mt-2"
                >
                    <List as="ul" size="small">
                        {urls.map((url, index) => (
                            <List.Item key={index}>{url}</List.Item>
                        ))}
                    </List>
                </p>

            </RadioGroup>



        </div>
    );

}

export default GlobalAllowListSelector