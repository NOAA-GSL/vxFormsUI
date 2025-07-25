package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
	"gopkg.in/yaml.v3"
	"github.com/couchbase/gocb/v2"
)

type FormTemplate struct {
	TemplateName   string
	Fields         map[string]interface{}
	SelectFields   map[string][]string
	SelectMode     string
	DisabledFields map[string]bool
}

type Credentials struct {
	CBHost       string   `yaml:"cb_host"`
	CBUser       string   `yaml:"cb_user"`
	CBPassword   string   `yaml:"cb_password"`
	CBBucket     string   `yaml:"cb_bucket"`
	CBScope      string   `yaml:"cb_scope"`
	CBCollection string   `yaml:"cb_collection"`
	Targets      []string `yaml:"targets"`
}

var (
	once          sync.Once
	myCredentials Credentials
)

var (
	jobSpecIDs          []string
	dataSourceIds       []string
	processSpecIds      []string
	ingestDocumentIds   []string
	subsets             []string
	regions             []string
	subDocTypes         []string
	subTypes            []string
	dataSourceSubTypes  []string
	dataSourceStatuses  []string
	statuses            []string
	dataSourceTypes     []string
	processSpecStatuses []string
	ttlTier             []string
	ttlTierSeconds      []string
)

func GetCBCredentials() Credentials {
	once.Do(func() {
		credentialsPath := os.Getenv("CREDENTIALS_FILE")
		if credentialsPath == "" {
			log.Fatal("CREDENTIALS_FILE environment variable not set - should contain the path to the credentials.yaml file")
		}
		if _, err := os.Stat(credentialsPath); err == nil {
			yamlFile, err := os.ReadFile(credentialsPath)
			if err != nil {
				log.Fatalf("GetCBCredentials: yamlFile.Get err   #%v ", err)
			}
			err = yaml.Unmarshal(yamlFile, &myCredentials)
			if err != nil {
				log.Fatalf("GetCBCredentials: Unmarshal: %v", err)
			}
		} else {
			log.Fatalf("Credentials file %v not found", credentialsPath)
		}
	})
	return myCredentials
}

func GetConnection(credentials Credentials) *gocb.Cluster {
	host := credentials.CBHost
	if !strings.Contains(host, "couchbase") {
		host = "couchbases://" + host
	}
	username := credentials.CBUser
	password := credentials.CBPassword
	bucketName := credentials.CBBucket
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
	return cluster
}

func UpsertFormData(id string, data map[string]interface{}) error {
	cluster := GetConnection(GetCBCredentials())
	bucket := cluster.Bucket(GetCBCredentials().CBBucket)
	collection := bucket.Collection("RUNTIME") // Always put this kind of metadata into the RUNTIME collection

	// Upsert the native map, not a JSON string
	_, err := collection.Upsert(id, data, &gocb.UpsertOptions{})
	if err != nil {
		return fmt.Errorf("failed to upsert data: %w", err)
	}
	return nil
}


