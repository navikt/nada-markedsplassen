import {createQueryKeyStore} from '@lukemorales/query-key-factory';
import {
    ClassifiedHosts,
} from "../../lib/rest/generatedDto";
import {HttpError} from "../../lib/rest/request";
import {getOnpremMapping} from "../../lib/rest/onpremmapping";
import {useQuery} from "@tanstack/react-query";

export const queries = createQueryKeyStore({
    onpremmapping: {
        all: null,
    }
});

export function useOnpremMapping() {
    return useQuery<ClassifiedHosts, HttpError>({
        ...queries.onpremmapping.all,
        queryFn: getOnpremMapping,
    });
}
