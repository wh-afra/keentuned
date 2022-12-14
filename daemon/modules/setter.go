package modules

import (
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"os"
	"strings"
	"sync"
)

type Setter struct {
	IdMap    map[int]int // key: total group Idx; value: real setter groupIdx
	Group    []bool
	ConfFile []string
}

type ResultProfileSet struct {
	Info    string
	Success bool
}

// Tune : tuning main process
func (tuner *Tuner) Set() error {
	var err error
	tuner.logName = log.ProfSet
	defer func() {
		if err != nil {
			tuner.rollback()
		}
	}()

	if err = tuner.initProfiles(); err != nil {
		log.Errorf(log.ProfSet, "init profiles %v", err)
		return fmt.Errorf("init profiles %v", err)
	}

	configFileALL, err := tuner.checkProfilePath()
	if err != nil {
		log.Errorf(log.ProfSet, "check file err %v", err)
		return fmt.Errorf("check file err %v", err)
	}

	requestInfoAll, err := tuner.getConfigParamInfo(configFileALL)
	if err != nil {
		log.Errorf(log.ProfSet, "Get config: %v", err)
		return fmt.Errorf("Get config: %v", err)
	}

	if err = tuner.prepareBeforeSet(); err != nil {
		log.Errorf(log.ProfSet, "prepare for setting %v", err)
		return fmt.Errorf("prepare for setting %v", err)
	}

	sucInfos, failedInfo, err := tuner.setConfiguration(requestInfoAll)
	if err != nil {
		var details string
		for _, detail := range failedInfo {
			details += fmt.Sprintln(detail)
		}

		log.Errorf(log.ProfSet, "Set failed: %v", details)
		return fmt.Errorf(details)
	}

	err = tuner.updateActive()
	if err != nil {
		return err
	}

	for groupIndex, v := range tuner.Setter.Group {
		successInfoArray, ok := sucInfos[groupIndex+1]
		if v && ok {
			for _, successInfo := range successInfoArray {
				prefix := fmt.Sprintf("target %d apply result: ", groupIndex+1)
				log.Infof(log.ProfSet, "%v Set %v successfully: %v ", utils.ColorString("green", "[OK]"), tuner.Setter.ConfFile[groupIndex], strings.TrimPrefix(successInfo, prefix))
			}
		}
	}

	return nil
}

func (tuner *Tuner) updateActive() error {
	activeFile := config.GetProfileWorkPath("active.conf")
	//?????????????????????
	var fileSet = fmt.Sprintln("name,group_info")
	var activeInfo = make(map[string][]string)
	for groupIndex, settable := range tuner.Setter.Group {
		if settable {
			fileName := file.GetPlainName(tuner.Setter.ConfFile[groupIndex])
			activeInfo[fileName] = append(activeInfo[fileName], fmt.Sprintf("group%v", groupIndex+1))
		}
	}

	for name, info := range activeInfo {
		fileSet += fmt.Sprintf("%s,%s\n", name, strings.Join(info, " "))
	}

	if err := UpdateActiveFile(activeFile, []byte(fileSet)); err != nil {
		log.Errorf(log.ProfSet, "Update active file err:%v", err)
		return fmt.Errorf("update active file err %v", err)
	}

	return nil
}

func (tuner *Tuner) checkProfilePath() (map[int]string, error) {
	filePathAll := make(map[int]string) //key???groupNo???value???.conf
	for groupIndex, v := range tuner.Setter.Group {
		if v {
			filePath := config.GetProfilePath(tuner.Setter.ConfFile[groupIndex])
			if filePath != "" {
				filePathAll[groupIndex] = filePath
			} else {
				return nil, fmt.Errorf("find the configuration file [%v] neither in[%v] nor in [%v]", tuner.Setter.ConfFile[groupIndex], fmt.Sprintf("%s/profile", config.KeenTune.Home), fmt.Sprintf("%s/profile", config.KeenTune.DumpHome))
			}
		}
	}

	return filePathAll, nil
}

func (tuner *Tuner) prepareBeforeSet() error {
	// step1. rollback the target machine
	err := tuner.rollback()
	if err != nil {
		return fmt.Errorf("rollback failed:\n%v", tuner.rollbackFailure)
	}

	// step2. clear the active file
	fileName := config.GetProfileWorkPath("active.conf")
	if err = UpdateActiveFile(fileName, []byte{}); err != nil {
		return fmt.Errorf("update active file failed, err:%v", err)
	}

	// step3. backup the target machine
	err = tuner.backup()
	if err != nil {
		return fmt.Errorf("backup failed:\n%v", tuner.backupFailure)
	}
	return nil
}

