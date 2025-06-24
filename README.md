# AVID Forms

This project is a Go module that uses the [Gin web framework](https://gin-gonic.com/) and [Bootstrap 5](https://getbootstrap.com/) to implement a web-based UI with the following features:

- **Form Selection:** Users can choose which form to use from a main page.
- **Dynamic Form Rendering:** The UI presents the appropriate form for creating the associated JSON document based on user input.
- **Default Version Value:** Any input field named `version` is pre-filled with the default value `"V01"`.
- **Back Navigation:** Each form includes a back button that returns the user to the main page.

The forms are based on the following JSON specifications:

- [JobSpecification](#form-jobspecification)
- [JobSetSpecification](#form-jobsetspecification)
- [IngestDocumentSpecification](#form-ingestdocumentspecification)
- [ProcessSpecification](#form-processspecificiation)
- [DataSourceSpecification](#form-datasourcespecification)

---

The following was used for chatgpt to create the initial module.

Create a GO module
That uses the gin web framework and bootstrap 5 to implement a web based UI
that:

1) Allows the user to select which form to use.
2) Presents the appropriate form for creating the associated json document from the users input to the form.
3) Has a default value of "V01" to any input named "version"
4) Has a back button on each form that returns the user to the main page.

using the following json for each form.

Form JobSpecification:

```json
{
  "data_source_id":"DS:RAOB:HRRR_OPS:operational:1730496755:1814400:V01",
  "ingest_document_ids":[
    "MD:V01:RAOB:PRS:HRRR_OPS:ingest:grib2",
    "MD:V01:RAOB:NTV:HRRR_OPS:ingest:grib2"
  ],
  "status": "active",
  "subDoc": "MODEL",
  "subDocType": "HRRR_OPS",
  "subType": "GRIB2",
  "subset": "RAOB",
  "type": "PS",
  "version": "V01"
}
```

Form JobSetSpecification:

```json
{
  "job_spec_ids":[
    "JOB:V01:RAOB:PRS:HRRR_OPS:ingest:grib2",
    "JOB:V01:RAOB:NTV:HRRR_OPS:ingest:grib2"
  ],
  "status": "active",
  "subDoc": "MODEL",
  "subDocType": "HRRR_OPS",
  "subType": "GRIB2",
  "subset": "RAOB",
  "type": "PS",
  "version": "V01"
}
```

Form IngestDocumentSpecification:

```json
{
  "builder_type": "NetcdfMetarObsBuilderV01",
  "docType": "ingest",
  "id": "MD:V01:METAR:obs:ingest:netcdf",
  "requires_time_interpolation": true,
  "subDocType": "netcdf",
  "subType": "obs",
  "subset": "METAR",
  "template": {
    "correctedTime": "",
    "data": {
      "*stationName": {
        "Ceiling": "&ceiling_transform|*skyCover,*skyLayerBase",
        "DewPoint": "&kelvin_to_fahrenheit|*dewpoint",
        "Reported Time": "&retrieve_from_netcdf|*timeObs",
        "Surface Pressure": "&handle_pressure|*altimeter",
        "Temperature": "&kelvin_to_fahrenheit|*temperature",
        "Visibility": "&handle_visibility|*visibility",
        "WD": "&retrieve_from_netcdf|*windDir",
        "WS": "&meterspersecond_to_milesperhour|*windSpeed",
        "name": "&handle_station|*stationName"
      }
    },
    "dataSourceId": "MADIS",
    "docType": "obs",
    "fcstValidEpoch": "&derive_valid_time_epoch|%Y%m%d_%H%M",
    "fcstValidISO": "&derive_valid_time_iso|%Y%m%d_%H%M",
    "id": "DD:V01:METAR:obs:&derive_valid_time_epoch|%Y%m%d_%H%M",
    "subset": "METAR",
    "type": "DD",
    "units": {
      "Ceiling": "ft",
      "DewPoint": "deg F",
      "RH": "percent",
      "Surface Pressure": "mb",
      "Temperature": "deg F",
      "Visibility": "miles",
      "WD": "degrees",
      "WS": "mph"
    },
    "version": "V01"
  },
  "type": "MD",
  "validTimeDelta": 1800,
  "validTimeInterval": 3600,
  "version": "V01"
}
```

Form ProcessSpecificiation:

```json
{
  "id":"PS:RAOB:GRIB2:MODEL:HRRR_OPS:1730496755:1814400:V01",
  "data_source_id":"DS:RAOB:HRRR_OPS:operational:1730496755:1814400:V01",
  "ingest_document_ids":[
    "MD:V01:RAOB:PRS:HRRR_OPS:ingest:grib2",
    "MD:V01:RAOB:NTV:HRRR_OPS:ingest:grib2"
  ],
  "status": "active",
  "subDoc": "MODEL",
  "subDocType": "HRRR_OPS",
  "subType": "GRIB2",
  "subset": "RAOB",
  "type": "PS",
  "version": "V01"
}
```

Form DataSourceSpecification:

```json
{
  "id": "DS:operational:HRRR_OPS:1730496755:0:1730498583:V01",
  "type": "DS",
  "sub_type": "{operational, retro, backfill, reprocess}",
  "name": "HRRR_OPS",
  "start_epoch": 1730496755,
  "duration": 0,
  "source_data_uri": "s3://noaa-hrrr-bdp-pds/",
  "file_mask": "hrrr.YYYYMMDD/conus/hrrr.tHHz.wrfsfcfHH.grib2",
  "source_data_type": "grib2",
  "ingest_location": "s3://noaa-hrrr-bdp-pds/",
  "bundle_location": "s3://vx-storage/import_bundles/",
"process_spec_id":"PS:METAR:GRIB2:MODEL:HRRR_OPS:1730496755:1814400:V0",
  "requestor": "randy pierce",
  "requestor_email":"randy.pierce@noaa.gov",
  "request_time": 1730498583,
  "status": "{initial, active, inactive}",
  "version": "0.1",
  "dsg_internal_uri": "...",
  "data_management_document_uri": "..."
  "TTLTier": 4
}
```

