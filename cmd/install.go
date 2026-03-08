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
	Short: "로그인 시 자동 실행 등록 (작업 스케줄러 + 방화벽)",
	Run: func(cmd *cobra.Command, args []string) {
		exePath, err := os.Executable()
		if err != nil {
			fmt.Printf("실행 파일 경로 오류: %v\n", err)
			return
		}
		exePath, _ = filepath.Abs(exePath)

		// 1. 작업 스케줄러 등록
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

		tmpFile := filepath.Join(os.TempDir(), "msgr_task.xml")
		if err := os.WriteFile(tmpFile, []byte(xml), 0644); err != nil {
			fmt.Printf("임시 파일 생성 실패: %v\n", err)
			return
		}
		defer os.Remove(tmpFile)

		out, err := exec.Command("schtasks", "/Create", "/TN", "msgr_listen", "/XML", tmpFile, "/F").CombinedOutput()
		if err != nil {
			fmt.Printf("작업 스케줄러 등록 실패: %v\n%s\n", err, string(out))
			return
		}
		fmt.Println("✅ 작업 스케줄러 등록 완료!")

		// 2. 방화벽 인바운드 규칙 추가
		out, err = exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
			"name=msgr_listen",
			"dir=in",
			"action=allow",
			"protocol=TCP",
			"localport=9999",
		).CombinedOutput()
		if err != nil {
			fmt.Printf("방화벽 규칙 추가 실패: %v\n%s\n", err, string(out))
			return
		}
		fmt.Println("✅ 방화벽 규칙 추가 완료!")

		fmt.Println("\n🎉 설치 완료! 로그인 시 자동으로 listen이 시작됩니다.")
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "자동 실행 등록 해제 (작업 스케줄러 + 방화벽)",
	Run: func(cmd *cobra.Command, args []string) {
		// 1. 작업 스케줄러 해제
		out, err := exec.Command("schtasks", "/Delete", "/TN", "msgr_listen", "/F").CombinedOutput()
		if err != nil {
			fmt.Printf("작업 스케줄러 해제 실패: %v\n%s\n", err, string(out))
		} else {
			fmt.Println("✅ 작업 스케줄러 해제 완료!")
		}

		// 2. 방화벽 규칙 삭제
		out, err = exec.Command("netsh", "advfirewall", "firewall", "delete", "rule",
			"name=msgr_listen",
		).CombinedOutput()
		if err != nil {
			fmt.Printf("방화벽 규칙 삭제 실패: %v\n%s\n", err, string(out))
		} else {
			fmt.Println("✅ 방화벽 규칙 삭제 완료!")
		}

		fmt.Println("\n🗑️ 제거 완료!")
	},
}
