import { CopyButton, Loader, Table } from "@navikt/ds-react";
import { useUpdateUrlAllowList, useWorkstationLogs } from "../queries";
import { formatDistanceToNow } from "date-fns";
import React from "react";
import { ColorFailed, ColorInfoText } from "../designTokens";
import { UrlTextDisplay } from "./urlItem";
import { LogEntry, WorkstationConnectivityWorkflow, WorkstationLogs } from "../../../lib/rest/generatedDto";
import { UseQueryResult } from "@tanstack/react-query";
import { HttpError } from "../../../lib/rest/request";

interface BlockedURLLogProps {
    entry: LogEntry;
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

interface ConnectivityLogProps {
    entry: any;
}
const ConnectivityLog = ({ entry }: ConnectivityLogProps) => {
    return <div className="grid grid-cols-[20%_80%] border-b border-gray-300 p-2">
        <div className="text-sm">
            {isNaN(new Date(entry.Timestamp).getTime())
                ? 'Ugyldig dato'
                : translateTime(formatDistanceToNow(new Date(entry.Timestamp), { addSuffix: true }))
            }
        </div>
        <div className="flex flex-row gap-2">
            <div style={{
                color: ColorFailed
            }}>Tilkoblingsfeil</div>
            <UrlTextDisplay url={`${entry.error}`} lengthLimitHead={40} lengthLimitTail={30} />
        </div>
    </div>;
}

const BlockedURLLog = ({ entry }: BlockedURLLogProps) => {
    return <div className="grid grid-cols-[20%_80%] border-b border-gray-300 p-2">
        <div className="text-sm">
            {isNaN(new Date(entry.Timestamp).getTime())
                ? 'Ugyldig dato'
                : translateTime(formatDistanceToNow(new Date(entry.Timestamp), { addSuffix: true }))
            }
        </div>
        <div>
            <div className="flex flex-row gap-2 justify-between">
                <div title={`${entry.HTTPRequest?.URL.Host}${entry.HTTPRequest?.URL.Path}`} className="flex flex-row gap-2">
                    <div style={{
                        color: ColorFailed
                    }}>URL blokkert</div>
                    <div className="flex flex-row items-start">
                        <CopyButton
                            title="Kopier URL"
                            color={ColorInfoText}
                            copyText={`${entry.HTTPRequest?.URL.Host}${entry.HTTPRequest?.URL.Path}`}
                            aria-text="Kopier URL"
                            size="xsmall"
                        />
                        <UrlTextDisplay url={`${entry.HTTPRequest?.URL.Host}${entry.HTTPRequest?.URL.Path}`} lengthLimitHead={20} lengthLimitTail={10} />
                    </div>
                </div>
            </div>
        </div>
    </div>;
}

export interface LogViewerProps {
    isLoading: boolean;
    logs: (UseQueryResult<WorkstationLogs, HttpError> | any) [];
}

export const LogViewer = ({ logs, isLoading}: LogViewerProps) => {
    return (
        <div className="w-180 h-140 border-blue-100 border rounded p-4">
            <div className="bg-gray-100">
                <div className="grid grid-cols-[20%_80%] border-b border-gray-300 p-2">
                    <div>
                        Tidspunkt
                    </div>
                    <div>
                        Meldling
                    </div>
                </div>

            </div>
            {isLoading && <Loader className="mt-6" size="small" title="Laster.." />}
            {logs.length > 0 ? <div className="overflow-y-auto h-110">
                {logs.map((url: any, i: number) => url.HTTPRequest?.URL?.Host ? (<div key={i}>
                    <BlockedURLLog key={i + url.HTTPRequest.URL.Host + url.Timestamp} entry={url} />
                </div>)
                    : <div key={i}><ConnectivityLog key={i} entry={url} />
                    </div>)}
            </div>
                : !isLoading
                    ? <div className="h-full flex justify-center text-sm mt-2">Ingen logger funnet den siste timen</div>
                    : null}
        </div >
    );
}