func GetFormTemplates() ([]FormTemplate, error) {
	cluster := GetConnection(GetCBCredentials())
	query := "SELECT * FROM vxdata._default.COMMON WHERE meta().id like '%TEMPLATE'"
	result, err := cluster.Query(query, &gocb.QueryOptions{})
	if err != nil {
		return nil, err
	}
	var templates []FormTemplate
	for result.Next() {
		var t FormTemplate
		var row map[string]interface{}
		if err := result.Row(&row); err != nil {
			continue
		}
		common, ok := row["COMMON"].(map[string]interface{})
		if !ok {
			continue
		}
		t.TemplateName, _ = common["templateName"].(string)
		fields := make(map[string]interface{}, 0)
		disabledFields := make(map[string]bool, 0)
		selectFields := make(map[string][]string, 0)
		template := common["template"].(map[string]interface{})
		var selectMode string = "multiple"
		for key := range template {
			disabledFields[key] = false
			if _, ok := template[key].(string); ok {
				vStr := template[key].(string)
				if strings.HasPrefix(vStr, "&") {
					selectMode = handleNamedFunction(vStr, selectMode, fields, key)
				} else {
					selectMode = handleFieldStr(vStr, fields, key)
				}
			} else {
				if strings.HasPrefix(key, "@") {
					// If the key is indicating a json, we handle it differently
					// store it as a JSON value
					jsonValue, err := json.MarshalIndent(template[key], "", "  ")
					if err != nil {
						log.Printf("Error marshalling template field %s: %v", key, err)
						fields[key] = fmt.Sprintf("Error: %v", err)
					} else {
						fields[key] = string(jsonValue)
					}
				} else {
					if strings.Contains(key, "Epoch") {
						template[key] = time.Now().Unix()
					}
					fields[key] = template[key]
				}
			}
			var val interface{}
			val = fields[key]
			// test if val starts with a # - if so treat it as a constant
			if strVal, ok := val.(string); ok && strings.HasPrefix(strVal, "#") {
				// If it starts with a #, treat it as a constant
				constantValue := strings.TrimPrefix(strVal, "#")
				fields[key] = constantValue
				disabledFields[key] = true // Mark this field as disabled
				continue
			}
			// Otherwise, handle the value based on its type

			switch val := val.(type) {
			case int:
				fields[key] = val
			case float64:
				if strings.Contains(key, "Epoch") || strings.Contains(key, "duration") {
					// If the key contains "Epoch" or "duration", convert it to an integer
					fields[key] = int(val)
				} else {
					fields[key] = fmt.Sprintf("%f", val)
				}
			case bool:
				fields[key] = val
			case []string:
				selectFields[key] = val
				fields[key] = val
			case []int:
				selectFields[key] = make([]string, len(val))
				for i, v := range val {
					selectFields[key][i] = fmt.Sprintf("%d", v)
				}
				fields[key] = val
			case []float64:
				selectFields[key] = make([]string, len(val))
				if strings.Contains(key, "Epoch") || strings.Contains(key, "duration") {
					// If the key contains "Epoch", convert float64 to int
					intSlice := make([]int, len(val))
					for i, v := range val {
						intSlice[i] = int(v)
					}
					fields[key] = intSlice
				} else {
					fields[key] = val
				}
			case string:
				fields[key] = val
			case []interface{}:
				valSlice := val
				opts := make([]string, 0, len(valSlice))
				for _, v := range valSlice {
					opts = append(opts, fmt.Sprintf("%v", v))
				}
				selectFields[key] = opts
				fields[key] = ""
			default:
				fields[key] = fmt.Sprintf("%v", val)
			}
		}
		t.Fields = fields
		t.SelectMode = selectMode
		t.SelectFields = selectFields
		t.DisabledFields = disabledFields
		templates = append(templates, t)
	}
	return templates, nil
}

func handleFieldStr(vStr string, fields map[string]interface{}, key string) string {
	fields[key] = vStr
	return ""
}

