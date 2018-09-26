package gorm

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/carltd/gtool/commands"
	"github.com/carltd/gtool/utils"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	j bool
)

var CmdGorm = &commands.Command{
	Usage: "gorm [-json] driverName datasourceName tableName [generatedPath]",
	Use:   "Invert the table to generate struct.",
	Options: `
    -json             true|false,Generate the struct json tag tag
    driverName        Database driver name, supported: mysql mymysql sqlite3 postgres
    datasourceName    Database connection uri,e.g.(root:123456@tcp(127.0.0.1:3306)/test?charset=utf8)
    tableName         Database table name
    generatedPath     Generated path
`,
	Run: Run,
}

var dbStruct = `package {{.pkgName}}

type {{.dbName}} struct {
{{.dbFiles}}}
`

func init() {
	CmdGorm.Flag.BoolVar(&j, "json", false, "Support Json tag true or false.")
	commands.Register(CmdGorm)
}

var db *gorm.DB

type OrmConfig struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	DbName   string `json:"dbname"`
	LogLevel string `json:"loglevel"`
}

func Run(cmd *commands.Command, args []string) int {
	if len(args) < 3 {
		utils.Output("Usage : orm [-json] driverName datasourceName tableName [generatedPath] ", utils.Warning)
		utils.Output(commands.ErrUseError, utils.Error)
		utils.Output("Too many arguments.", utils.Error)
		return 1
	}

	cmd.Flag.Parse(args)
	args = cmd.Flag.Args()
	if j {
		args = args[1:]
	}

	// Datasource parsing.
	sourceName := args[1]
	source1 := strings.Split(sourceName, "@") // root:123	tcp(%s)/test?charset=utf8
	source2 := strings.Split(source1[1], "/") // tcp(%s)	test?charset=utf8
	lenTcp := len(source2[0])
	user := strings.Split(source1[0], ":")
	i1 := strings.IndexAny(source2[1], "?")
	cf := &OrmConfig{
		Driver:   args[0],
		Host:     source2[0][4 : lenTcp-1],
		User:     user[0],
		Password: user[1],
		DbName:   source2[1][:i1],
		LogLevel: "LOG_DEBUG",
	}
	var err error

	dns := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", cf.User, cf.Password, cf.Host, cf.DbName)
	db, err = gorm.Open(cf.Driver, dns)
	if err != nil {
		utils.Output("gorm : Connecting to database failed, err:" + err.Error(), utils.Error)
		return 1
	}

	var gen string
	if len(args) > 3 {
		gen = args[3]
	}

	if err := reverse(cf.DbName, args[2], gen, j); err != nil {
		utils.Output("gorm : Generate struct to failed, err:" + err.Error(), utils.Error)
		return 1
	}

	return 2
}

