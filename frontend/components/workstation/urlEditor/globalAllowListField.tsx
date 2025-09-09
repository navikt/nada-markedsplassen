import {ChevronDownIcon, ChevronUpIcon, TasklistIcon, CheckmarkCircleIcon } from "@navikt/aksel-icons";
import { Radio, RadioGroup, Stack, Tag, List } from "@navikt/ds-react";
import { useState } from "react";

interface GlobalAllowListSelectorProps {
    urls: string[];
    optIn: boolean;
    onChange: (value: boolean) => void;
}

const GlobalAllowListField = ({ onChange, urls, optIn }: GlobalAllowListSelectorProps) => {
    const description = "Noen åpninger mot internett har mange nytte av og vi har derfor valgt å åpne disse som standard for alle brukere. Men, du står fritt til å ikke åpne for disse."
    const [showList, setShowList] = useState(false);

    return (
        <div className="bg-white rounded-lg shadow-sm border p-6 mb-6">
            <div className="flex items-center gap-3 mb-4">
                <div className="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center">
                    <CheckmarkCircleIcon className="w-5 h-5 text-green-600" />
                </div>
                <div>
                    <h3 className="font-semibold text-gray-900">Sentralt administrerte åpninger</h3>
                    <p className="text-sm text-gray-600 mt-1">
                        Standard URL-åpninger som er forhåndsgodkjent for alle brukere
                    </p>
                </div>
                <div className="ml-auto">
                    <Tag variant={optIn ? "success" : "neutral"} size="small">
                        {optIn ? "Aktivert" : "Deaktivert"}
                    </Tag>
                </div>
            </div>

            <div className="bg-green-50 border border-green-200 rounded-lg p-4 mb-4">
                <p className="text-green-800 text-sm">
                    {description}
                </p>
            </div>

            <RadioGroup
                legend=""
                defaultValue={optIn}
                value={optIn}
                onChange={onChange}
                className="mb-4"
            >
                <Stack gap="4" direction={{ xs: "column", sm: "row" }} wrap={false}>
                    <Radio value={true} className="flex-1">
                        <div className="flex flex-col">
                            <span className="font-medium">Behold åpninger</span>
                            <span className="text-sm text-gray-600">Anbefalt for de fleste brukere</span>
                        </div>
                    </Radio>
                    <Radio value={false} className="flex-1">
                        <div className="flex flex-col">
                            <span className="font-medium">Ikke behold åpninger</span>
                            <span className="text-sm text-gray-600">Kun egendefinerte URL-er</span>
                        </div>
                    </Radio>
                </Stack>
            </RadioGroup>

            <div className="border-t pt-4">
                <button
                    type="button"
                    onClick={() => setShowList(!showList)}
                    className="flex items-center gap-2 text-blue-600 hover:text-blue-800 text-sm font-medium transition-colors"
                >
                    <TasklistIcon className="w-4 h-4" />
                    <span>{showList ? "Skjul" : "Vis"} URL-listen ({urls.length} URL-er)</span>
                    {showList ? <ChevronUpIcon className="w-4 h-4" /> : <ChevronDownIcon className="w-4 h-4" />}
                </button>

                {showList && (
                    <div className="mt-4 border border-gray-200 bg-gray-50 rounded-lg p-4">
                        <div className="text-sm text-gray-700 mb-3">
                            <strong>Globalt tillatte URL-er:</strong>
                        </div>
                        <div className="max-h-60 overflow-y-auto">
                            <List as="ul" size="small" className="space-y-1">
                                {urls.map((url, index) => (
                                    <List.Item key={index} className="bg-white px-3 py-2 rounded border text-gray-700 font-mono text-sm">
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

}

export default GlobalAllowListField
