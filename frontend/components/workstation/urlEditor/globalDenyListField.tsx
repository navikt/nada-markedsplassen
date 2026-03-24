import { ExclamationmarkTriangleIcon, ChevronDownIcon, ChevronUpIcon, TasklistIcon, XMarkOctagonIcon } from "@navikt/aksel-icons";
import { List, Tag, Box } from "@navikt/ds-react";
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
        <div className="bg-white rounded-lg shadow-sm border p-6 mb-6">
            <div className="flex items-center gap-3 mb-4">
                <div className="w-8 h-8 bg-ax-danger-200 rounded-full flex items-center justify-center">
                    <XMarkOctagonIcon className="w-5 h-5 text-ax-danger-700" />
                </div>
                <div>
                    <h3 className="font-semibold text-ax-neutral-1000">Globalt blokkerte URL-er</h3>
                    <p className="text-sm text-ax-neutral-700 mt-1">
                        URL-er som er permanent blokkert av systemadministrator
                    </p>
                </div>
                <div className="ml-auto">
                    <Tag variant="error" size="small">
                        {urls.length} blokkert{urls.length !== 1 ? 'e' : ''}
                    </Tag>
                </div>
            </div>
            <div className="bg-ax-danger-100 border border-ax-danger-300 rounded-lg p-4 mb-4">
                <div className="flex items-start gap-2">
                    <ExclamationmarkTriangleIcon className="w-4 h-4 text-ax-danger-700 mt-0.5 flex-shrink-0" />
                    <p className="text-ax-danger-900 text-sm">
                        <strong>Advarsel:</strong> Disse URL-ene er globalt blokkert og kan ikke aksesseres fra Knast,
                        uavhengig av dine personlige innstillinger.
                    </p>
                </div>
            </div>
            <div className="border-t pt-4">
                <button
                    type="button"
                    onClick={() => setShowList(!showList)}
                    className="flex items-center gap-2 text-ax-accent-700 hover:text-ax-accent-900 text-sm font-medium transition-colors"
                >
                    <TasklistIcon className="w-4 h-4" />
                    <span>{showList ? "Skjul" : "Vis"} blokkerte URL-er ({urls.length} URL-er)</span>
                    {showList ? <ChevronUpIcon className="w-4 h-4" /> : <ChevronDownIcon className="w-4 h-4" />}
                </button>

                {showList && (
                    <div className="mt-4 border border-ax-danger-300 bg-ax-danger-100 rounded-lg p-4">
                        <div className="text-sm text-ax-danger-900 mb-3">
                            <strong>Permanent blokkerte URL-er:</strong>
                        </div>
                        <div className="max-h-60 overflow-y-auto">
                            <div className="space-y-1"><Box marginBlock="space-12" asChild><List data-aksel-migrated-v8 as="ul" size="small">
                                        {urls.map((url, index) => (
                                            <List.Item key={index} className="bg-white px-3 py-2 rounded border border-ax-danger-300 text-ax-danger-800 font-mono text-sm flex items-center gap-2">
                                                <XMarkOctagonIcon className="w-4 h-4 text-ax-danger-600 flex-shrink-0" />
                                                {url}
                                            </List.Item>
                                        ))}
                                    </List></Box></div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default GlobalDenyListField;
