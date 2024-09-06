import { Loader } from "@navikt/ds-react";
import { useRouter } from "next/router";
import { EditInsightProductMetadataFields, EditInsightProductMetadataForm } from "../../components/insightProducts/editInsightProduct";
import { use } from "react";
import { useGetInsightProduct } from "../../lib/rest/insightProducts";

const EditInsightProductPage = () => {
    const router = useRouter();
    const id = router.query.id;
    const insightProductQuery = useGetInsightProduct(id as string);

    if (insightProductQuery.error) {
        return <div>{insightProductQuery.error.message}</div>;
    }

    if (insightProductQuery.loading || !insightProductQuery.insightProduct) {
        return <Loader></Loader>;
    }

    const insightProduct = insightProductQuery.insightProduct;

    const formFields: EditInsightProductMetadataFields = {
        id: id as string,
        name: insightProduct.name,
        description: insightProduct.description,
        keywords: insightProduct.keywords,
        teamkatalogenURL: insightProduct.teamkatalogenURL || "",
        group: insightProduct.group,
        type: insightProduct.type,
        link: insightProduct.link
    };

    return (
        <div>
            <EditInsightProductMetadataForm {...formFields} />
        </div>
    );
};

export default EditInsightProductPage;