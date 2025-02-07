import {createQueryKeyStore} from '@lukemorales/query-key-factory';
import {useQuery} from 'react-query';
import {
    ClassifiedHosts,
} from "../../lib/rest/generatedDto";
import {HttpError} from "../../lib/rest/request";
import {getOnpremMapping} from "../../lib/rest/onpremmapping";

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
