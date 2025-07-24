package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	// Custom function to check if a string contains a substring
	r.SetFuncMap(template.FuncMap{
		"isString": func(v interface{}) bool {
			_, ok := v.(string)
			return ok
		},
		"Contains": func(s interface{}, substr string) bool {
			return strings.Contains(fmt.Sprintf("%v", s), substr)
		},
		"HasPrefix": func(s interface{}, prefix string) bool {
			return strings.HasPrefix(fmt.Sprintf("%v", s), prefix)
		},
		"TrimPrefix": func(s interface{}, prefix string) string {
			return strings.TrimPrefix(fmt.Sprintf("%v", s), prefix)
		},
		"ToJSON": func(v interface{}) string {
			var encodeBuffer bytes.Buffer
			var indentBuffer bytes.Buffer

			encoder := json.NewEncoder(&encodeBuffer)
			encoder.SetEscapeHTML(false)
			_ = encoder.Encode(v)
			_ = json.Indent(&indentBuffer, bytes.TrimRight(encodeBuffer.Bytes(), "\n"), "", "  ")
			return indentBuffer.String()
		},
		"SafeHtml": func(s string) template.HTML {
			return template.HTML(s)
		},
		"IsString": func(v interface{}) bool {
			_, ok := v.(string)
			return ok
		},
	})
	r.Static("/static", "./static")
	r.Static("/img", "./static/img")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		templates, err := GetFormTemplates()
		if err != nil {
			c.String(http.StatusInternalServerError, "Error loading forms")
			return
		}
		data := gin.H{
			"FlagLogo":       "./static/img/us_flag_small.png",
			"GovLogo":        "./static/img/icon-dot-gov.svg",
			"HttpsLogo":      "./static/img/icon-https.svg",
			"TransparentGif": "./static/img/noaa_transparent.gif",
			"ProductLink":    "/",
			"ProductText":    "vxFormsUI",
			"AgencyLink":     "https://gsl.noaa.gov/",
			"AgencyText":     "Global Systems Laboratory",
			"BugsLink":       "https://github.com/NOAA-GSL/vxFormsUI/issues",
			"BugsText":       "Bugs/Issues (GitHub)",
			"EmailText":      "mailto:mats.gsl@noaa.gov?Subject=Feedback from vxFormsUI",
			"forms":          templates,
		}

		c.HTML(http.StatusOK, "index.html", data)
	})

	r.GET("/form/:name", func(c *gin.Context) {
		name := c.Param("name")
		templates, _ := GetFormTemplates()
		var selected FormTemplate
		for _, t := range templates {
			if t.TemplateName == name {
				selected = t
				break
			}
		}
		jobSpecIDs, _ := GetJobSpecIDs()
		c.HTML(http.StatusOK, "form.html", gin.H{"form": selected, "jobSpecIDs": jobSpecIDs})
	})

	r.POST("/commit-json", func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.BindJSON(&data); err != nil {
			c.String(http.StatusBadRequest, "Invalid JSON")
			return
		}
		id, ok := data["id"].(string)
		if !ok || strings.Contains(id, "*") || id == "" {
			c.String(http.StatusBadRequest, "Error: The id field is missing or contains '*'. Cannot commit.")
			return
		}

		// Assume you have a function UpsertFormData(id string, data map[string]interface{}) error
		err := UpsertFormData(id, data)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to upsert data to database")
			return
		}
		c.String(http.StatusOK, fmt.Sprintf("Upserted form data with id: %s", id))

	})

	r.GET("/retrieve-json", func(c *gin.Context) {
		id := c.Query("id")
		if id == "" {
			c.String(http.StatusBadRequest, "Missing id")
			return
		}
		data, err := RetrieveFormData(id)
		if err != nil {
			c.String(http.StatusNotFound, "Not found")
			return
		}
		c.JSON(http.StatusOK, data)
	})

	r.GET("/list-ds-ids", func(c *gin.Context) {
		docType := c.Query("type")
		if docType == "" {
			docType = "DS" // default to DS if not provided
		}
		ids, err := ListIDS(docType)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to list ids")
			return
		}
		c.JSON(http.StatusOK, ids)
	})

	r.Run(":8080")
}
