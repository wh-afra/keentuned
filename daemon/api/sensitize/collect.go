package sensitize

import (
	com "keentune/daemon/api/common"
	"keentune/daemon/api/param"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"fmt"
)

type Service struct {
}

// Collect run sensitize collect service
func (s *Service) Collect(flag param.TuneFlag, reply *string) error {
	if com.SystemRun {
		log.Errorf("", "An instance is running. You can wait the process finish or run \"keentune %v stop\" and try a new job again, if you want give up the old job.", com.GetRunningTask())
		return fmt.Errorf("Collect failed, an instance is running. You can wait the process finish or run \"keentune %v stop\" and try a new job again, if you want give up the old job.", com.GetRunningTask())
	}

	go runCollect(flag)
	return nil
}
func runCollect(flag param.TuneFlag) {
	com.SystemRun = true
	com.IsSensitizing = true
	log.SensitizeCollect += ":" + flag.Log
	
	defer func() {
		log.ClearCliLog(log.SensitizeCollect)
		config.ProgramNeedExit <- true
		<-config.ServeFinish
		com.SystemRun = false
		com.IsSensitizing = false
	}()

	go com.HeartbeatCheck()
	log.Infof(log.SensitizeCollect, "Step1. Parameter auto Sensitize start, using algorithm = %v.", config.KeenTune.Sensitize.Algorithm)

	if com.IsDataNameUsed(flag.Name) {
		log.Errorf(log.SensitizeCollect, "Please check the data name specified, already exists")
		return
	}

	if err := param.TuningImpl(flag, "collect"); err != nil {
		return
	}
}
