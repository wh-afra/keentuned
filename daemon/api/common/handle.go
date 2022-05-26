package common

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	// LimitBytes
	LimitBytes = 1024 * 1024 * 5
)

var compiler = "\"([^\"]+)\""

func registerRouter() {
	http.HandleFunc("/benchmark_result", handler)
	http.HandleFunc("/apply_result", handler)
	http.HandleFunc("/sensitize_result", handler)
	http.HandleFunc("/status", status)
	http.HandleFunc("/cmd", command)
	http.HandleFunc("/write", write)
}

func write(w http.ResponseWriter, r *http.Request) {
	var result = new(string)
	if strings.ToUpper(r.Method) != "POST" {
		*result = fmt.Sprintf("request method '%v' is not supported", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(*result))
		return
	}

	var err error
	defer func() {
		w.WriteHeader(http.StatusOK)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("{\"suc\": false, \"msg\": \"%v\"}", err.Error())))
			log.Errorf("", "write operation: %v", err)
			return
		}

		w.Write([]byte(fmt.Sprintf("{\"suc\": true, \"msg\": \"%s\"}", *result)))
		log.Infof("", "write operation: %v", *result)
	}()

	bytes, err := ioutil.ReadAll(&io.LimitedReader{R: r.Body, N: LimitBytes})
	if err != nil {
		return
	}

	var req struct {
		Name string `json:"name"`
		Info string `json:"info"`
	}

	err = json.Unmarshal(bytes, &req)
	if err != nil {
		err = fmt.Errorf("parse request info failed: %v", err)
		return
	}

	fullName := getFullPath(req.Name)
		
	parts := strings.Split(fullName, "/")
	if !file.IsPathExist(strings.Join(parts[:len(parts)-1], "/")) {
		os.MkdirAll(strings.Join(parts[:len(parts)-1], "/"), os.ModePerm)
	}

	err = ioutil.WriteFile(fullName, []byte(req.Info), 0755)
	if err != nil {
		return
	}

	*result = fmt.Sprintf("write file '%v' successfully.", req.Name)
	return
}

func getFullPath(name string) string {
	var fullName string
	if strings.HasPrefix(name, "/") {
		return name
	}

	if strings.Contains(name, "profile/") {
		fullName = fmt.Sprintf("%v/%v", config.KeenTune.DumpHome, name)
		return fullName
	}

	fullName = fmt.Sprintf("%v/profile/%v", config.KeenTune.DumpHome, name)
	return fullName
}

func handler(w http.ResponseWriter, r *http.Request) {
	// check request method
	var msg string
	if strings.ToUpper(r.Method) != "POST" {
		msg = fmt.Sprintf("request method [%v] is not found", r.Method)
		log.Error("", msg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		return
	}

	bytes, err := ioutil.ReadAll(&io.LimitedReader{R: r.Body, N: LimitBytes})
	defer report(r.URL.Path, bytes, err)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"suc": true, "msg": ""}`))
	return
}

func report(url string, value []byte, err error) {
	if err != nil {
		msg := fmt.Sprintf("read request info err:%v", err)
		log.Error("", "report value to chan err:%v", msg)
	}

	if strings.Contains(url, "benchmark_result") {
		var benchResult struct {
			BenchID int `json:"bench_id"`
		}
		err := json.Unmarshal(value, &benchResult)
		if err != nil {
			fmt.Printf("unmarshal bench id err: %v", err)
			return
		}

		if config.IsInnerBenchRequests[benchResult.BenchID] && benchResult.BenchID > 0 {
			config.BenchmarkResultChan[benchResult.BenchID] <- value
		}

		return
	}

	if strings.Contains(url, "apply_result") {
		var applyResult struct {
			ID int `json:"target_id"`
		}
		err := json.Unmarshal(value, &applyResult)
		if err != nil {
			fmt.Printf("unmarshal apply target id err: %v", err)
			return
		}

		if config.IsInnerApplyRequests[applyResult.ID] && applyResult.ID > 0 {
			config.ApplyResultChan[applyResult.ID] <- value
		}

		return
	}

	if strings.Contains(url, "sensitize_result") && config.IsInnerSensitizeRequests[1] {
		config.SensitizeResultChan <- value
		return
	}
}

func status(w http.ResponseWriter, r *http.Request) {
	// check request method
	var msg string
	if strings.ToUpper(r.Method) != "GET" {
		msg = fmt.Sprintf("request method [%v] is not found", r.Method)
		log.Error("", msg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "alive"}`))
	return
}

func command(w http.ResponseWriter, r *http.Request) {
	var result = new(string)
	var err error
	defer func() {
		w.WriteHeader(http.StatusOK)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("{\"suc\": false, \"msg\": \"%v\"}", err.Error())))
			return
		}

		w.Write([]byte(fmt.Sprintf("{\"suc\": true, \"msg\": \"%s\"}", *result)))
		return
	}()

	if strings.ToUpper(r.Method) != "POST" {
		err = fmt.Errorf("request method \"%v\" is not supported", r.Method)
		return
	}

	var cmd string
	cmd, err = getCmd(r.Body)
	if err != nil {
		return
	}

	err = execCmd(cmd, result)
	if err != nil {
		return
	}
}

