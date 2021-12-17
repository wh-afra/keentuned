package profile

import (
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	m "keentune/daemon/modules"
	"strings"
)

// List run profile list service
func (s *Service) List(flag string, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ProfList]
		log.ClearCliLog(log.ProfList)
	}()

	proFileList, err := walkProfileAllFiles()
	if err != nil {
		log.Errorf(log.ProfList, "Walk file path failed: %v", err)
		return fmt.Errorf("Walk file path failed: %v", err)
	}

	var fileListInfo string
	activeFileName := m.GetProfileWorkPath("active.conf")
	activeNameBytes, _ := ioutil.ReadFile(activeFileName)

	for _, value := range proFileList {
		if string(activeNameBytes) == value {
			fileListInfo += fmt.Sprintf("%s\t%v\n", utils.ColorString("GREEN", "[active]"), value)
			continue
		}

		if value == "active.conf" {
			continue
		}

		fileListInfo += fmt.Sprintf("[available]\t%v\n", value)
	}

	if len(fileListInfo) == 0 {
		log.Info(log.ProfList, "There is no profile configuration file exists, please execute keentune param dump first.")
		return nil
	}

	log.Infof(log.ProfList, "%v", strings.TrimSuffix(fileListInfo, "\n"))

	return nil
}

func walkProfileAllFiles() ([]string, error) {
	proFileList, err := file.WalkFilePath(m.GetProfileWorkPath(""), "", false)
	if err != nil {
		return proFileList, fmt.Errorf("walk dump folder failed :%v", err)
	}

	homeFileList, err := file.WalkFilePath(m.GetProfileHomePath(""), ".conf", false)
	if err != nil {
		return proFileList, fmt.Errorf("walk home folder failed :%v", err)
	}

	proFileList = append(proFileList, homeFileList...)
	return proFileList, nil
}