func handleNamedFunction(vStr string, selectMode string, fields map[string]interface{}, key string) string {
	funcName := strings.TrimPrefix(vStr, "&")
	switch funcName {
	case "getSubTypes":
		selectMode = ""
		subTypes, err := GetSubTypes()
		if err != nil {
			log.Printf("Error getting sub types: %v", err)
		} else {
			fields[key] = subTypes
		}
	case "getDataSourceTypes":
		selectMode = ""
		dataSourceTypes, err := GetDataSourceTypes()
		if err != nil {
			log.Printf("Error getting data source types: %v", err)
		} else {
			fields[key] = dataSourceTypes
		}
	case "getSubDocTypes":
		selectMode = ""
		subDocTypes, err := GetSubDocTypes()
		if err != nil {
			log.Printf("Error getting sub document types: %v", err)
		} else {
			fields[key] = subDocTypes
		}
	case "getDataSourceId":
		ids, err := GetDataSourceIds()
		if err != nil {
			log.Printf("Error getting data source IDs: %v", err)
		} else {
			fields[key] = ids
		}
	case "getProcessSpecIds":
		ids, err := GetProcessSpecIds()
		if err != nil {
			log.Printf("Error getting process spec IDs: %v", err)
		} else {
			fields[key] = ids
		}
	case "getIngestDocumentIds":
		ids, err := GetIngestDocumentIds()
		if err != nil {
			log.Printf("Error getting ingest document IDs: %v", err)
		} else {
			fields[key] = ids
		}
	case "getSubsets":
		selectMode = ""
		subsets, err := GetSubsets()
		if err != nil {
			log.Printf("Error getting subsets: %v", err)
			fields[key] = "Error retrieving subsets"
		} else {
			fields[key] = subsets
		}
	case "getRegions":
		selectMode = ""
		regions, err := GetRegions()
		if err != nil {
			log.Printf("Error getting regions: %v", err)
			fields[key] = "Error retrieving regions"
		} else {
			fields[key] = regions
		}
	case "getCTCSubDocTypes":
		selectMode = ""
		subDocTypes, err := GetCTCSubDocTypes()
		if err != nil {
			log.Printf("Error getting CTC sub document types: %v", err)
			fields[key] = "Error retrieving CTC sub document types"
		} else {
			fields[key] = subDocTypes
		}
	case "getTTLTier":
		selectMode = ""
		tiers, err := GetTTLTier()
		if err != nil {
			log.Printf("Error getting TTL tier: %v", err)
			fields[key] = "Error retrieving TTL tier"
		} else {
			fields[key] = tiers
		}
	case "getDataSourceSubTypes":
		selectMode = ""
		subTypes, err := GetDataSourceSubTypes()
		if err != nil {
			log.Printf("Error getting data source sub types: %v", err)
			fields[key] = "Error retrieving data source sub types"
		} else {
			fields[key] = subTypes
		}
	case "getDataSourceStatuses":
		selectMode = ""
		statuses, err := GetDataSourceStatuses()
		if err != nil {
			log.Printf("Error getting data source statuses: %v", err)
			fields[key] = "Error retrieving data source statuses"
		} else {
			fields[key] = statuses
		}
	case "getStatuses":
		selectMode = ""
		statuses, err := GetStatuses()
		if err != nil {
			log.Printf("Error getting statuses: %v", err)
			fields[key] = "Error retrieving statuses"
		} else {
			fields[key] = statuses
		}
	case "getProcessSpecStatuses":
		selectMode = ""
		statuses, err := GetProcessSpecStatuses()
		if err != nil {
			log.Printf("Error getting process spec statuses: %v", err)
			fields[key] = "Error retrieving process spec statuses"
		} else {
			fields[key] = statuses
		}
	default:
		if funcName != "" {
			log.Printf("Unknown function call: %s", funcName)
			fields[key] = fmt.Sprintf("Unknown function: %s", funcName)
		}
	}
	return selectMode
}

func GetJobSpecIDs() ([]string, error) {
	if jobSpecIDs == nil {
		cluster := GetConnection(GetCBCredentials())
		query := "SELECT meta().id FROM vxdata._default.COMMON WHERE type = 'JOB'"
		result, err := cluster.Query(query, &gocb.QueryOptions{})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			var row map[string]interface{}
			if err := result.Row(&row); err == nil {
				if t, ok := row["id"].(string); ok {
					jobSpecIDs = append(jobSpecIDs, fmt.Sprintf("%v", t))
				}
			}
		}
	}
	return jobSpecIDs, nil
}

func GetDataSourceIds() ([]string, error) {
	if dataSourceIds == nil {
		cluster := GetConnection(GetCBCredentials())
		query := "SELECT meta().id FROM vxdata._default.RUNTIME WHERE type = 'DS'"
		result, err := cluster.Query(query, &gocb.QueryOptions{})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			var row map[string]interface{}
			if err := result.Row(&row); err == nil {
				if t, ok := row["id"].(string); ok {
					dataSourceIds = append(dataSourceIds, fmt.Sprintf("%v", t))
				}
			}
		}
	}
	return dataSourceIds, nil
}

func GetProcessSpecIds() ([]string, error) {
	if processSpecIds == nil {
		cluster := GetConnection(GetCBCredentials())
		query := "SELECT meta().id FROM vxdata._default.RUNTIME WHERE type = 'PS'"
		result, err := cluster.Query(query, &gocb.QueryOptions{})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			var row map[string]interface{}
			if err := result.Row(&row); err == nil {
				if t, ok := row["id"].(string); ok {
					processSpecIds = append(processSpecIds, fmt.Sprintf("%v", t))
				}
			}
		}
	}
	return processSpecIds, nil
}

func GetIngestDocumentIds() ([]string, error) {
	if ingestDocumentIds == nil {
		cluster := GetConnection(GetCBCredentials())
		query := "SELECT meta().id FROM vxdata._default.RUNTIME WHERE type = 'IS' and docType = 'ingest'"
		result, err := cluster.Query(query, &gocb.QueryOptions{})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			var row map[string]interface{}
			if err := result.Row(&row); err == nil {
				if t, ok := row["id"].(string); ok {
					ingestDocumentIds = append(ingestDocumentIds, fmt.Sprintf("%v", t))
				}
			}
		}
	}
	return ingestDocumentIds, nil
}

