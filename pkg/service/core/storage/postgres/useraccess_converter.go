package postgres

import (
	"strings"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/database/gensql"
	"github.com/navikt/nada-backend/pkg/service"
)

type UserAccessesConverter []gensql.GetUserAccessesRow

type dataproductBuilder struct {
	DataproductID   uuid.UUID
	DataproductName string
	Datasets        map[uuid.UUID]*service.UserAccessDatasets
}

func (rows UserAccessesConverter) To() (service.UserAccesses, error) {
	grantedMap := make(map[uuid.UUID]*dataproductBuilder)
	serviceAccountMap := make(map[uuid.UUID]*dataproductBuilder)

	for _, row := range rows {
		targetMap := grantedMap
		if strings.HasPrefix(row.AccessSubject, "serviceAccount:") {
			targetMap = serviceAccountMap
		}

		dp, exists := targetMap[row.DataproductID]
		if !exists {
			dp = &dataproductBuilder{
				DataproductID:   row.DataproductID,
				DataproductName: row.DataproductName,
				Datasets:        make(map[uuid.UUID]*service.UserAccessDatasets),
			}
			targetMap[row.DataproductID] = dp
		}
		ds, exists := dp.Datasets[row.DatasetID]
		if !exists {
			ds = &service.UserAccessDatasets{
				DatasetID:   row.DatasetID,
				DatasetName: row.DatasetName,
				Accesses:    make([]service.Access, 0),
			}
			dp.Datasets[row.DatasetID] = ds
		}
		access := rowToAccess(row)
		ds.Accesses = append(ds.Accesses, access)
	}

	return service.UserAccesses{
		Granted:               buildDataproducts(grantedMap),
		ServiceAccountGranted: buildDataproducts(serviceAccountMap),
	}, nil

}

func buildDataproducts(dpMap map[uuid.UUID]*dataproductBuilder) []service.UserAccessDataproduct {
	dataproducts := make([]service.UserAccessDataproduct, 0, len(dpMap))
	for _, dp := range dpMap {
		datasets := make([]service.UserAccessDatasets, 0, len(dp.Datasets))
		for _, ds := range dp.Datasets {
			datasets = append(datasets, *ds)
		}
		dataproducts = append(dataproducts, service.UserAccessDataproduct{
			DataproductID:   dp.DataproductID,
			DataproductName: dp.DataproductName,
			Datasets:        datasets,
		})
	}
	return dataproducts
}

func rowToAccess(row gensql.GetUserAccessesRow) service.Access {
	return service.Access{
		ID:        row.AccessID,
		Subject:   row.AccessSubject,
		Granter:   row.AccessGranter,
		Owner:     row.AccessOwner,
		Expires:   nullTimeToPtr(row.AccessExpires),
		Created:   row.AccessCreated,
		Revoked:   nullTimeToPtr(row.AccessRevoked),
		DatasetID: row.AccessDatasetID,
		Platform:  row.AccessPlatform,
		AccessRequest: &service.AccessRequest{
			ID:          row.AccessRequestID.UUID,
			DatasetID:   row.AccessID,
			Subject:     row.AccessRequestOwner.String,
			SubjectType: strings.Split(row.AccessRequestSubject.String, ":")[0],
			Created:     row.AccessRequestCreated.Time,
			Expires:     nullTimeToPtr(row.AccessRequestExpires),
			Closed:      nullTimeToPtr(row.AccessRequestClosed),
			Granter:     nullStringToPtr(row.AccessRequestGranter),
			Owner:       row.AccessRequestOwner.String,
			Reason:      nullStringToPtr(row.AccessRequestReason),
			Status:      service.AccessRequestStatus(row.AccessRequestStatus.AccessRequestStatusType),
			Polly: &service.Polly{
				ID: row.PollyID.UUID,
				QueryPolly: service.QueryPolly{
					ExternalID: row.PollyExternalID.String,
					Name:       row.PollyName.String,
					URL:        row.PollyUrl.String,
				},
			},
		},
	}
}

