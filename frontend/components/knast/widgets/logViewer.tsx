import { CopyButton, Loader, Table } from "@navikt/ds-react";
import { useUpdateUrlAllowList, useWorkstationLogs } from "../queries";
import { formatDistanceToNow } from "date-fns";
import React from "react";
import { ColorFailed, ColorInfoText } from "../designTokens";
import { ExpendableTextDisplay } from "./urlItem";
import { WorkstationConnectivityWorkflow, WorkstationLogs } from "../../../lib/rest/generatedDto";
import { UseQueryResult } from "@tanstack/react-query";
import { HttpError } from "../../../lib/rest/request";

export interface Log {
    timestamp: string;
    type: "URL blokkert" | "Tilkoblingsfeil";
    message: string;
}

const translateTime = (timeEng: string) => {
    return timeEng.replace("seconds", "sekunder")
        .replace("second", "sekund")
        .replace("minutes", "minutter")
        .replace("minute", "minutt")
        .replace("hours", "timer")
        .replace("hour", "time")
        .replace("days", "dager")
        .replace("day", "dag")
        .replace("weeks", "uker")
        .replace("week", "uke")
        .replace("months", "måneder")
        .replace("month", "måned")
        .replace("years", "år")
        .replace("year", "år")
        .replace("ago", "siden")
        .replace("less than a", "< 1")
        .replace("about ", "ca.");
}

export interface LogViewerProps {
    isLoading: boolean;
    logs: Log[];
}

export const LogViewer = ({ logs, isLoading }: LogViewerProps) => {
    return (
        <div className="w-180 h-140 border-blue-100 border rounded p-4">
            <div className="bg-gray-100">
                <div className="grid grid-cols-[20%_20%_60%] border-b border-gray-300 p-2">
                    <div>
                        Tidspunkt
                    </div>
                    <div>
                        Type
                    </div>
                    <div>
                        Melding
                    </div>
                </div>

            </div>
            {isLoading && <Loader className="mt-6" size="small" title="Laster.." />}
            {logs.length > 0 ? <div className="overflow-y-auto h-110">
                {logs.map((entry: any, i: number) => <div className="grid grid-cols-[20%_20%_60%] border-b border-gray-300 p-2">
                    <div className="text-sm">
                        {isNaN(new Date(entry.timestamp).getTime())
                            ? 'Ugyldig dato'
                            : translateTime(formatDistanceToNow(new Date(entry.timestamp), { addSuffix: true }))
                        }
                    </div>
                    <div>
                        <div style={{
                            color: ColorFailed
                        }}>{entry.type}</div>
                    </div>
                    <div className="flex flex-row gap-1">
                        <ExpendableTextDisplay text={entry.message} lengthLimitHead={20} lengthLimitTail={15} />
                    </div>
                </div>)}
            </div>
                : !isLoading
                    ? <div className="h-full flex justify-center text-sm mt-2">Ingen logger funnet den siste timen</div>
                    : null}
        </div >
    );
}