func GetSubsets() ([]string, error) {
	if subsets == nil {
		cluster := GetConnection(GetCBCredentials())
		query := "SELECT DISTINCT subset FROM vxdata._default.COMMON"
		result, err := cluster.Query(query, &gocb.QueryOptions{})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			var row map[string]interface{}
			if err := result.Row(&row); err == nil {
				if t, ok := row["subset"].(string); ok {
					subsets = append(subsets, fmt.Sprintf("%v", t))
				}
			}
		}
	}
	return subsets, nil
}


func GetRegions() ([]string, error) {
	if regions == nil {
		cluster := GetConnection(GetCBCredentials())
		query := "SELECT name FROM vxdata._default.COMMON WHERE type = 'MD' AND docType='region'"
		result, err := cluster.Query(query, &gocb.QueryOptions{})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			var row map[string]interface{}
			if err := result.Row(&row); err == nil {
				if t, ok := row["name"].(string); ok {
					regions = append(regions, fmt.Sprintf("%v", t))
				}
			}
		}
	}
	return regions, nil
}


func GetSubDocTypes() ([]string, error) {
	if subDocTypes == nil {
		cluster := GetConnection(GetCBCredentials())
		query := "SELECT DISTINCT subDocType FROM vxdata._default.COMMON WHERE type = 'MD' and docType = 'ingest'"
		result, err := cluster.Query(query, &gocb.QueryOptions{})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			var row map[string]interface{}
			if err := result.Row(&row); err == nil {
				if t, ok := row["subDocType"].(string); ok {
					if t == "SQL" {
						continue
					}
					subDocTypes = append(subDocTypes, fmt.Sprintf("%v", t))
				}
			}
		}
	}
	return subDocTypes, nil
}

func GetCTCSubDocTypes() ([]string, error) {
	return []string{"CEILING", "VISIBILITY"}, nil
}

func GetSubTypes() ([]string, error) {
	if subTypes == nil {
		cluster := GetConnection(GetCBCredentials())
		query := "SELECT DISTINCT subType FROM vxdata._default.COMMON WHERE type = 'MD' and docType = 'ingest'"
		result, err := cluster.Query(query, &gocb.QueryOptions{})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			var row map[string]interface{}
			if err := result.Row(&row); err == nil {
				if t, ok := row["subType"].(string); ok {
					subTypes = append(subTypes, fmt.Sprintf("%v", t))
				}
			}
		}
	}
	return subTypes, nil
}

func GetDataSourceSubTypes() ([]string, error) {
	if dataSourceSubTypes == nil {
		cluster := GetConnection(GetCBCredentials())
		query := "SELECT subTypes FROM vxdata._default.RUNTIME WHERE meta().id = 'MD:V01:DataSourceSubTypes'"
		result, err := cluster.Query(query, &gocb.QueryOptions{})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			var row map[string]interface{}
			if err := result.Row(&row); err == nil {
				if t, ok := row["subTypes"].([]interface{}); ok {
					for _, v := range t {
						dataSourceSubTypes = append(dataSourceSubTypes, fmt.Sprintf("%v", v))
					}
				}
			}
		}
	}
	return dataSourceSubTypes, nil
}

func GetDataSourceStatuses() ([]string, error) {
	if dataSourceStatuses == nil {
		cluster := GetConnection(GetCBCredentials())
		query := "SELECT statuses FROM vxdata._default.COMMON WHERE meta().id = 'MD:V01:DataSourceStatuses'"
		result, err := cluster.Query(query, &gocb.QueryOptions{})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			var row map[string]interface{}
			if err := result.Row(&row); err == nil {
				if t, ok := row["statuses"].([]interface{}); ok {
					for _, v := range t {
						dataSourceStatuses = append(dataSourceStatuses, fmt.Sprintf("%v", v))
					}
				}
			}
		}
	}
	return dataSourceStatuses, nil
}

func GetStatuses() ([]string, error) {
	if statuses == nil {
		cluster := GetConnection(GetCBCredentials())
		query := "SELECT statuses FROM vxdata._default.COMMON WHERE meta().id = 'MD:V01:Statuses'"
		result, err := cluster.Query(query, &gocb.QueryOptions{})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			var row map[string]interface{}
			if err := result.Row(&row); err == nil {
				if t, ok := row["statuses"].([]interface{}); ok {
					for _, v := range t {
						statuses = append(statuses, fmt.Sprintf("%v", v))
					}
				}
			}
		}
	}
	return statuses, nil
}

