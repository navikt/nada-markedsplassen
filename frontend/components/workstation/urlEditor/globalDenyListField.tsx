import { ExclamationmarkTriangleIcon } from "@navikt/aksel-icons";
import { Link, List } from "@navikt/ds-react";
import { useState } from "react";

interface GlobalDenyListFieldProps {
    urls: string[];
}

const GlobalDenyListField = ({ urls }: GlobalDenyListFieldProps) => {
    const [showList, setShowList] = useState(false);
    
    if (!urls || urls.length === 0) {
        return null;
    }

    return (
        <div className="flex gap-2 flex-col">
            <div className="flex items-center gap-2">
                <ExclamationmarkTriangleIcon className="text-orange-500" fontSize="1.5rem" />
                <span className="font-medium text-sm">Globalt blokkerte URLer</span>
                <Link href="#" onClick={() => setShowList(!showList)}>
                    {showList ? "Skjul listen" : "Vis listen"}
                </Link>
            </div>
            <div
                hidden={!showList}
                className="border border-orange-200 bg-orange-50 p-2 rounded-md mt-2"
            >
                <div className="text-xs text-orange-700 mb-2">
                    Disse URL-ene er globalt blokkert og kan ikke aksesseres fra Knast.
                </div>
                <List as="ul" size="small">
                    {urls.map((url, index) => (
                        <List.Item key={index} className="text-red-700">{url}</List.Item>
                    ))}
                </List>
            </div>
        </div>
    );
};

export default GlobalDenyListField;
