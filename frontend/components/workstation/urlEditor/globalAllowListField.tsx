import {ChevronDownIcon, ChevronUpIcon, TasklistIcon, CheckmarkCircleIcon } from "@navikt/aksel-icons";
import { Radio, RadioGroup, Stack, Tag, List, Box } from "@navikt/ds-react";
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
                <div className="w-8 h-8 bg-ax-success-200 rounded-full flex items-center justify-center">
                    <CheckmarkCircleIcon className="w-5 h-5 text-ax-success-700" />
                </div>
                <div>
                    <h3 className="font-semibold text-ax-neutral-1000">Sentralt administrerte åpninger</h3>
                    <p className="text-sm text-ax-neutral-700 mt-1">
                        Standard URL-åpninger som er forhåndsgodkjent for alle brukere
                    </p>
                </div>
                <div className="ml-auto">
                    <Tag variant={optIn ? "success" : "neutral"} size="small">
                        {optIn ? "Aktivert" : "Deaktivert"}
                    </Tag>
                </div>
            </div>
            <div className="bg-ax-success-100 border border-ax-success-300 rounded-lg p-4 mb-4">
                <p className="text-ax-success-900 text-sm">
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
                <Stack gap="space-16" direction={{ xs: "column", sm: "row" }} wrap={false}>
                    <Radio value={true} className="flex-1">
                        <div className="flex flex-col">
                            <span className="font-medium">Behold åpninger</span>
                            <span className="text-sm text-ax-neutral-700">Anbefalt for de fleste brukere</span>
                        </div>
                    </Radio>
                    <Radio value={false} className="flex-1">
                        <div className="flex flex-col">
                            <span className="font-medium">Ikke behold åpninger</span>
                            <span className="text-sm text-ax-neutral-700">Kun egendefinerte URL-er</span>
                        </div>
                    </Radio>
                </Stack>
            </RadioGroup>
            <div className="border-t pt-4">
                <button
                    type="button"
                    onClick={() => setShowList(!showList)}
                    className="flex items-center gap-2 text-ax-accent-700 hover:text-ax-accent-900 text-sm font-medium transition-colors"
                >
                    <TasklistIcon className="w-4 h-4" />
                    <span>{showList ? "Skjul" : "Vis"} URL-listen ({urls.length} URL-er)</span>
                    {showList ? <ChevronUpIcon className="w-4 h-4" /> : <ChevronDownIcon className="w-4 h-4" />}
                </button>

                {showList && (
                    <div className="mt-4 border border-ax-neutral-300 bg-ax-neutral-100 rounded-lg p-4">
                        <div className="text-sm text-ax-neutral-800 mb-3">
                            <strong>Globalt tillatte URL-er:</strong>
                        </div>
                        <div className="max-h-60 overflow-y-auto">
                            <div className="space-y-1"><Box marginBlock="space-12" asChild><List data-aksel-migrated-v8 as="ul" size="small">
                                        {urls.map((url, index) => (
                                            <List.Item key={index} className="bg-white px-3 py-2 rounded border text-ax-neutral-800 font-mono text-sm">
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

}

export default GlobalAllowListField
