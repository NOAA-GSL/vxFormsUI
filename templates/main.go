package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type JobSpecification struct {
	DataSourceID string   `json:"data_source_id"`
	IngestDocIDs []string `json:"ingest_document_ids"`
	Status       string   `json:"status"`
	SubDoc       string   `json:"subDoc"`
	SubDocType   string   `json:"subDocType"`
	SubType      string   `json:"subType"`
	Subset       string   `json:"subset"`
	Type         string   `json:"type"`
	Version      string   `json:"version"`
}

type JobSetSpecification struct {
	JobSpecIDs []string `json:"job_spec_ids"`
	Status     string   `json:"status"`
	SubDoc     string   `json:"subDoc"`
	SubDocType string   `json:"subDocType"`
	SubType    string   `json:"subType"`
	Subset     string   `json:"subset"`
	Type       string   `json:"type"`
	Version    string   `json:"version"`
}

type IngestDocumentSpecification struct {
	BuilderType               string      `json:"builder_type"`
	DocType                   string      `json:"docType"`
	ID                        string      `json:"id"`
	RequiresTimeInterpolation bool        `json:"requires_time_interpolation"`
	SubDocType                string      `json:"subDocType"`
	SubType                   string      `json:"subType"`
	Subset                    string      `json:"subset"`
	Template                  interface{} `json:"template"`
	Type                      string      `json:"type"`
	ValidTimeDelta            int         `json:"validTimeDelta"`
	ValidTimeInterval         int         `json:"validTimeInterval"`
	Version                   string      `json:"version"`
}

type ProcessSpecification struct {
	ID           string   `json:"id"`
	DataSourceID string   `json:"data_source_id"`
	IngestDocIDs []string `json:"ingest_document_ids"`
	Status       string   `json:"status"`
	SubDoc       string   `json:"subDoc"`
	SubDocType   string   `json:"subDocType"`
	SubType      string   `json:"subType"`
	Subset       string   `json:"subset"`
	Type         string   `json:"type"`
	Version      string   `json:"version"`
}

