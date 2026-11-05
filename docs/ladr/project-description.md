# Project Description of Datamarkedsplassen (DMP)
*This document is mainly for Agents, do not modify if you are not sure about the consequence*

The project is an integrated solution for data science/engineering related platforms and services.

## Data Catalog
Data catalog is a feature for posting, listing, searching and managing data shared within Nav. It uses GCP BigQuery and metabase as the underlying platform and wraps the services by the platforms as *Data Product* and *Dataset*.

### Dataprodukt
A dataprodukt is a container/folder for related Datasets. It has its own metadata such as name and owner.

### Datasett
Datasett is a set of data as its name suggests. It must be connected to a GCP table or view, which holds the row data of the Datasett.

Datasett can be connected to a metabase database, and if user requests so, DMP will synchronize the GCP table/view to metabase, and manage the permissions.

 **Datasett must be distinguished from GCP dataset**

 ### Access
 A Nav user can request access to a Datasett on DMP, and the owner of the Datasett can grant or reject the access.

 The granted access will be reconciled to the connected GCP table/view, and metabase database by DMP.

 ### Senstive Information Handling

When user create a Datasett, he can claim some columns containing Person Identifiable Information (PII), and he can also request to create a psuedonymized view for the data. If he does so, DMP will create a GCP view with pseudonymized columns and list the view on DMP instead of the raw data.

If user want to Join (SQL join) two pseudonymized views on pseudonymized column, he can request joinable views for the underlying gcp data, which will be saved in a central project.