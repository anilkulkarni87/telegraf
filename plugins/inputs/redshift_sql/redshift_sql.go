package redshift_sql

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"strings"

	// register in driver.
	_ "github.com/lib/pq"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Postgresql struct {
	Address          string
	Outputaddress    string
	Databases        []string
	OrderedColumns   []string
	AllColumns       []string
	AdditionalTags   []string
	sanitizedAddress string
	Query            []struct {
		Sqlquery    string
		Version     int
		Withdbname  bool
		Tagvalue    string
		Measurement string
	}
	Debug bool
}

type query []struct {
	Sqlquery    string
	Version     int
	Withdbname  bool
	Tagvalue    string
	Measurement string
}

var ignoredColumns = map[string]bool{"stats_reset": true}

var sampleConfig = `
## specify address via a url matching:
  address = "host=localhost port=... user=postgres password=... dbname=... sslmode=disable"
  # outputaddress = "db01"
  ## Example :
  ## The mandatory "measurement" value can be used to override the default
  ## output measurement name ("postgresql").
  #
  ## Structure :
  ## [[inputs.postgresql_extensible.query]]
  ##   sqlquery string
  ##   version string
  ##   withdbname boolean
  ##   tagvalue string (comma separated)
  ##   measurement string
  [[inputs.postgresql_extensible.query]]
    sqlquery="SELECT * FROM pg_stat_database"
    version=901
    withdbname=false
    tagvalue=""
    measurement=""
  [[inputs.postgresql_extensible.query]]
    sqlquery="SELECT * FROM pg_stat_bgwriter"
    version=901
    withdbname=false
    tagvalue="postgresql.stats"
`

func (p *Postgresql) SampleConfig() string {
	return sampleConfig
}

func (p *Postgresql) Description() string {
	return "Read metrics from one or many postgresql servers"
}

func (p *Postgresql) IgnoredColumns() map[string]bool {
	return ignoredColumns
}

var localhost = "host=localhost sslmode=disable"

func (p *Postgresql) Gather(acc telegraf.Accumulator) error {
	var (
		err         error
		db          *sql.DB
		sql_query   string
		//query_addon string
		//db_version  int
		//query       string
		tag_value   string
		meas_name   string
	)
	
	if p.Address == "" || p.Address == "localhost" {
		p.Address = localhost
	}
	
	if db, err = sql.Open("postgres", p.Address); err != nil {
		return err
	}
	defer db.Close()
	

for i := range p.Query {
	sql_query = p.Query[i].Sqlquery
	tag_value = p.Query[i].Tagvalue
	if p.Query[i].Measurement != "" {
			meas_name = p.Query[i].Measurement
		} else {
			meas_name = "postgresql"
		}
		
			rows, err := db.Query(sql_query)
	if err != nil {
		return err
	}

	defer rows.Close()

	// grab the column information from the result
	p.OrderedColumns, err = rows.Columns()
	if err != nil {
		return err
	} else {
		for _, v := range p.OrderedColumns {
			p.AllColumns = append(p.AllColumns, v)
		}
	}
	p.AdditionalTags = nil
	if tag_value != "" {
		tag_list := strings.Split(tag_value, ",")
		for t := range tag_list {
			p.AdditionalTags = append(p.AdditionalTags, tag_list[t])
		}
	}

	for rows.Next() {
		err = p.accRow(meas_name, rows, acc)
		if err != nil {
			return err
		}
	}
}
return nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}


func (p *Postgresql) accRow(meas_name string, row scanner, acc telegraf.Accumulator) error {
	var columnVars []interface{}
	var dbname bytes.Buffer

	// this is where we'll store the column name with its *interface{}
	columnMap := make(map[string]*interface{})

	for _, column := range p.OrderedColumns {
		columnMap[column] = new(interface{})
	}

	// populate the array of interface{} with the pointers in the right order
	for i := 0; i < len(columnMap); i++ {
		columnVars = append(columnVars, columnMap[p.OrderedColumns[i]])
	}

	// deconstruct array of variables and send to Scan
	err := row.Scan(columnVars...)

	if err != nil {
		return err
	}
	if columnMap["datname"] != nil {
		// extract the database name from the column map
		dbname.WriteString((*columnMap["datname"]).(string))
	} else {
		dbname.WriteString("postgres")
	}

	//var tagAddress string
	//tagAddress, err = p.SanitizedAddress()
	//if err != nil {
	//	return err
	//}

	// Process the additional tags

	tags := map[string]string{}
	//tags["server"] = tagAddress
	tags["db"] = dbname.String()
	fields := make(map[string]interface{})
COLUMN:
	for col, val := range columnMap {
		log.Printf("D! postgresql_extensible: column: %s = %T: %s\n", col, *val, *val)
		_, ignore := ignoredColumns[col]
		if ignore || *val == nil {
			continue
		}

		for _, tag := range p.AdditionalTags {
			if col != tag {
				continue
			}
			switch v := (*val).(type) {
			case string:
				tags[col] = v
			case []byte:
				tags[col] = string(v)
			case int64, int32, int:
				tags[col] = fmt.Sprintf("%d", v)
			default:
				log.Println("failed to add additional tag", col)
			}
			continue COLUMN
		}
		if v, ok := (*val).([]byte); ok {
			fields[col] = string(v)
		} else {
			fields[col] = *val
		}
	}
	acc.AddFields(meas_name, fields, tags)
	return nil
}
func init() {
	inputs.Add("redshift_sql", func() telegraf.Input {
		return &Postgresql{}
	})
}