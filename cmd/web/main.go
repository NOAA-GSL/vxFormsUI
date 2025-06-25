package main

import (
	"encoding/json"
	"os"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/couchbase/gocb/v2"
	"sync"
	"log"
	"gopkg.in/yaml.v3"
	"strings"
	"time"
)

type Credentials struct {
	CBHost       string   `yaml:"cb_host"`
	CBUser       string   `yaml:"cb_user"`
	CBPassword   string   `yaml:"cb_password"`
	CBBucket     string   `yaml:"cb_bucket"`
	CBScope      string   `yaml:"cb_scope"`
	CBCollection string   `yaml:"cb_collection"`
}

var statements map[string]string = map[string]string {
	"job_spec_ids": "SELECT meta().id FROM vxdata._default.COMMON WHERE type = \"JOB\"",
	"ingest_document_ids": "SELECT meta().id FROM vxdata._default.METAR WHERE type=\"MD\" AND docType=\"ingest\"",
}

var (
	myCredentials Credentials
	once          sync.Once
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


// getCBCredentials loads and returns a map of Couchbase credentials from a YAML file specified
// by the CREDENTIALS_FILE environment variable. The credentials are loaded only once using sync.Once.
// If the environment variable is not set, the file does not exist, or there is an error reading or
// unmarshalling the file, the function logs a fatal error and terminates the program.
// Returns the loaded credentials as a Credentials struct.
func getCBCredentials() Credentials {

	once.Do(func() {
			credentialsPath := os.Getenv("CREDENTIALS_FILE")
			if credentialsPath == "" {
					log.Fatal("CREDENTIALS_FILE environment variable not set - should contain the path to the credentials.yaml file")
			}
			if _, err := os.Stat(credentialsPath); err == nil {
					yamlFile, err := os.ReadFile(credentialsPath)
					if err != nil {
							log.Fatalf("getCBCredentials: yamlFile.Get err   #%v ", err)
					}
					err = yaml.Unmarshal(yamlFile, &myCredentials)
					if err != nil {
							log.Fatalf("getCBCredentials: Unmarshal: %v", err)
					}
			} else {
					log.Fatalf("Credentials file %v not found", credentialsPath)
			}
	})
	return myCredentials
}


// getConnection establishes a connection to a Couchbase cluster and retrieves a specific collection.
// It uses the provided credentials to authenticate and configure the connection.
//
// Parameters:
//   - credentials (Credentials): A struct containing the necessary connection details, including:
//   - CBHost: The hostname or IP address of the Couchbase server.
//   - Cb_user: The username for authentication.
//   - CBPassword: The password for authentication.
//
// Notes:
//   - The function applies the "wan-development" profile to optimize latency for remote connections.
//   - Potential errors include:
//   - Invalid credentials leading to authentication failure.
//   - Network issues causing connection timeouts or unreachable hosts.
//   - Invalid Couchbase host URL or misconfigured bucket, scope, or collection names.
//   - The function logs a fatal error and exits if any step in the connection process fails.
//
// Returns:
//   - *gocb.Cluster: A pointer to the connected Couchbase cluster instance.
//   - *gocb.Collection: A pointer to the specified collection within the cluster.
func getConnection(credentials Credentials) (*gocb.Cluster, *gocb.Collection) {
	host := credentials.CBHost
	if !strings.Contains(host, "couchbase") {
			host = "couchbases://" + host
	}
	username := credentials.CBUser
	password := credentials.CBPassword
	bucketName := credentials.CBBucket
	scopeName := credentials.CBScope
	collectionName := credentials.CBCollection

	options := gocb.ClusterOptions{
			Authenticator: gocb.PasswordAuthenticator{
					Username: username,
					Password: password,
			},
	}

	// Sets a pre-configured profile called "wan-development" to help avoid latency issues
	// when accessing Capella from a different Wide Area Network
	// or Availability Zone (e.g. your laptop).
	err := options.ApplyProfile(gocb.ClusterConfigProfileWanDevelopment)
	if err != nil {
			log.Fatalf("getConnection: Failed to apply WAN development profile 'wan-development': %v. Please check your Couchbase configuration.", err)
	}
	// Initialize the Connection
	cluster, err := gocb.Connect(host, options)
	if err != nil {
			log.Fatalf("getConnection: Failed to connect to Couchbase at host '%s': %v", host, err)
	}
	bucket := cluster.Bucket(bucketName)
	err = bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
			log.Fatalf("getConnection: Bucket initialization failed: %v", err)
	}
	collection := cluster.Bucket(bucketName).Scope(scopeName).Collection(collectionName)
	return cluster, collection
}


func formHandler(c *gin.Context) {
	formType := c.Param("type")
	var data map[string]interface{}
	var selectOptions map[string][]string


	selectOptions = make(map[string][]string)

	if formType == "JobSetSpecification" {
		var err error
		selectOptions["job_spec_ids"], err = loadOptions("job_spec_ids")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load job_spec_ids: %v", err)
			return
		}
	}
	if formType == "JobSpecification" || formType == "ProcessSpecification" {
		var err error
		selectOptions["ingest_document_ids"], err = loadOptions("ingest_document_ids")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load ingest_document_ids: %v", err)
			return
		}
	}
	if formType == "IngestDocumentSpecification" {
		var err error
		selectOptions["ingest_document_ids"], err = loadOptions("ingest_document_ids")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load ingest_document_ids: %v", err)
			return
		}
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

// Couchbase upsert handler
func commitHandler(c *gin.Context) {
	var err error
    jsonData := c.PostForm("json")
    if jsonData == "" {
        c.String(400, "No JSON data provided")
        return
    }
	// Get Couchbase credentials and establish connection
	_, collection := getConnection(getCBCredentials())
    // Unmarshal to get an ID field (assumes all JSON has an "id" field)
    var doc map[string]interface{}
    if err := json.Unmarshal([]byte(jsonData), &doc); err != nil {
        c.String(400, "Invalid JSON: %v", err)
        return
    }
    id, ok := doc["id"].(string)
    if !ok || id == "" {
        c.String(400, "JSON must contain an 'id' field")
        return
    }

    // Upsert the document
    _, err = collection.Upsert(id, doc, nil)
    if err != nil {
        c.String(500, "Failed to upsert document: %v", err)
        return
    }

    c.HTML(200, "result.tmpl", gin.H{
        "Result": "Document committed successfully with ID: " + id,
    })
}


// loadOptions retrieves IDs from Couchbase using a N1QL query
func loadOptions(statement_id string) ([]string, error) {

	credentials := getCBCredentials()
	cluster, _ := getConnection(credentials)

	query := statements[statement_id]
    rows, err := cluster.Query(query, &gocb.QueryOptions{})
    if err != nil {
        return nil, err
    }

    var ids []string
    for rows.Next() {
        var row struct {
            ID string `json:"id"`
        }
        if err := rows.Row(&row); err != nil {
            return nil, err
        }
        ids = append(ids, row.ID)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return ids, nil
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