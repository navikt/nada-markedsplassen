import { Loader } from "@navikt/ds-react";
import { useRouter } from "next/router";
import { EditMetabaseResource } from "../../components/metabaseDashboards/EditMetabaseResource";
import { useGetMetabaseDocument } from "../../lib/rest/metabaseDocuments";

const EditMetabaseDocumentPage = () => {
    const router = useRouter();
    const id = router.query.id;
    const metabaseDocumentQuery = useGetMetabaseDocument(id as string);

    if (metabaseDocumentQuery.error) {
        return <div>{metabaseDocumentQuery.error.message}</div>;
    }

    if (metabaseDocumentQuery.isLoading || !metabaseDocumentQuery.data) {
        return <Loader/>;
    }

    const metabaseDocument = metabaseDocumentQuery.data;

    return (
        <div>
            <EditMetabaseResource
                resourceType="document"
                id={id as string}
                name={metabaseDocument.name}
                description={metabaseDocument.description}
                keywords={metabaseDocument.keywords}
                teamkatalogenURL={metabaseDocument.teamkatalogenURL || ""}
                group={metabaseDocument.group}
                link={metabaseDocument.link}
                teamID={metabaseDocument.teamID || ""}
            />
        </div>
    );
};

export default EditMetabaseDocumentPage;