func (tuner *Tuner) getConfigParamInfo(configFileALL map[int]string) (map[int][]map[string]interface{}, error) {

	retRequestAll := map[int][]map[string]interface{}{}
	for groupIndex, configFile := range configFileALL {

		resultMap, err := file.ConvertConfFileToJson(configFile)
		if err != nil {
			return nil, fmt.Errorf("convert file '%v' %v", configFile, err)
		}

		var mergedParam = make([]config.DBLMap, config.PRILevel)
		config.ReadProfileParams(resultMap, mergedParam)

		tuner.updateMergeParam(groupIndex, resultMap)

		retRequest := make([]map[string]interface{}, config.PRILevel)
		for index, paramMap := range mergedParam {
			if paramMap == nil {
				continue
			}
			if retRequest[index] == nil {
				retRequest[index] = make(map[string]interface{})
			}
			retRequest[index]["data"] = paramMap
			retRequest[index]["resp_ip"] = config.RealLocalIP
			retRequest[index]["resp_port"] = config.KeenTune.Port
		}
		retRequestAll[groupIndex] = retRequest

	}
	return retRequestAll, nil
}

func (tuner *Tuner) setConfiguration(requestAll map[int][]map[string]interface{}) (map[int][]string, map[int]string, error) {
	var applyResult = make(map[int]map[string]ResultProfileSet)

	//groupIndex???target-group-x   x= groupIndex + 1
	for groupIndex, requestAllPriority := range requestAll {
		for _, request := range requestAllPriority {
			if request == nil {
				continue
			}
			wg := sync.WaitGroup{}
			for _, target := range tuner.Group {
				if target.GroupNo == groupIndex+1 {
					for _, ip := range target.IPs {
						index := config.KeenTune.IPMap[ip]
						wg.Add(1)
						go tuner.set(request, &wg, applyResult, index, ip, target.Port)
					}
				}
			}
			wg.Wait()
		}
	}
	return tuner.analysisApplyResults(applyResult)
}

func (tuner *Tuner) analysisApplyResults(applyResultAll map[int]map[string]ResultProfileSet) (map[int][]string, map[int]string, error) {
	var failedInfo map[int]string
	var successInfo map[int][]string
	var failFlag = false

	failedInfo = make(map[int]string)
	successInfo = make(map[int][]string)

	for applyResultIndex := range applyResultAll {
		for result := range applyResultAll[applyResultIndex] {
			if !applyResultAll[applyResultIndex][result].Success {
				failedInfo[applyResultIndex] += applyResultAll[applyResultIndex][result].Info
				failFlag = true
				continue
			}
			successInfo[applyResultIndex] = append(successInfo[applyResultIndex], applyResultAll[applyResultIndex][result].Info)
		}
		failedInfo[applyResultIndex] = strings.TrimSuffix(failedInfo[applyResultIndex], ";")
	}
	if len(successInfo) == 0 {
		return nil, failedInfo, fmt.Errorf("all failed, details:%v", failedInfo)
	}
	if failFlag {
		return successInfo, failedInfo, fmt.Errorf("partial failed")
	}
	return successInfo, nil, nil
}

func (tuner *Tuner) set(request map[string]interface{}, wg *sync.WaitGroup, applyResultAll map[int]map[string]ResultProfileSet, index int, ip string, port string) {
	config.IsInnerApplyRequests[index] = true
	defer func() {
		wg.Done()
		config.IsInnerApplyRequests[index] = false
	}()

	var applyResult = make(map[string]ResultProfileSet)
	if requestPriority, ok := request["data"]; ok {
		for priorityDomain := range requestPriority.(map[string]map[string]interface{}) {
			uri := fmt.Sprintf("%s:%s/configure", ip, port)
			resp, err := http.RemoteCall("POST", uri, utils.ConcurrentSecurityMap(request, []string{"target_id", "readonly"}, []interface{}{index, false}))
			if err != nil {
				applyResult[priorityDomain] = ResultProfileSet{
					Info:    fmt.Sprintf("target %v apply remote call: [%v] %v;", index, priorityDomain, err),
					Success: false,
				}
				return
			}

			setResult, _, err := GetApplyResult(resp, index)
			if err != nil {
				applyResult[priorityDomain] = ResultProfileSet{
					Info:    fmt.Sprintf("target %v set '%v' %v;", index, priorityDomain, err),
					Success: false,
				}
			} else {
				applyResult[priorityDomain] = ResultProfileSet{
					Info:    fmt.Sprintf("target %v apply result: [%v] %v", index, priorityDomain, setResult),
					Success: true,
				}
			}

			resultSave, ok := applyResultAll[index]
			if !ok {
				resultSave = make(map[string]ResultProfileSet)
				applyResultAll[index] = resultSave
			}
			resultSave[priorityDomain] = applyResult[priorityDomain]
		}
	}
}

func UpdateActiveFile(fileName string, info []byte) error {
	if err := ioutil.WriteFile(fileName, info, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func (tuner *Tuner) updateMergeParam(index int, resultMap map[string]map[string]interface{}) {
	gpIdx := tuner.Setter.IdMap[index]
	if gpIdx < 1 || len(tuner.Group)+1 <= gpIdx {
		return
	}

	var retMap = make(map[string]interface{})
	for name, value := range resultMap {
		retMap[name] = value
	}

	tuner.Group[gpIdx-1].MergedParam = retMap
}

