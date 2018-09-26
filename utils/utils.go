package utils

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// WriteToFile creates a file and writes content to it
func WriteToFile(filename, content string) {
	f, err := os.Create(filename)
	if err != nil {

	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {

	}
}

// Template to replace
func Tmpl(text string, data interface{}, wr io.Writer) error {

	t := template.New("Usage")
	template.Must(t.Parse(text))

	return t.Execute(wr, data)
}

// Commands
func ExeCmd(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("StdoutPipe: %s", err.Error())
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("StderrPipe: %s ", err.Error())
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	bytesErr, err := ioutil.ReadAll(stderr)
	if err != nil {
		return bytesErr, fmt.Errorf("ReadAll stderr: %s", err.Error())
	}

	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("ReadAll stdout: %s", err.Error())
	}
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("%s:err:%s", bytesErr, err.Error())
	}
	return bytes, nil
}

func CheckEnv(appname string) (packpath string, err error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" && strings.Compare(runtime.Version(), "go1.8") >= 0 {
		gopath = DefaultGOPATH()
	}

	currpath, _ := os.Getwd()
	currpath = filepath.Join(currpath, appname)
	gopathStr := filepath.SplitList(gopath)

	for _, gpath := range gopathStr {
		gsrcpath := filepath.Join(gpath, "src")
		if strings.HasPrefix(strings.ToLower(currpath), strings.ToLower(gsrcpath)) {
			packpath = strings.Replace(currpath[len(gsrcpath)+1:], string(filepath.Separator), "/", -1)
			return
		}
	}

	return packpath, errors.New("You current workdir is not inside $GOPATH/src.")
}

func DefaultGOPATH() string {
	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	} else if runtime.GOOS == "plan9" {
		env = "home"
	}
	if home := os.Getenv(env); home != "" {
		return filepath.Join(home, "go")
	}
	return ""
}


func StrFirstToUpper(str string) string {
	temp := strings.Split(str, "_")
	var upperStr string
	for y := 0; y < len(temp); y++ {
		vv := []rune(temp[y])
		for i := 0; i < len(vv); i++ {
			if i == 0 {
				vv[i] -= 32
				upperStr += string(vv[i]) // + string(vv[i+1])
			} else {
				upperStr += string(vv[i])
			}
		}
	}
	return upperStr
}

// Console output as color
type brush func(string) string

func newBrush(color string) brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + "m" + text + reset
	}
}

const (
	Emergency = iota
	Alert
	Critical
	Error
	Warning
	Notice
	Info
	Debug
)

var colors = []brush{
	newBrush("1;37"), // Emergency          white
	newBrush("1;36"), // Alert              cyan
	newBrush("1;35"), // Critical           magenta
	newBrush("1;31"), // Error              red
	newBrush("1;33"), // Warning            yellow
	newBrush("1;32"), // Notice             green
	newBrush("1;34"), // Informational      blue
	newBrush("1;44"), // Debug              Background blue
}

func Output(msg string, level int){
	fmt.Println(colors[level](msg))
}