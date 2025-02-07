import {fetchTemplate} from "./request";
import { buildUrl } from "./apiUrl";

const onpremMappingPath = buildUrl('onpremMapping')
const buildGetOnpremMappingURL = () => onpremMappingPath()()

export const getOnpremMapping = async () => {
    const url = buildGetOnpremMappingURL();
    return fetchTemplate(url);
}
