package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", nil)
	})

	r.GET("/form/:type", formHandler)
	r.POST("/form/:type", submitHandler)

	r.Run(":8080")
}

func formHandler(c *gin.Context) {
	formType := c.Param("type")
	var data map[string]interface{}
	var selectOptions map[string][]string
	selectOptions = make(map[string][]string)

	if formType == "JobSetSpecification" {
		selectOptions["job_spec_ids"] = loadOptions("static/job_spec_ids.json", "job_spec_ids")
	}
	if formType == "JobSpecification" || formType == "ProcessSpecification" {
		selectOptions["ingest_document_ids"] = loadOptions("static/ingest_document_ids.json", "ingest_document_ids")
	}
	if formType == "IngestDocumentSpecification" {
		selectOptions["ingest_document_ids"] = loadOptions("static/ingest_document_ids.json", "ingest_document_ids")
	}

	data = getFormTemplate(formType)

	c.HTML(http.StatusOK, "form.tmpl", gin.H{
		"FormType": formType,
		"Data": data,
		"SelectOptions": selectOptions,
	})
}

func submitHandler(c *gin.Context) {
	formType := c.Param("type")
	c.Request.ParseForm()
	formData := make(map[string]interface{})
	for k, v := range c.Request.PostForm {
		if len(v) > 1 {
			formData[k] = v
		} else {
			formData[k] = v[0]
		}
	}
	if _, ok := formData["version"]; !ok {
		formData["version"] = "V01"
	}
	jsonData, _ := json.MarshalIndent(formData, "", "  ")
	c.HTML(http.StatusOK, "result.tmpl", gin.H{
		"FormType": formType,
		"JSON": string(jsonData),
	})
}

func loadOptions(path, key string) []string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	var m map[string][]string
	json.Unmarshal(b, &m)
	return m[key]
}

func getFormTemplate(formType string) map[string]interface{} {
	switch formType {
	case "JobSpecification":
		return map[string]interface{}{
			"data_source_id": "DD",
			"ingest_document_ids": "",
			"status": "active",
			"subDoc": "MODEL",
			"subDocType": "HRRR_OPS",
			"subType": "GRIB2",
			"subset": "RAOB",
			"type": "PS",
			"version": "V01",
		}
	case "JobSetSpecification":
		return map[string]interface{}{
			"job_spec_ids": "",
			"status": "active",
			"subDoc": "MODEL",
			"subDocType": "HRRR_OPS",
			"subType": "GRIB2",
			"subset": "RAOB",
			"type": "PS",
			"version": "V01",
		}
	case "IngestDocumentSpecification":
		return map[string]interface{}{
			"builder_type": "NetcdfMetarObsBuilderV01",
			"docType": "ingest",
			"id": "MD:",
			"requires_time_interpolation": true,
			"subDocType": "netcdf",
			"subType": "obs",
			"subset": "METAR",
			"template": map[string]interface{}{},
			"type": "MD",
			"validTimeDelta": 1800,
			"validTimeInterval": 3600,
			"version": "V01",
		}
	case "ProcessSpecification":
		return map[string]interface{}{
			"id": "PS:RAOB:GRIB2:MODEL:HRRR_OPS:1730496755:1814400:V01",
			"data_source_id": "DS:",
			"ingest_document_ids": "",
			"status": "active",
			"subDoc": "MODEL",
			"subDocType": "HRRR_OPS",
			"subType": "GRIB2",
			"subset": "RAOB",
			"type": "PS",
			"version": "V01",
		}
	case "DataSourceSpecification":
		return map[string]interface{}{
			"id": "DS:operational:HRRR_OPS:1730496755:0:1730498583:V01",
			"type": "DS",
			"sub_type": "operational",
			"name": "HRRR_OPS",
			"start_epoch": 1730496755,
			"duration": 0,
			"source_data_uri": "s3://noaa-hrrr-bdp-pds/",
			"file_mask": "hrrr.YYYYMMDD/conus/hrrr.tHHz.wrfsfcfHH.grib2",
			"source_data_type": "grib2",
			"ingest_location": "s3://noaa-hrrr-bdp-pds/",
			"bundle_location": "s3://vx-storage/import_bundles/",
			"process_spec_id": "PS:METAR:GRIB2:MODEL:HRRR_OPS:1730496755:1814400:V0",
			"requestor": "randy pierce",
			"requestor_email": "randy.pierce@noaa.gov",
			"request_time": 1730498583,
			"status": "initial",
			"version": "0.1",
			"dsg_internal_uri": "...",
			"data_management_document_uri": "...",
			"TTLTier": 4,
		}
	default:
		return map[string]interface{}{}
	}
}