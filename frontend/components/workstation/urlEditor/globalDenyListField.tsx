import { ExclamationmarkTriangleIcon, ChevronDownIcon, ChevronUpIcon, TasklistIcon, XMarkOctagonIcon } from "@navikt/aksel-icons";
import { List, Tag } from "@navikt/ds-react";
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
                <div className="w-8 h-8 bg-red-100 rounded-full flex items-center justify-center">
                    <XMarkOctagonIcon className="w-5 h-5 text-red-600" />
                </div>
                <div>
                    <h3 className="font-semibold text-gray-900">Globalt blokkerte URL-er</h3>
                    <p className="text-sm text-gray-600 mt-1">
                        URL-er som er permanent blokkert av systemadministrator
                    </p>
                </div>
                <div className="ml-auto">
                    <Tag variant="error" size="small">
                        {urls.length} blokkert{urls.length !== 1 ? 'e' : ''}
                    </Tag>
                </div>
            </div>

            <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-4">
                <div className="flex items-start gap-2">
                    <ExclamationmarkTriangleIcon className="w-4 h-4 text-red-600 mt-0.5 flex-shrink-0" />
                    <p className="text-red-800 text-sm">
                        <strong>Advarsel:</strong> Disse URL-ene er globalt blokkert og kan ikke aksesseres fra Knast,
                        uavhengig av dine personlige innstillinger.
                    </p>
                </div>
            </div>

            <div className="border-t pt-4">
                <button
                    type="button"
                    onClick={() => setShowList(!showList)}
                    className="flex items-center gap-2 text-blue-600 hover:text-blue-800 text-sm font-medium transition-colors"
                >
                    <TasklistIcon className="w-4 h-4" />
                    <span>{showList ? "Skjul" : "Vis"} blokkerte URL-er ({urls.length} URL-er)</span>
                    {showList ? <ChevronUpIcon className="w-4 h-4" /> : <ChevronDownIcon className="w-4 h-4" />}
                </button>

                {showList && (
                    <div className="mt-4 border border-red-200 bg-red-50 rounded-lg p-4">
                        <div className="text-sm text-red-800 mb-3">
                            <strong>Permanent blokkerte URL-er:</strong>
                        </div>
                        <div className="max-h-60 overflow-y-auto">
                            <List as="ul" size="small" className="space-y-1">
                                {urls.map((url, index) => (
                                    <List.Item key={index} className="bg-white px-3 py-2 rounded border border-red-200 text-red-700 font-mono text-sm flex items-center gap-2">
                                        <XMarkOctagonIcon className="w-4 h-4 text-red-500 flex-shrink-0" />
                                        {url}
                                    </List.Item>
                                ))}
                            </List>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default GlobalDenyListField;
