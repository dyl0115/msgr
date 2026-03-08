package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "로그인 시 자동 실행 등록 (작업 스케줄러)",
	Run: func(cmd *cobra.Command, args []string) {
		exePath, err := os.Executable()
		if err != nil {
			fmt.Printf("실행 파일 경로 오류: %v\n", err)
			return
		}
		exePath, _ = filepath.Abs(exePath)

		// 작업 스케줄러 XML 생성
		xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-16"?>
<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <Triggers>
    <LogonTrigger>
      <Enabled>true</Enabled>
    </LogonTrigger>
  </Triggers>
  <Principals>
    <Principal id="Author">
      <LogonType>InteractiveToken</LogonType>
      <RunLevel>LeastPrivilege</RunLevel>
    </Principal>
  </Principals>
  <Settings>
    <MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy>
    <DisallowStartIfOnBatteries>false</DisallowStartIfOnBatteries>
    <StopIfGoingOnBatteries>false</StopIfGoingOnBatteries>
    <ExecutionTimeLimit>PT0S</ExecutionTimeLimit>
    <Enabled>true</Enabled>
  </Settings>
  <Actions Context="Author">
    <Exec>
      <Command>%s</Command>
      <Arguments>listen</Arguments>
    </Exec>
  </Actions>
</Task>`, exePath)

		// 임시 XML 파일 저장
		tmpFile := filepath.Join(os.TempDir(), "msgr_task.xml")
		if err := os.WriteFile(tmpFile, []byte(xml), 0644); err != nil {
			fmt.Printf("임시 파일 생성 실패: %v\n", err)
			return
		}
		defer os.Remove(tmpFile)

		// 작업 스케줄러 등록
		out, err := exec.Command("schtasks", "/Create", "/TN", "msgr_listen", "/XML", tmpFile, "/F").CombinedOutput()
		if err != nil {
			fmt.Printf("등록 실패: %v\n%s\n", err, string(out))
			return
		}

		fmt.Println("✅ 자동 실행 등록 완료! 로그인 시 자동으로 listen이 시작됩니다.")
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "자동 실행 등록 해제",
	Run: func(cmd *cobra.Command, args []string) {
		out, err := exec.Command("schtasks", "/Delete", "/TN", "msgr_listen", "/F").CombinedOutput()
		if err != nil {
			fmt.Printf("해제 실패: %v\n%s\n", err, string(out))
			return
		}
		fmt.Println("✅ 자동 실행 해제 완료!")
	},
}