type DataSourceSpecification struct {
	ID                        string `json:"id"`
	Type                      string `json:"type"`
	SubType                   string `json:"sub_type"`
	Name                      string `json:"name"`
	StartEpoch                int64  `json:"start_epoch"`
	Duration                  int    `json:"duration"`
	SourceDataURI             string `json:"source_data_uri"`
	FileMask                  string `json:"file_mask"`
	SourceDataType            string `json:"source_data_type"`
	IngestLocation            string `json:"ingest_location"`
	BundleLocation            string `json:"bundle_location"`
	ProcessSpecID             string `json:"process_spec_id"`
	Requestor                 string `json:"requestor"`
	RequestorEmail            string `json:"requestor_email"`
	RequestTime               int64  `json:"request_time"`
	Status                    string `json:"status"`
	Version                   string `json:"version"`
	DSGInternalURI            string `json:"dsg_internal_uri"`
	DataManagementDocumentURI string `json:"data_management_document_uri"`
	TTLTier                   int    `json:"TTLTier"`
}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "select_form.html", nil)
	})

	r.GET("/form/job", func(c *gin.Context) {
		c.HTML(http.StatusOK, "form_job.html", nil)
	})
	r.GET("/form/jobset", func(c *gin.Context) {
		c.HTML(http.StatusOK, "form_jobset.html", nil)
	})
	r.GET("/form/ingest", func(c *gin.Context) {
		c.HTML(http.StatusOK, "form_ingest.html", nil)
	})
	r.GET("/form/process", func(c *gin.Context) {
		c.HTML(http.StatusOK, "form_process.html", nil)
	})
	r.GET("/form/datasource", func(c *gin.Context) {
		c.HTML(http.StatusOK, "form_datasource.html", nil)
	})

	r.POST("/submit/job", func(c *gin.Context) {
		doc := JobSpecification{
			DataSourceID: c.PostForm("data_source_id"),
			IngestDocIDs: []string{
				c.PostForm("ingest1"),
				c.PostForm("ingest2"),
			},
			Status:     c.PostForm("status"),
			SubDoc:     c.PostForm("subDoc"),
			SubDocType: c.PostForm("subDocType"),
			SubType:    c.PostForm("subType"),
			Subset:     c.PostForm("subset"),
			Type:       c.PostForm("type"),
			Version:    c.PostForm("version"),
		}
		c.JSON(http.StatusOK, doc)
	})

	r.POST("/submit/jobset", func(c *gin.Context) {
		doc := JobSetSpecification{
			JobSpecIDs: []string{
				c.PostForm("job1"),
				c.PostForm("job2"),
			},
			Status:     c.PostForm("status"),
			SubDoc:     c.PostForm("subDoc"),
			SubDocType: c.PostForm("subDocType"),
			SubType:    c.PostForm("subType"),
			Subset:     c.PostForm("subset"),
			Type:       c.PostForm("type"),
			Version:    c.PostForm("version"),
		}
		c.JSON(http.StatusOK, doc)
	})

	r.POST("/submit/ingest", func(c *gin.Context) {
		var template map[string]interface{}
		_ = json.Unmarshal([]byte(c.PostForm("template")), &template)
		doc := IngestDocumentSpecification{
			BuilderType:               c.PostForm("builder_type"),
			DocType:                   c.PostForm("docType"),
			ID:                        c.PostForm("id"),
			RequiresTimeInterpolation: c.PostForm("requires_time_interpolation") == "true",
			SubDocType:                c.PostForm("subDocType"),
			SubType:                   c.PostForm("subType"),
			Subset:                    c.PostForm("subset"),
			Template:                  template,
			Type:                      c.PostForm("type"),
			ValidTimeDelta:            atoi(c.PostForm("validTimeDelta")),
			ValidTimeInterval:         atoi(c.PostForm("validTimeInterval")),
			Version:                   c.PostForm("version"),
		}
		c.JSON(http.StatusOK, doc)
	})

	r.POST("/submit/process", func(c *gin.Context) {
		doc := ProcessSpecification{
			ID:           c.PostForm("id"),
			DataSourceID: c.PostForm("data_source_id"),
			IngestDocIDs: []string{c.PostForm("ingest1"), c.PostForm("ingest2")},
			Status:       c.PostForm("status"),
			SubDoc:       c.PostForm("subDoc"),
			SubDocType:   c.PostForm("subDocType"),
			SubType:      c.PostForm("subType"),
			Subset:       c.PostForm("subset"),
			Type:         c.PostForm("type"),
			Version:      c.PostForm("version"),
		}
		c.JSON(http.StatusOK, doc)
	})

	r.POST("/submit/datasource", func(c *gin.Context) {
		doc := DataSourceSpecification{
			ID:                        c.PostForm("id"),
			Type:                      c.PostForm("type"),
			SubType:                   c.PostForm("sub_type"),
			Name:                      c.PostForm("name"),
			StartEpoch:                atoi64(c.PostForm("start_epoch")),
			Duration:                  atoi(c.PostForm("duration")),
			SourceDataURI:             c.PostForm("source_data_uri"),
			FileMask:                  c.PostForm("file_mask"),
			SourceDataType:            c.PostForm("source_data_type"),
			IngestLocation:            c.PostForm("ingest_location"),
			BundleLocation:            c.PostForm("bundle_location"),
			ProcessSpecID:             c.PostForm("process_spec_id"),
			Requestor:                 c.PostForm("requestor"),
			RequestorEmail:            c.PostForm("requestor_email"),
			RequestTime:               atoi64(c.PostForm("request_time")),
			Status:                    c.PostForm("status"),
			Version:                   c.PostForm("version"),
			DSGInternalURI:            c.PostForm("dsg_internal_uri"),
			DataManagementDocumentURI: c.PostForm("data_management_document_uri"),
			TTLTier:                   atoi(c.PostForm("TTLTier")),
		}
		c.JSON(http.StatusOK, doc)
	})

	r.Run(":8080")
}

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func atoi64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}