func execCmd(inputCmd string, result *string) error {
	cmd := exec.Command("/bin/bash", "-c", inputCmd)
	// Create get command output pipeline
	stderr, _ := cmd.StderrPipe()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("can not obtain stdout pipe for command:%s\n", err)
	}

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("command start err: %v", err)
	}

	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("ReadAll Stdout: %v", err.Error())
	}

	if len(bytes) == 0 {
		bytes, _ = ioutil.ReadAll(stderr)
	}

	if err = cmd.Wait(); err != nil {
		parts := strings.Split(string(bytes), "failed, msg: ")
		if len(parts) > 0 {
			return fmt.Errorf("%s", getMsg(parts[len(parts)-1], inputCmd))
		}

		return fmt.Errorf("%s", getMsg(string(bytes), inputCmd))
	}

	if strings.Contains(string(bytes), "Y(yes)/N(no)") {
		msg := strings.Split(string(bytes), "Y(yes)/N(no)")
		if len(msg) != 2 {
			return fmt.Errorf("get result %v", string(bytes))
		}

		*result = getMsg(msg[1], inputCmd)
		return nil
	}

	*result = getMsg(string(bytes), inputCmd)

	return nil
}

func getMsg(origin, cmd string) string {
	if strings.Contains(cmd, "-h") || strings.Contains(cmd, "jobs") {
		return origin
	}

	pureMSg := strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(
				origin, "\x1b[1;40;32m", ""),
			"\x1b[0m", ""),
		"\x1b[1;40;31m", "")

	changeLinefeed := strings.ReplaceAll(pureMSg, "\n", "\\n")
	changeTab := strings.ReplaceAll(changeLinefeed, "\t", " ")
	return strings.ReplaceAll(strings.TrimSuffix(changeTab, "\\n"), "\"", "'")
}

func getCmd(body io.ReadCloser) (string, error) {
	bytes, err := ioutil.ReadAll(&io.LimitedReader{R: body, N: LimitBytes})
	if err != nil {
		return "", err
	}

	var reqInfo struct {
		Cmd string `json:"cmd"`
	}

	err = json.Unmarshal(bytes, &reqInfo)
	if err != nil {
		return "", err
	}

	if strings.Contains(reqInfo.Cmd, "delete") {
		return "echo y|" + reqInfo.Cmd, nil
	}

	if strings.Contains(reqInfo.Cmd, "param tune") {
		return handleTuneCmd(reqInfo.Cmd)
	}

	return reqInfo.Cmd, nil
}

func handleTuneCmd(originCmd string) (string, error) {
	if !strings.Contains(originCmd, "--config") {
		return originCmd, nil
	}

	matched, err := parseConfigFlag(originCmd)
	if err != nil {
		return matched, err
	}

	retCmd := strings.ReplaceAll(originCmd, matched, config.TuneTempConf)

	return retCmd, nil
}

func parseConfigFlag(originCmd string) (string, error) {
	configPart := strings.Split(originCmd, "--config")
	if len(configPart) < 2 {
		return "", fmt.Errorf("split --config length less than 2")
	}

	re, err := regexp.Compile(compiler)
	if err != nil {
		return "", err
	}

	matched := ""
	parts := re.FindAllString(configPart[1], -1)
	switch len(parts) {
	case 0:
		return "", fmt.Errorf("find all is empty in '%v'", originCmd)
	case 1:
		if len(strings.Trim(parts[0], " ")) == 0 {
			return "", fmt.Errorf("parse config part0 is empty")
		}

		matched = parts[0]
	default:
		if len(parts[0]) > 0 {
			matched = parts[0]
			break
		}
		matched = parts[1]
	}

	if matched == "" {
		return "", fmt.Errorf("config info not found in '%v'", originCmd)
	}

	err = ioutil.WriteFile(config.TuneTempConf, []byte(strings.Trim(matched, "\"")), 0666)
	if err != nil {
		return "", err
	}

	return matched, nil
}

func parseFlag(originCmd, flagName string, short ...string) (string, error) {
	var flagParts []string
	if strings.Contains(originCmd, flagName) {
		flagParts = strings.Split(originCmd, flagName)
	}

	if len(flagParts) == 0 && len(short) > 0 {
		if strings.Contains(originCmd, short[0]) {
			flagParts = strings.Split(originCmd, short[0])
		}
	}

	if len(flagParts) < 2 {
		return "", fmt.Errorf("%v is null", flagName)
	}

	values := strings.Split(flagParts[1], " ")
	if len(values) == 0 {
		return "", fmt.Errorf("--config is null")
	}

	var flagValue string
	flagValue = strings.Trim(values[0], " ")
	if flagValue != "" {
		return flagValue, nil
	}

	if len(values) < 2 {
		return "", fmt.Errorf("%v value is null", flagName)
	}

	flagValue = strings.Trim(values[1], " ")
	if flagValue == "" {
		return "", fmt.Errorf("%v value is empty", flagName)
	}

	return flagValue, nil
}

