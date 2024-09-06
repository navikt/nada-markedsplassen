import * as yup from 'yup'

export const bigQuery = yup.object().shape({
  projectID: yup.string().required(),
  dataset: yup.string().required(),
  table: yup.string().required(),
})

export const updateDataproductValidation = yup.object().shape({
  name: yup.string().required('Du må fylle inn navn'),
  description: yup.string(),
  slug: yup.string(),
  repo: yup.string(),
  teamContact: yup.string().nullable(),
  teamkatalogenURL: yup.string().nullable().required('Du må velge et team i Teamkatalogen'),
  productAreaID: yup.string().nullable(),
})

export const newDataproductValidation = updateDataproductValidation.concat(
  yup.object().shape({
    group: yup.string().required('Velg et eierteam for produktet'),
    bigquery: bigQuery,
  })
)

export const owner = yup.object().shape({
  group: yup.string().required('trenger teamnavn'),
  teamkatalogen: yup.string(),
})

export const storyValidation = yup.object().shape({
  group: yup.string().required('trenger teamnavn'),
  teamkatalogenURL: yup.string().required('Du må velge et team i Teamkatalogen'),
})

export const editStoryValidation = yup.object().shape({
  teamkatalogenURL: yup.string().required('Du må velge et team i Teamkatalogen'),
})