func reverse(aliasName, tableName, gen string, isJson bool) error {
	var fieldStr string

	eg := db.Table(tableName)

	// Query runs a raw sql and storage records as []map[string][]byte
	var results []map[string][]byte
	r, err := eg.Raw("show columns from " + tableName).Rows()
	if err != nil {
		return err
	}

	fields, _ := r.Columns()
	scanResultContainers := make([]interface{}, len(fields))
	for i := 0; i < len(fields); i++ {
		var scanResultContainer interface{}
		scanResultContainers[i] = &scanResultContainer
	}

	// Next prepares the next result row for reading with the Scan method. It
	// returns true on success, or false if there is no next result row or an error
	// happened while preparing it. Err should be consulted to distinguish between
	// the two cases.
	//
	// Every call to Scan, even the first one, must be preceded by a call to Next.
	for r.Next() {
		result := make(map[string][]byte)

		if err := r.Scan(scanResultContainers...); err != nil {
			return err
		}
		for ii, key := range fields {
			rawValue := reflect.Indirect(reflect.ValueOf(scanResultContainers[ii]))
			//if row is null then ignore
			if rawValue.Interface() == nil {
				result[key] = []byte{}
				continue
			}

			if data, err := value2Bytes(&rawValue); err == nil {
				result[key] = data
			} else {
				return err // !nashtsai! REVIEW, should return err or just error log?
			}
		}
		results = append(results, result)
	}

	// Traverse the records and process them into the gorm model.
	for _, res := range results {

		//Filed
		field := string(res["Field"])
		//Key
		//	PRI:primary key
		//	UNI:unique key
		//	MUL:index
		key := string(res["Key"])
		//Extra
		//	auto_increment
		extra := string(res["Extra"])
		// Type
		fieldType := res["Type"]

		var (
			fieldDataType string
			gormTag       []string
			gormTags      string
			jsonTag       string
		)
		gormTag = append(gormTag, "column:"+field)

		switch key {
		case "PRI": // Primary key
			gormTag = append(gormTag, "primary_key")
		// TODO UNI:unique key
		// TODO MUL:index
		default:

		}
		switch extra {
		case "auto_increment":
			gormTag = append(gormTag, "AUTO_INCREMENT")
		}

		isUnsigned := bytes.HasSuffix(fieldType, []byte("unsigned"))
		switch {
		case bytes.HasPrefix(fieldType, []byte("int")):
			fieldDataType = "int32"
			if isUnsigned {
				fieldDataType = "uint32"
			}
		case bytes.HasPrefix(res["Type"], []byte("smallint")):
			fieldDataType = "int16"
			if isUnsigned {
				fieldDataType = "uint16"
			}
		case bytes.HasPrefix(res["Type"], []byte("tinyint")):
			fieldDataType = "int8"
			if isUnsigned {
				fieldDataType = "byte"
			}
		case bytes.HasPrefix(res["Type"], []byte("bigint")):
			fieldDataType = "int64"
			if isUnsigned {
				fieldDataType = "uint64"
			}
		case bytes.HasPrefix(res["Type"], []byte("mediumint")):
			fieldDataType = "int32"
			if isUnsigned {
				fieldDataType = "uint32"
			}
		case bytes.HasPrefix(res["Type"], []byte("float")):
			fieldDataType = "float64"
		case bytes.HasPrefix(res["Type"], []byte("double")):
			fieldDataType = "float64"
		case bytes.HasPrefix(res["Type"], []byte("decimal")):
			fieldDataType = "string"
		case bytes.HasPrefix(res["Type"], []byte("date")):
			fieldDataType = "string"
		case bytes.HasPrefix(res["Type"], []byte("datetime")):
			fieldDataType = "time.Time"
			//xormType = "DateTime"
		case bytes.HasPrefix(res["Type"], []byte("time")):
			fieldDataType = "string"
		case bytes.HasPrefix(res["Type"], []byte("varchar")):
			fieldDataType = "string"
		case bytes.HasPrefix(res["Type"], []byte("char")):
			fieldDataType = "string"
		case bytes.HasPrefix(res["Type"], []byte("text")):
			fieldDataType = "string"
		}

		if isJson {
			jsonTag = " json:\"" + field + "\""
		}

		gormTags = fmt.Sprintf("gorm:\"%s\"", strings.Join(gormTag, ";"))
		fieldStr += fmt.Sprintf("%s	%s `%s%s`\n", utils.StrFirstToUpper(field), fieldDataType, gormTags, jsonTag)
	}

	// Generate the file path
	// e.g. a/b/c
	var (
		dbName string
		upperK int
	)
	for k, v := range tableName {
		if 0 == k || upperK == k {
			dbName += strings.ToUpper(string(v))
		} else if "_" == string(v) {
			upperK = k + 1
		} else {
			dbName += string(v)
		}

	}

	wr := os.Stdout
	var pkgName string
	if has := strings.HasSuffix(gen, "/"); !has {
		gen = gen + "/"
	}
	fileName := gen + tableName

	if gen != "/" {
		generatePath := strings.Split(gen, "/")
		length := len(generatePath)
		if length > 0 {
			if f, _ := os.Stat(fileName + ".go"); f != nil {
				return fmt.Errorf("%s.go File already exists, please retry after deletion.", fileName)
			}
			//generateFile := generatePath[length-1]
			pkgName = generatePath[length-2]
			wr, _ = os.Create(fileName + ".go")
		}
	}

	data := template.FuncMap{"pkgName": pkgName, "dbName": dbName, "dbFiles": template.HTML(fieldStr)}
	if err := utils.Tmpl(dbStruct, data, wr); err != nil {
		return fmt.Errorf("Generate %s->%s failed.", aliasName, tableName)
	}

	// TODO go fmt file
	if gen != "/" {
		// Go fmt
		_, err := utils.ExeCmd("go", "fmt", fileName+".go")
		if err != nil {
			return fmt.Errorf("Generate go fmt %s.go failed. err:%s", fileName, err.Error())
		}

		utils.Output(fileName + ".go generated successfully.", utils.Info)
	} else {
		utils.Output("Struct generated successfully.", utils.Info)
	}
	return nil
}

func value2Bytes(rawValue *reflect.Value) ([]byte, error) {
	str, err := value2String(rawValue)
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}

func value2String(rawValue *reflect.Value) (str string, err error) {
	var c_TIME_DEFAULT       time.Time

	aa := reflect.TypeOf((*rawValue).Interface())
	vv := reflect.ValueOf((*rawValue).Interface())
	switch aa.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str = strconv.FormatInt(vv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		str = strconv.FormatUint(vv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		str = strconv.FormatFloat(vv.Float(), 'f', -1, 64)
	case reflect.String:
		str = vv.String()
	case reflect.Array, reflect.Slice:
		switch aa.Elem().Kind() {
		case reflect.Uint8:
			data := rawValue.Interface().([]byte)
			str = string(data)
			if str == "\x00" {
				str = "0"
			}
		default:
			err = fmt.Errorf("Unsupported struct type %v", vv.Type().Name())
		}
		// time type
	case reflect.Struct:
		if aa.ConvertibleTo(reflect.TypeOf(c_TIME_DEFAULT)) {
			str = vv.Convert(reflect.TypeOf(c_TIME_DEFAULT)).Interface().(time.Time).Format(time.RFC3339Nano)
		} else {
			err = fmt.Errorf("Unsupported struct type %v", vv.Type().Name())
		}
	case reflect.Bool:
		str = strconv.FormatBool(vv.Bool())
	case reflect.Complex128, reflect.Complex64:
		str = fmt.Sprintf("%v", vv.Complex())
		/* TODO: unsupported types below
		   case reflect.Map:
		   case reflect.Ptr:
		   case reflect.Uintptr:
		   case reflect.UnsafePointer:
		   case reflect.Chan, reflect.Func, reflect.Interface:
		*/
	default:
		err = fmt.Errorf("Unsupported struct type %v", vv.Type().Name())
	}
	return
}
