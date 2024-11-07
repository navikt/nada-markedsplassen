package service

import "github.com/navikt/nada-backend/pkg/errs"

const (
	CodeGoogleCloud        = errs.Code("GOOGLE_CLOUD")
	CodeMetabase           = errs.Code("METABASE")
	CodeSlackNotification  = errs.Code("SLACK_NOTIFICATION")
	CodeTransactionalQueue = errs.Code("TRANSACTIONAL_QUEUE")
	CodeDatabase           = errs.Code("DATABASE")
	CodeInternalDecoding   = errs.Code("DECODING")
	CodeInternalEncoding   = errs.Code("ENCODING")
	CodeExternalEncoding   = errs.Code("EXTERNAL_ENCODING")
)

const (
	ParamDataset        = errs.Parameter("dataset")
	ParamAccessRequest  = errs.Parameter("accessRequest")
	ParamUser           = errs.Parameter("user")
	ParamOwner          = errs.Parameter("owner")
	ParamPiiTags        = errs.Parameter("piiTags")
	ParamDataProduct    = errs.Parameter("dataProduct")
	ParamSchema         = errs.Parameter("schema")
	ParamInsightProduct = errs.Parameter("insightProduct")
	ParamDatasource     = errs.Parameter("datasource")
	ParamGroupEmail     = errs.Parameter("groupEmail")
	ParamPolly          = errs.Parameter("polly")
	ParamProductArea    = errs.Parameter("productArea")
	ParamStory          = errs.Parameter("story")
	ParamJob            = errs.Parameter("job")
)