func GetDataSourceTypes() ([]string, error) {
	if dataSourceTypes == nil {
		cluster := GetConnection(GetCBCredentials())
		query := "SELECT types FROM vxdata._default.COMMON WHERE meta().id = 'MD:V01:DataSourceTypes'"
		result, err := cluster.Query(query, &gocb.QueryOptions{})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			var row map[string]interface{}
			if err := result.Row(&row); err == nil {
				if t, ok := row["types"].([]interface{}); ok {
					for _, v := range t {
						dataSourceTypes = append(dataSourceTypes, fmt.Sprintf("%v", v))
					}
				}
			}
		}
	}
	return dataSourceTypes, nil
}

func GetProcessSpecStatuses() ([]string, error) {
	if processSpecStatuses == nil {
		cluster := GetConnection(GetCBCredentials())
		query := "SELECT statuses FROM vxdata._default.COMMON WHERE meta().id = 'MD:V01:ProcessSpecStatuses'"
		result, err := cluster.Query(query, &gocb.QueryOptions{})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			var row map[string]interface{}
			if err := result.Row(&row); err == nil {
				if t, ok := row["statuses"].([]interface{}); ok {
					for _, v := range t {
						processSpecStatuses = append(processSpecStatuses, fmt.Sprintf("%v", v))
					}
				}
			}
		}
	}
	return processSpecStatuses, nil
}

func GetTTLTier() ([]string, error) {
	if ttlTier == nil {
		cluster := GetConnection(GetCBCredentials())
		query := "SELECT Tiers FROM vxdata._default.COMMON WHERE meta().id = 'MD:V01:TTLTiers'"
		result, err := cluster.Query(query, &gocb.QueryOptions{})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			var row map[string]interface{}
			if err := result.Row(&row); err == nil {
				if t, ok := row["Tiers"].([]interface{}); ok {
					for _, v := range t {
						ttlTier = append(ttlTier, fmt.Sprintf("%v", v))
					}
				}
			}
		}
	}
	return ttlTier, nil
}

func GetTTLTierSeconds() ([]string, error) {
	if ttlTierSeconds == nil {
		cluster := GetConnection(GetCBCredentials())
		query := "SELECT  TierSeconds FROM vxdata._default.COMMON WHERE meta().id = 'MD:V01:TTLTiers'"
		result, err := cluster.Query(query, &gocb.QueryOptions{})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			var row map[string]interface{}
			if err := result.Row(&row); err == nil {
				if t, ok := row["TierSeconds"].([]interface{}); ok {
					for _, v := range t {
						ttlTierSeconds = append(ttlTierSeconds, fmt.Sprintf("%v", v))
					}
				}
			}
		}
	}
	return ttlTierSeconds, nil
}

// Add this function:
func RetrieveFormData(id string) (map[string]interface{}, error) {
	cluster := GetConnection(GetCBCredentials())
	bucket := cluster.Bucket(GetCBCredentials().CBBucket)
	collection := bucket.Collection("RUNTIME")
	var result map[string]interface{}
	getResult, err := collection.Get(id, &gocb.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve data: %w", err)
	}
	err = getResult.Content(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode content: %w", err)
	}
	return result, nil
}

// Update ListDSIDs to ListIDS and update its references
func ListIDS(docType string) ([]string, error) {
	cluster := GetConnection(GetCBCredentials())
	query := fmt.Sprintf("SELECT meta().id FROM vxdata._default.RUNTIME WHERE type = '%s'", docType)
	result, err := cluster.Query(query, &gocb.QueryOptions{})
	if err != nil {
		return nil, err
	}
	var ids []string
	for result.Next() {
		var row struct {
			ID string `json:"id"`
		}
		if err := result.Row(&row); err != nil {
			continue
		}
		ids = append(ids, row.ID)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no %s IDs found", docType)
	}
	return ids, nil
}

type TopNavData struct {
	FlagLogo       string
	GovLogo        string
	HttpsLogo      string
	TransparentGif string
	ProductLink    string
	ProductText    string
	AgencyLink     string
	AgencyText     string
	BugsLink       string
	BugsText       string
	EmailText      string
	AlertMessage   string
}

